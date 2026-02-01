# Performance Optimization

> Last Updated: 2025-01-24

Comprehensive guide to profiling and optimizing Linkrift for maximum performance across backend (Go), frontend (Vite/React), and infrastructure layers.

---

## Table of Contents

- [Overview](#overview)
- [Go Backend Optimization](#go-backend-optimization)
  - [Profiling with pprof](#profiling-with-pprof)
  - [Memory Optimization](#memory-optimization)
  - [CPU Optimization](#cpu-optimization)
  - [Concurrency Patterns](#concurrency-patterns)
- [Frontend Optimization](#frontend-optimization)
  - [Vite Build Optimization](#vite-build-optimization)
  - [React Performance](#react-performance)
  - [Bundle Size Optimization](#bundle-size-optimization)
- [Caching Strategies](#caching-strategies)
  - [Redis Caching](#redis-caching)
  - [HTTP Caching](#http-caching)
  - [CDN Configuration](#cdn-configuration)
- [Database Optimization](#database-optimization)
- [Infrastructure Tuning](#infrastructure-tuning)
- [Monitoring & Benchmarking](#monitoring--benchmarking)

---

## Overview

Linkrift is designed for high performance:

**Performance Targets:**
| Metric | Target | Current |
|--------|--------|---------|
| Redirect latency (p50) | < 10ms | 5ms |
| Redirect latency (p99) | < 50ms | 25ms |
| API latency (p50) | < 50ms | 30ms |
| API latency (p99) | < 200ms | 120ms |
| Throughput | > 50k req/s | 65k req/s |
| Memory per instance | < 100MB | 60MB |

---

## Go Backend Optimization

### Profiling with pprof

Go's built-in `pprof` tool provides comprehensive profiling capabilities.

**Enable pprof endpoint:**

```go
// cmd/server/main.go
package main

import (
    "net/http"
    _ "net/http/pprof" // Import for side effects
)

func main() {
    // pprof endpoints available at /debug/pprof/
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()

    // Start main server
    startServer()
}
```

**CPU Profiling:**

```bash
# Collect 30-second CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Interactive commands in pprof
(pprof) top 10          # Top 10 CPU consumers
(pprof) list funcName   # Show source with CPU usage
(pprof) web             # Open flame graph in browser
(pprof) svg > cpu.svg   # Export flame graph
```

**Memory Profiling:**

```bash
# Heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# In pprof
(pprof) top 10 -cum     # Top memory allocators
(pprof) list funcName   # Show allocations in function

# Allocation profiling
go tool pprof http://localhost:6060/debug/pprof/allocs
```

**Block Profiling (Contention):**

```go
// Enable block profiling
runtime.SetBlockProfileRate(1)
```

```bash
go tool pprof http://localhost:6060/debug/pprof/block
```

**Goroutine Analysis:**

```bash
# Current goroutine stack traces
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

**Continuous Profiling with pyroscope:**

```go
import "github.com/pyroscope-io/client/pyroscope"

func main() {
    pyroscope.Start(pyroscope.Config{
        ApplicationName: "linkrift-api",
        ServerAddress:   "http://pyroscope:4040",
        ProfileTypes: []pyroscope.ProfileType{
            pyroscope.ProfileCPU,
            pyroscope.ProfileAllocObjects,
            pyroscope.ProfileAllocSpace,
            pyroscope.ProfileInuseObjects,
            pyroscope.ProfileInuseSpace,
        },
    })
    defer pyroscope.Stop()
}
```

### Memory Optimization

**1. Reduce Allocations:**

```go
// Bad: Allocates new slice on each call
func processLinks(links []Link) []ProcessedLink {
    result := make([]ProcessedLink, 0)
    for _, link := range links {
        result = append(result, processLink(link))
    }
    return result
}

// Good: Pre-allocate with known capacity
func processLinks(links []Link) []ProcessedLink {
    result := make([]ProcessedLink, 0, len(links))
    for _, link := range links {
        result = append(result, processLink(link))
    }
    return result
}

// Better: Reuse slice when possible
var processedPool = sync.Pool{
    New: func() interface{} {
        return make([]ProcessedLink, 0, 100)
    },
}

func processLinks(links []Link) []ProcessedLink {
    result := processedPool.Get().([]ProcessedLink)[:0]
    for _, link := range links {
        result = append(result, processLink(link))
    }
    // Return to pool after use
    return result
}
```

**2. Use sync.Pool for Frequent Allocations:**

```go
// Buffer pool for JSON encoding
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func encodeJSON(v interface{}) ([]byte, error) {
    buf := bufferPool.Get().(*bytes.Buffer)
    buf.Reset()
    defer bufferPool.Put(buf)

    encoder := json.NewEncoder(buf)
    if err := encoder.Encode(v); err != nil {
        return nil, err
    }

    // Make a copy since we're returning the buffer to the pool
    result := make([]byte, buf.Len())
    copy(result, buf.Bytes())
    return result, nil
}
```

**3. String Concatenation:**

```go
// Bad: Creates many intermediate strings
func buildURL(scheme, host, path string) string {
    return scheme + "://" + host + "/" + path
}

// Good: Use strings.Builder
func buildURL(scheme, host, path string) string {
    var b strings.Builder
    b.Grow(len(scheme) + 3 + len(host) + 1 + len(path))
    b.WriteString(scheme)
    b.WriteString("://")
    b.WriteString(host)
    b.WriteString("/")
    b.WriteString(path)
    return b.String()
}
```

**4. Avoid Interface{} When Possible:**

```go
// Bad: Interface{} causes allocations
func processValue(v interface{}) {
    // Type assertion and boxing overhead
}

// Good: Use generics (Go 1.18+)
func processValue[T any](v T) {
    // No allocation for value types
}
```

### CPU Optimization

**1. Optimize Hot Paths:**

```go
// The redirect handler is the hottest path
func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Fast path: extract short code directly
    path := r.URL.Path
    if len(path) > 1 {
        shortCode := path[1:] // Skip leading slash

        // Try cache first (fastest)
        if url, ok := h.cache.Get(shortCode); ok {
            w.Header().Set("Location", url)
            w.WriteHeader(http.StatusMovedPermanently)
            return
        }

        // Database lookup
        link, err := h.repo.FindByShortCode(r.Context(), shortCode)
        if err == nil {
            h.cache.Set(shortCode, link.OriginalURL)
            w.Header().Set("Location", link.OriginalURL)
            w.WriteHeader(http.StatusMovedPermanently)
            return
        }
    }

    http.NotFound(w, r)
}
```

**2. Use Efficient Data Structures:**

```go
// For frequently accessed data, consider specialized structures
import "github.com/cespare/xxhash/v2"

// Fast hash map for cache keys
type FastCache struct {
    shards [256]map[string]string
    locks  [256]sync.RWMutex
}

func (c *FastCache) getShard(key string) int {
    return int(xxhash.Sum64String(key) % 256)
}

func (c *FastCache) Get(key string) (string, bool) {
    shard := c.getShard(key)
    c.locks[shard].RLock()
    val, ok := c.shards[shard][key]
    c.locks[shard].RUnlock()
    return val, ok
}
```

**3. Avoid Reflection:**

```go
// Bad: Uses reflection
func validateStruct(v interface{}) error {
    val := reflect.ValueOf(v)
    // ... reflection operations
}

// Good: Use code generation or explicit validation
func (l *Link) Validate() error {
    if l.ShortCode == "" {
        return ErrMissingShortCode
    }
    if l.OriginalURL == "" {
        return ErrMissingURL
    }
    return nil
}
```

### Concurrency Patterns

**1. Worker Pools:**

```go
type ClickProcessor struct {
    clicks chan Click
    wg     sync.WaitGroup
}

func NewClickProcessor(workers int) *ClickProcessor {
    p := &ClickProcessor{
        clicks: make(chan Click, 10000),
    }

    for i := 0; i < workers; i++ {
        p.wg.Add(1)
        go p.worker()
    }

    return p
}

func (p *ClickProcessor) worker() {
    defer p.wg.Done()
    batch := make([]Click, 0, 100)
    ticker := time.NewTicker(time.Second)

    for {
        select {
        case click, ok := <-p.clicks:
            if !ok {
                if len(batch) > 0 {
                    p.flushBatch(batch)
                }
                return
            }
            batch = append(batch, click)
            if len(batch) >= 100 {
                p.flushBatch(batch)
                batch = batch[:0]
            }
        case <-ticker.C:
            if len(batch) > 0 {
                p.flushBatch(batch)
                batch = batch[:0]
            }
        }
    }
}
```

**2. Connection Pooling:**

```go
// Database connection pool
db, err := sql.Open("postgres", connString)
db.SetMaxOpenConns(100)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(1 * time.Minute)

// HTTP client pool
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 100,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 10 * time.Second,
}
```

---

## Frontend Optimization

### Vite Build Optimization

**vite.config.ts:**

```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { compression } from 'vite-plugin-compression2';
import { visualizer } from 'rollup-plugin-visualizer';

export default defineConfig({
  plugins: [
    react(),
    compression({
      algorithm: 'gzip',
      threshold: 1024,
    }),
    compression({
      algorithm: 'brotliCompress',
      threshold: 1024,
    }),
    visualizer({
      filename: 'stats.html',
      open: true,
    }),
  ],

  build: {
    target: 'es2020',
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
      },
    },
    rollupOptions: {
      output: {
        manualChunks: {
          // Split vendor chunks
          'react-vendor': ['react', 'react-dom'],
          'router': ['react-router-dom'],
          'query': ['@tanstack/react-query'],
          'ui': ['@radix-ui/react-dialog', '@radix-ui/react-dropdown-menu'],
          'charts': ['recharts'],
        },
      },
    },
    chunkSizeWarningLimit: 500,
    reportCompressedSize: true,
  },

  // Optimize dependencies
  optimizeDeps: {
    include: ['react', 'react-dom', 'react-router-dom'],
  },
});
```

### React Performance

**1. Memoization:**

```tsx
// Memoize expensive components
const LinkListItem = React.memo(function LinkListItem({
  link,
  onDelete,
}: LinkListItemProps) {
  return (
    <div className="link-item">
      <span>{link.shortCode}</span>
      <span>{link.clicks} clicks</span>
      <button onClick={() => onDelete(link.id)}>Delete</button>
    </div>
  );
});

// Memoize callbacks
function LinkList({ links }: LinkListProps) {
  const handleDelete = useCallback((id: string) => {
    deleteLink(id);
  }, []);

  return (
    <div>
      {links.map((link) => (
        <LinkListItem
          key={link.id}
          link={link}
          onDelete={handleDelete}
        />
      ))}
    </div>
  );
}
```

**2. Virtualization for Large Lists:**

```tsx
import { useVirtualizer } from '@tanstack/react-virtual';

function VirtualizedLinkList({ links }: { links: Link[] }) {
  const parentRef = useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: links.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 50,
    overscan: 5,
  });

  return (
    <div ref={parentRef} className="h-96 overflow-auto">
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          position: 'relative',
        }}
      >
        {virtualizer.getVirtualItems().map((virtualItem) => (
          <div
            key={virtualItem.key}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              height: `${virtualItem.size}px`,
              transform: `translateY(${virtualItem.start}px)`,
            }}
          >
            <LinkListItem link={links[virtualItem.index]} />
          </div>
        ))}
      </div>
    </div>
  );
}
```

**3. Code Splitting:**

```tsx
// Lazy load routes
const Analytics = lazy(() => import('./pages/Analytics'));
const Settings = lazy(() => import('./pages/Settings'));

function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/analytics" element={<Analytics />} />
        <Route path="/settings" element={<Settings />} />
      </Routes>
    </Suspense>
  );
}
```

### Bundle Size Optimization

**1. Tree Shaking:**

```typescript
// Bad: Imports entire library
import _ from 'lodash';

// Good: Import only what you need
import debounce from 'lodash/debounce';

// Or use lodash-es for better tree shaking
import { debounce } from 'lodash-es';
```

**2. Analyze Bundle:**

```bash
# Generate bundle analysis
npm run build -- --mode production
npx vite-bundle-visualizer

# Check bundle size
npx size-limit
```

**3. size-limit configuration:**

```json
// package.json
{
  "size-limit": [
    {
      "path": "dist/assets/*.js",
      "limit": "200 KB"
    },
    {
      "path": "dist/assets/*.css",
      "limit": "50 KB"
    }
  ]
}
```

---

## Caching Strategies

### Redis Caching

```go
type CacheService struct {
    client *redis.Client
}

// Multi-tier caching strategy
func (c *CacheService) GetLink(ctx context.Context, shortCode string) (*Link, error) {
    key := fmt.Sprintf("link:%s", shortCode)

    // Try cache first
    data, err := c.client.Get(ctx, key).Bytes()
    if err == nil {
        var link Link
        if err := json.Unmarshal(data, &link); err == nil {
            return &link, nil
        }
    }

    return nil, ErrCacheMiss
}

func (c *CacheService) SetLink(ctx context.Context, link *Link) error {
    key := fmt.Sprintf("link:%s", link.ShortCode)
    data, err := json.Marshal(link)
    if err != nil {
        return err
    }

    // Set with TTL based on link popularity
    ttl := c.calculateTTL(link)
    return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *CacheService) calculateTTL(link *Link) time.Duration {
    // Hot links: shorter TTL to keep fresh
    if link.ClicksLast24h > 1000 {
        return 5 * time.Minute
    }
    // Normal links: moderate TTL
    if link.ClicksLast24h > 100 {
        return 1 * time.Hour
    }
    // Cold links: longer TTL
    return 24 * time.Hour
}
```

### HTTP Caching

```go
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Set cache headers for static assets
    w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

    // For API responses
    w.Header().Set("Cache-Control", "private, max-age=60")
    w.Header().Set("ETag", calculateETag(data))

    // Check If-None-Match for conditional requests
    if r.Header.Get("If-None-Match") == etag {
        w.WriteHeader(http.StatusNotModified)
        return
    }
}
```

### CDN Configuration

**Cloudflare Page Rules:**

```
# Cache redirects at edge
URL: lnkr.ft/*
Cache Level: Cache Everything
Edge Cache TTL: 1 hour
Browser Cache TTL: 1 minute

# Cache static assets aggressively
URL: app.linkrift.io/assets/*
Cache Level: Cache Everything
Edge Cache TTL: 1 year
Browser Cache TTL: 1 year
```

---

## Database Optimization

**Indexes:**

```sql
-- Essential indexes for redirect performance
CREATE INDEX CONCURRENTLY idx_links_short_code ON links(short_code);
CREATE INDEX CONCURRENTLY idx_links_short_code_active ON links(short_code) WHERE deleted_at IS NULL;

-- Analytics indexes
CREATE INDEX CONCURRENTLY idx_clicks_link_created ON link_clicks(link_id, created_at DESC);
CREATE INDEX CONCURRENTLY idx_clicks_created ON link_clicks(created_at DESC);

-- Composite index for common queries
CREATE INDEX CONCURRENTLY idx_links_user_created ON links(user_id, created_at DESC);
```

**Query Optimization:**

```sql
-- Use EXPLAIN ANALYZE to understand query performance
EXPLAIN ANALYZE SELECT * FROM links WHERE short_code = 'abc123';

-- Avoid SELECT *
SELECT id, short_code, original_url FROM links WHERE short_code = 'abc123';

-- Use prepared statements
PREPARE get_link AS SELECT id, short_code, original_url FROM links WHERE short_code = $1;
EXECUTE get_link('abc123');
```

**Connection Pooling:**

```go
import "github.com/jackc/pgx/v5/pgxpool"

config, _ := pgxpool.ParseConfig(connString)
config.MaxConns = 100
config.MinConns = 10
config.MaxConnLifetime = 30 * time.Minute
config.MaxConnIdleTime = 5 * time.Minute
config.HealthCheckPeriod = 1 * time.Minute

pool, err := pgxpool.NewWithConfig(ctx, config)
```

---

## Infrastructure Tuning

### Linux Kernel Parameters

```bash
# /etc/sysctl.conf

# Network performance
net.core.somaxconn = 65535
net.core.netdev_max_backlog = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_fin_timeout = 10
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_keepalive_time = 600
net.ipv4.tcp_keepalive_probes = 5
net.ipv4.tcp_keepalive_intvl = 15

# File descriptors
fs.file-max = 2097152
fs.nr_open = 2097152

# Memory
vm.swappiness = 10
vm.dirty_ratio = 60
vm.dirty_background_ratio = 2
```

### Go Runtime Tuning

```bash
# Environment variables
GOMAXPROCS=4                    # Match CPU cores
GOGC=100                        # GC target percentage (default)
GOMEMLIMIT=3750MiB             # Memory limit for 4GB container
```

---

## Monitoring & Benchmarking

### Key Metrics to Track

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    redirectLatency = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "linkrift_redirect_duration_seconds",
            Help:    "Redirect request latency",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
        },
        []string{"status"},
    )

    cacheHitRate = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "linkrift_cache_requests_total",
            Help: "Cache request counts",
        },
        []string{"result"}, // "hit" or "miss"
    )

    activeConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "linkrift_active_connections",
            Help: "Current active connections",
        },
    )
)
```

### Benchmark Tests

```go
func BenchmarkRedirectHandler(b *testing.B) {
    handler := NewRedirectHandler(mockRepo, mockCache)

    b.ReportAllocs()
    b.ResetTimer()

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            req := httptest.NewRequest("GET", "/abc123", nil)
            rec := httptest.NewRecorder()
            handler.ServeHTTP(rec, req)
        }
    })
}
```

Run benchmarks:

```bash
go test -bench=. -benchmem -benchtime=10s ./...
```

---

## Summary

Key optimization principles for Linkrift:

1. **Profile First**: Always measure before optimizing
2. **Optimize Hot Paths**: Focus on redirect and lookup operations
3. **Cache Aggressively**: Use Redis and HTTP caching effectively
4. **Minimize Allocations**: Use pools and pre-allocation
5. **Right-Size Infrastructure**: Scale horizontally for stateless services
6. **Monitor Continuously**: Track latency, throughput, and errors

For questions or performance issues, open a GitHub issue or join our Discord community.
