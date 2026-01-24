# Troubleshooting Guide

**Last Updated: 2025-01-24**

This document provides comprehensive troubleshooting procedures for common issues in Linkrift, including debugging techniques, log analysis, and recovery procedures.

---

## Table of Contents

- [Overview](#overview)
- [Go Runtime Issues](#go-runtime-issues)
  - [Memory Issues](#memory-issues)
  - [Goroutine Leaks](#goroutine-leaks)
  - [Panic Recovery](#panic-recovery)
  - [Performance Issues](#performance-issues)
- [Vite Build Issues](#vite-build-issues)
  - [Build Failures](#build-failures)
  - [Hot Module Replacement Issues](#hot-module-replacement-issues)
  - [Asset Processing Issues](#asset-processing-issues)
  - [Dependency Issues](#dependency-issues)
- [Database Issues](#database-issues)
  - [Connection Issues](#connection-issues)
  - [Query Performance](#query-performance)
  - [Lock Contention](#lock-contention)
  - [Replication Issues](#replication-issues)
- [Redirect Service Issues](#redirect-service-issues)
  - [High Latency](#high-latency)
  - [404 Errors](#404-errors)
  - [Cache Issues](#cache-issues)
  - [Rate Limiting](#rate-limiting)
- [Custom Domain Issues](#custom-domain-issues)
  - [DNS Configuration](#dns-configuration)
  - [SSL Certificate Issues](#ssl-certificate-issues)
  - [Domain Verification](#domain-verification)
- [Authentication Issues](#authentication-issues)
  - [PASETO Token Issues](#paseto-token-issues)
  - [Session Management](#session-management)
  - [OAuth Integration](#oauth-integration)
- [Debugging with Delve](#debugging-with-delve)
  - [Setting Up Delve](#setting-up-delve)
  - [Common Debugging Scenarios](#common-debugging-scenarios)
  - [Remote Debugging](#remote-debugging)
- [Profiling with pprof](#profiling-with-pprof)
  - [CPU Profiling](#cpu-profiling)
  - [Memory Profiling](#memory-profiling)
  - [Goroutine Analysis](#goroutine-analysis)
- [Log Analysis](#log-analysis)
  - [Finding Relevant Logs](#finding-relevant-logs)
  - [Log Patterns](#log-patterns)
  - [Correlation IDs](#correlation-ids)
- [Recovery Procedures](#recovery-procedures)
  - [Service Recovery](#service-recovery)
  - [Data Recovery](#data-recovery)
  - [Emergency Procedures](#emergency-procedures)

---

## Overview

This guide follows a systematic approach to troubleshooting:

1. **Identify** - Recognize symptoms and categorize the issue
2. **Investigate** - Gather logs, metrics, and diagnostic information
3. **Isolate** - Narrow down the root cause
4. **Resolve** - Apply the fix
5. **Verify** - Confirm the issue is resolved
6. **Document** - Record the issue and resolution

---

## Go Runtime Issues

### Memory Issues

#### Symptoms
- High memory usage in metrics
- OOM (Out of Memory) kills
- Slow garbage collection

#### Investigation

```bash
# Check memory metrics
curl -s http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof -top heap.prof

# Check GC stats
curl -s http://localhost:6060/debug/pprof/heap?debug=1 | head -50

# Check for memory growth over time
watch -n 5 'curl -s http://localhost:8080/metrics | grep go_memstats'
```

```go
// Add memory debugging
package main

import (
    "fmt"
    "runtime"
)

func printMemStats() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    fmt.Printf("Alloc = %v MiB\n", m.Alloc/1024/1024)
    fmt.Printf("TotalAlloc = %v MiB\n", m.TotalAlloc/1024/1024)
    fmt.Printf("Sys = %v MiB\n", m.Sys/1024/1024)
    fmt.Printf("NumGC = %v\n", m.NumGC)
    fmt.Printf("HeapObjects = %v\n", m.HeapObjects)
}
```

#### Common Causes and Solutions

| Cause | Solution |
|-------|----------|
| Unbounded slice growth | Use fixed-size buffers or pools |
| Leaked goroutines | Add context cancellation |
| Large response bodies | Stream responses, use pagination |
| Cached data growth | Implement TTL and max size limits |

```go
// Fix: Use sync.Pool for frequently allocated objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

func processRequest(data []byte) {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf)
    // Use buf...
}
```

### Goroutine Leaks

#### Symptoms
- Increasing goroutine count
- Slow response times
- Memory growth

#### Investigation

```bash
# Get goroutine dump
curl -s http://localhost:6060/debug/pprof/goroutine?debug=2 > goroutines.txt

# Count goroutines by state
curl -s http://localhost:6060/debug/pprof/goroutine?debug=1 | grep -E "^[0-9]+ @"

# Watch goroutine count
watch -n 5 'curl -s http://localhost:8080/metrics | grep go_goroutines'
```

```go
// Analyze goroutine dump
// Look for patterns like:
// - Many goroutines waiting on same channel
// - Blocked on network I/O
// - Waiting for locks
```

#### Common Causes and Solutions

```go
// Problem: Goroutine leak from unclosed channel
func leaky() {
    ch := make(chan int)
    go func() {
        val := <-ch  // Blocks forever if nothing sent
        fmt.Println(val)
    }()
}

// Solution: Use context for cancellation
func fixed(ctx context.Context) {
    ch := make(chan int)
    go func() {
        select {
        case val := <-ch:
            fmt.Println(val)
        case <-ctx.Done():
            return
        }
    }()
}

// Problem: HTTP client without timeout
func leakyHTTP() {
    resp, _ := http.Get("https://example.com")
    // May hang forever
}

// Solution: Use context with timeout
func fixedHTTP(ctx context.Context) {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, "GET", "https://example.com", nil)
    resp, _ := http.DefaultClient.Do(req)
    defer resp.Body.Close()
}
```

### Panic Recovery

#### Investigation

```bash
# Check application logs for panics
grep -i "panic" /var/log/linkrift/app.log

# Check system logs
journalctl -u linkrift | grep -i panic
```

#### Recovery Middleware

```go
// internal/middleware/recovery.go
package middleware

import (
    "net/http"
    "runtime/debug"

    "go.uber.org/zap"
    "linkrift/internal/logger"
    "linkrift/internal/metrics"
)

func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                stack := debug.Stack()

                logger.Log.Error("Panic recovered",
                    zap.Any("error", err),
                    zap.String("stack", string(stack)),
                    zap.String("path", r.URL.Path),
                    zap.String("method", r.Method),
                )

                metrics.ErrorsTotal.WithLabelValues("panic", "http").Inc()

                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### Performance Issues

#### Investigation

```bash
# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Trace
curl -o trace.out 'http://localhost:6060/debug/pprof/trace?seconds=5'
go tool trace trace.out

# Block profile
go tool pprof http://localhost:6060/debug/pprof/block
```

#### Common Performance Issues

```go
// Problem: String concatenation in loop
func slow(items []string) string {
    result := ""
    for _, item := range items {
        result += item + ","  // Creates new string each iteration
    }
    return result
}

// Solution: Use strings.Builder
func fast(items []string) string {
    var builder strings.Builder
    for _, item := range items {
        builder.WriteString(item)
        builder.WriteString(",")
    }
    return builder.String()
}

// Problem: Unnecessary allocations
func slowProcess(data []byte) {
    str := string(data)  // Allocates new string
    // process str
}

// Solution: Use unsafe conversion when safe
import "unsafe"

func fastProcess(data []byte) {
    str := unsafe.String(&data[0], len(data))  // No allocation
    // process str (don't modify data while using str)
}
```

---

## Vite Build Issues

### Build Failures

#### Symptoms
- `npm run build` fails
- TypeScript errors
- Module resolution errors

#### Investigation

```bash
# Clear cache and reinstall
rm -rf node_modules
rm package-lock.json
npm install

# Check for conflicting versions
npm ls

# Verbose build output
npm run build -- --debug

# Check TypeScript errors
npx tsc --noEmit
```

#### Common Causes and Solutions

```bash
# Issue: Node version mismatch
node --version
# Should be 18+ for Vite 5

# Solution: Use correct Node version
nvm use 20

# Issue: Missing peer dependencies
npm install --legacy-peer-deps

# Issue: Corrupted cache
npm cache clean --force
rm -rf node_modules/.vite
```

### Hot Module Replacement Issues

#### Symptoms
- Changes not reflecting in browser
- Full page reloads instead of HMR
- WebSocket connection errors

#### Investigation

```javascript
// Check Vite config
// vite.config.ts
export default defineConfig({
  server: {
    hmr: {
      overlay: true,  // Show error overlay
    },
  },
});
```

```bash
# Check if HMR port is accessible
curl -v http://localhost:5173/__vite_hmr

# Check browser console for WebSocket errors
# Look for: "WebSocket connection to 'ws://localhost:5173/' failed"
```

#### Solutions

```typescript
// vite.config.ts - Fix HMR for Docker/WSL
export default defineConfig({
  server: {
    hmr: {
      host: 'localhost',
      port: 5173,
      clientPort: 5173,
    },
    watch: {
      usePolling: true,  // For Docker/WSL
    },
  },
});
```

### Asset Processing Issues

#### Symptoms
- Images not loading
- CSS not applied
- Font loading failures

#### Investigation

```bash
# Check build output
ls -la dist/assets/

# Verify asset paths in built files
grep -r "assets/" dist/

# Check public directory
ls -la public/
```

#### Solutions

```typescript
// vite.config.ts - Configure asset handling
export default defineConfig({
  build: {
    assetsDir: 'assets',
    assetsInlineLimit: 4096,  // Inline assets < 4KB
  },
  resolve: {
    alias: {
      '@': '/src',
      '@assets': '/src/assets',
    },
  },
});
```

### Dependency Issues

#### Symptoms
- Module not found errors
- Version conflicts
- ESM/CJS compatibility issues

#### Investigation

```bash
# Check for duplicate packages
npm ls react

# Find peer dependency issues
npm ls 2>&1 | grep -i "peer dep"

# Check package exports
node -e "console.log(require.resolve('package-name'))"
```

#### Solutions

```typescript
// vite.config.ts - Fix CommonJS dependencies
export default defineConfig({
  optimizeDeps: {
    include: ['problematic-cjs-package'],
  },
  build: {
    commonjsOptions: {
      transformMixedEsModules: true,
    },
  },
});
```

---

## Database Issues

### Connection Issues

#### Symptoms
- "connection refused" errors
- Timeouts
- Pool exhaustion

#### Investigation

```bash
# Check PostgreSQL status
systemctl status postgresql

# Check connections
psql -U linkrift -d linkrift -c "
SELECT count(*), state
FROM pg_stat_activity
WHERE datname = 'linkrift'
GROUP BY state;
"

# Check max connections
psql -U postgres -c "SHOW max_connections;"

# Check connection pool metrics
curl -s http://localhost:8080/metrics | grep db_connections
```

#### Solutions

```go
// Configure connection pool properly
db, err := sql.Open("postgres", connStr)
if err != nil {
    return err
}

db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(5 * time.Minute)
```

```sql
-- Increase max connections if needed
ALTER SYSTEM SET max_connections = 200;
-- Requires restart

-- Check blocked connections
SELECT blocked_locks.pid AS blocked_pid,
       blocking_locks.pid AS blocking_pid,
       blocked_activity.query AS blocked_query,
       blocking_activity.query AS blocking_query
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;
```

### Query Performance

#### Symptoms
- Slow API responses
- High CPU on database server
- Timeout errors

#### Investigation

```sql
-- Find slow queries
SELECT query, calls, total_time, mean_time, rows
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 20;

-- Check for missing indexes
SELECT relname, seq_scan, idx_scan, seq_tup_read
FROM pg_stat_user_tables
WHERE seq_scan > idx_scan
ORDER BY seq_tup_read DESC;

-- Analyze query plan
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT * FROM links WHERE short_code = 'abc123';
```

#### Solutions

```sql
-- Add missing index
CREATE INDEX CONCURRENTLY idx_links_short_code ON links(short_code);

-- Update statistics
ANALYZE links;

-- Optimize query
-- Before: SELECT * FROM links WHERE LOWER(short_code) = 'abc123'
-- After: Use case-insensitive collation or functional index
CREATE INDEX idx_links_short_code_lower ON links(LOWER(short_code));
```

### Lock Contention

#### Symptoms
- Queries waiting
- Deadlock errors
- Slow writes

#### Investigation

```sql
-- Check for locks
SELECT
    pg_stat_activity.pid,
    pg_stat_activity.query,
    pg_locks.mode,
    pg_locks.granted
FROM pg_stat_activity
JOIN pg_locks ON pg_stat_activity.pid = pg_locks.pid
WHERE pg_locks.granted = false;

-- Check lock waits
SELECT * FROM pg_stat_activity WHERE wait_event_type = 'Lock';

-- Find blocking queries
SELECT blocked.pid AS blocked_pid,
       blocked.query AS blocked_query,
       blocking.pid AS blocking_pid,
       blocking.query AS blocking_query
FROM pg_stat_activity blocked
JOIN pg_stat_activity blocking ON blocking.pid = ANY(pg_blocking_pids(blocked.pid));
```

#### Solutions

```go
// Use advisory locks for critical sections
func (s *Service) CreateLink(ctx context.Context, userID string, url string) error {
    // Acquire advisory lock
    _, err := s.db.ExecContext(ctx, "SELECT pg_advisory_lock($1)", hashUserID(userID))
    if err != nil {
        return err
    }
    defer s.db.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", hashUserID(userID))

    // Critical section
    return s.createLinkInternal(ctx, userID, url)
}
```

### Replication Issues

#### Symptoms
- Replication lag
- Stale reads
- Replica disconnects

#### Investigation

```sql
-- On primary: Check replication status
SELECT
    client_addr,
    state,
    sent_lsn,
    write_lsn,
    flush_lsn,
    replay_lsn,
    pg_wal_lsn_diff(sent_lsn, replay_lsn) AS replication_lag
FROM pg_stat_replication;

-- On replica: Check replication status
SELECT
    pg_is_in_recovery(),
    pg_last_wal_receive_lsn(),
    pg_last_wal_replay_lsn(),
    pg_last_xact_replay_timestamp();
```

---

## Redirect Service Issues

### High Latency

#### Symptoms
- P99 latency > 100ms
- Slow redirects
- Timeout errors

#### Investigation

```bash
# Check redirect latency metrics
curl -s http://localhost:8080/metrics | grep redirect_latency

# Profile redirect endpoint
curl -w "@curl-format.txt" -o /dev/null -s "http://localhost:8080/abc123"

# Check cache hit rate
curl -s http://localhost:8080/metrics | grep cache
```

```go
// Add detailed timing logs
func (s *RedirectService) Resolve(ctx context.Context, code string) (string, error) {
    start := time.Now()
    defer func() {
        logger.FromContext(ctx).Debug("Redirect timing",
            zap.String("code", code),
            zap.Duration("total", time.Since(start)),
        )
    }()

    // Cache lookup
    cacheStart := time.Now()
    url, err := s.cache.Get(ctx, code)
    logger.FromContext(ctx).Debug("Cache lookup",
        zap.Duration("duration", time.Since(cacheStart)),
        zap.Bool("hit", err == nil),
    )

    if err == nil {
        return url, nil
    }

    // Database lookup
    dbStart := time.Now()
    link, err := s.repo.GetByCode(ctx, code)
    logger.FromContext(ctx).Debug("Database lookup",
        zap.Duration("duration", time.Since(dbStart)),
    )

    return link.URL, err
}
```

#### Solutions

```go
// Implement connection pooling for Redis
var redisPool = &redis.Pool{
    MaxIdle:     10,
    MaxActive:   100,
    IdleTimeout: 240 * time.Second,
    Dial: func() (redis.Conn, error) {
        return redis.Dial("tcp", "localhost:6379")
    },
}

// Use read replicas for database
func (r *Repository) GetByCode(ctx context.Context, code string) (*Link, error) {
    // Use read replica for redirects
    return r.readDB.GetByCode(ctx, code)
}
```

### 404 Errors

#### Symptoms
- Valid links returning 404
- Intermittent 404s
- 404 after recent creation

#### Investigation

```bash
# Check if link exists in database
psql -U linkrift -d linkrift -c "
SELECT * FROM links WHERE short_code = 'abc123';
"

# Check cache state
redis-cli GET "link:abc123"

# Check application logs
grep "abc123" /var/log/linkrift/app.log | tail -20
```

#### Common Causes

```go
// Issue: Race condition between cache invalidation and write
// Solution: Use write-through caching
func (s *Service) CreateLink(ctx context.Context, link *Link) error {
    // Write to database
    if err := s.repo.Create(ctx, link); err != nil {
        return err
    }

    // Immediately populate cache
    return s.cache.Set(ctx, link.ShortCode, link.OriginalURL, 24*time.Hour)
}

// Issue: Case sensitivity
// Solution: Normalize short codes
func normalizeCode(code string) string {
    return strings.ToLower(strings.TrimSpace(code))
}
```

### Cache Issues

#### Symptoms
- Low cache hit rate
- Stale data served
- Cache misses for popular links

#### Investigation

```bash
# Check Redis memory
redis-cli INFO memory

# Check eviction stats
redis-cli INFO stats | grep evicted

# Check TTL of a key
redis-cli TTL "link:abc123"

# Check cache hit/miss metrics
curl -s http://localhost:8080/metrics | grep -E "cache_(hits|misses)"
```

#### Solutions

```go
// Implement cache warming for popular links
func (s *Service) WarmCache(ctx context.Context) error {
    // Get top 1000 links by click count
    links, err := s.repo.GetTopLinks(ctx, 1000)
    if err != nil {
        return err
    }

    pipe := s.redis.Pipeline()
    for _, link := range links {
        pipe.Set(ctx, "link:"+link.ShortCode, link.OriginalURL, 24*time.Hour)
    }
    _, err = pipe.Exec(ctx)
    return err
}

// Implement cache-aside with single-flight
var group singleflight.Group

func (s *Service) GetURL(ctx context.Context, code string) (string, error) {
    // Check cache
    if url, err := s.cache.Get(ctx, code); err == nil {
        return url, nil
    }

    // Single-flight to prevent thundering herd
    result, err, _ := group.Do(code, func() (interface{}, error) {
        link, err := s.repo.GetByCode(ctx, code)
        if err != nil {
            return "", err
        }

        // Populate cache
        s.cache.Set(ctx, code, link.URL, 24*time.Hour)
        return link.URL, nil
    })

    return result.(string), err
}
```

### Rate Limiting

#### Symptoms
- 429 Too Many Requests errors
- Legitimate traffic blocked
- Inconsistent limiting

#### Investigation

```bash
# Check rate limit metrics
curl -s http://localhost:8080/metrics | grep rate_limit

# Check Redis rate limit keys
redis-cli KEYS "ratelimit:*" | head -20

# Test rate limit
for i in {1..150}; do
    curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/abc123
done | sort | uniq -c
```

#### Solutions

```go
// Implement sliding window rate limiting
type RateLimiter struct {
    redis  *redis.Client
    limit  int
    window time.Duration
}

func (rl *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
    now := time.Now().UnixNano()
    windowStart := now - rl.window.Nanoseconds()

    pipe := rl.redis.Pipeline()

    // Remove old entries
    pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

    // Count current entries
    countCmd := pipe.ZCard(ctx, key)

    // Add current request
    pipe.ZAdd(ctx, key, &redis.Z{Score: float64(now), Member: now})

    // Set expiry
    pipe.Expire(ctx, key, rl.window)

    _, err := pipe.Exec(ctx)
    if err != nil {
        return false, err
    }

    return countCmd.Val() < int64(rl.limit), nil
}
```

---

## Custom Domain Issues

### DNS Configuration

#### Symptoms
- Custom domain not resolving
- SSL certificate errors
- Redirect loops

#### Investigation

```bash
# Check DNS resolution
dig +short custom.example.com

# Check CNAME record
dig CNAME custom.example.com

# Verify expected value
dig TXT _linkrift.custom.example.com

# Check from DNS propagation
curl "https://dns.google/resolve?name=custom.example.com&type=A"
```

#### Solutions

```markdown
## Required DNS Configuration

For custom domain `custom.example.com`:

1. **CNAME Record**
   ```
   custom.example.com CNAME linkrift.io
   ```

2. **Verification TXT Record**
   ```
   _linkrift.custom.example.com TXT "linkrift-verify=abc123"
   ```

3. **Alternative: A Record** (if CNAME not possible)
   ```
   custom.example.com A 203.0.113.10
   ```
```

### SSL Certificate Issues

#### Symptoms
- Certificate errors in browser
- Let's Encrypt failures
- Certificate expiration

#### Investigation

```bash
# Check certificate details
echo | openssl s_client -servername custom.example.com -connect custom.example.com:443 2>/dev/null | openssl x509 -noout -text

# Check certificate expiry
echo | openssl s_client -servername custom.example.com -connect custom.example.com:443 2>/dev/null | openssl x509 -noout -dates

# Check Let's Encrypt logs
journalctl -u certbot

# Check ACME challenges
curl http://custom.example.com/.well-known/acme-challenge/test
```

#### Solutions

```bash
# Manually renew certificate
certbot certonly --nginx -d custom.example.com

# Check certificate renewal
certbot renew --dry-run

# Force certificate renewal
certbot renew --force-renewal -d custom.example.com
```

```go
// Automated certificate management
type CertManager struct {
    acmeClient *acme.Client
    storage    CertStorage
}

func (cm *CertManager) EnsureCertificate(ctx context.Context, domain string) error {
    // Check if certificate exists and is valid
    cert, err := cm.storage.Get(ctx, domain)
    if err == nil && time.Until(cert.NotAfter) > 30*24*time.Hour {
        return nil  // Certificate is valid
    }

    // Request new certificate
    return cm.obtainCertificate(ctx, domain)
}
```

### Domain Verification

#### Symptoms
- Domain verification fails
- Domain shows as unverified
- Verification token mismatch

#### Investigation

```bash
# Check verification record
dig TXT _linkrift.custom.example.com

# Compare with expected
psql -U linkrift -d linkrift -c "
SELECT domain, verification_token, verified_at
FROM custom_domains
WHERE domain = 'custom.example.com';
"
```

#### Solutions

```go
// Domain verification service
func (s *DomainService) Verify(ctx context.Context, domain string) error {
    // Get expected token
    record, err := s.repo.GetByDomain(ctx, domain)
    if err != nil {
        return err
    }

    // Query DNS
    txtRecords, err := net.LookupTXT("_linkrift." + domain)
    if err != nil {
        return fmt.Errorf("DNS lookup failed: %w", err)
    }

    // Check for matching token
    expected := "linkrift-verify=" + record.VerificationToken
    for _, txt := range txtRecords {
        if txt == expected {
            return s.repo.MarkVerified(ctx, domain)
        }
    }

    return fmt.Errorf("verification token not found")
}
```

---

## Authentication Issues

### PASETO Token Issues

#### Symptoms
- "invalid token" errors
- Token expiration issues
- Token parsing failures

#### Investigation

```go
// Debug token validation
func debugToken(tokenString string) {
    parser := paseto.NewParser()

    token, err := parser.ParseV4Local(symmetricKey, tokenString, nil)
    if err != nil {
        fmt.Printf("Parse error: %v\n", err)
        return
    }

    fmt.Printf("Claims: %+v\n", token.Claims())
    fmt.Printf("Expiration: %v\n", token.Expiration())
    fmt.Printf("IssuedAt: %v\n", token.IssuedAt())
}
```

```bash
# Check token in request
curl -v -H "Authorization: Bearer <token>" http://localhost:8080/api/links

# Decode token (for debugging only - never log production tokens)
echo "<token>" | base64 -d
```

#### Common Causes and Solutions

```go
// Issue: Clock skew between services
// Solution: Add clock tolerance
parser := paseto.NewParser(
    paseto.WithIssuer("linkrift"),
    paseto.WithNotBefore(),
    paseto.WithClockTolerance(30 * time.Second),  // Allow 30 second skew
)

// Issue: Key rotation
// Solution: Support multiple keys
type TokenValidator struct {
    currentKey  []byte
    previousKey []byte  // For graceful rotation
}

func (v *TokenValidator) Validate(token string) (*Claims, error) {
    // Try current key first
    claims, err := v.validateWithKey(token, v.currentKey)
    if err == nil {
        return claims, nil
    }

    // Fall back to previous key
    if v.previousKey != nil {
        return v.validateWithKey(token, v.previousKey)
    }

    return nil, err
}
```

### Session Management

#### Symptoms
- Unexpected logouts
- Session not persisting
- Concurrent session issues

#### Investigation

```bash
# Check session in Redis
redis-cli GET "session:abc123"

# Check session TTL
redis-cli TTL "session:abc123"

# List all sessions for user
redis-cli KEYS "session:user:*"
```

#### Solutions

```go
// Implement sliding session expiration
func (s *SessionService) Validate(ctx context.Context, sessionID string) (*Session, error) {
    session, err := s.store.Get(ctx, sessionID)
    if err != nil {
        return nil, err
    }

    // Extend session on each valid request
    if time.Until(session.ExpiresAt) < 15*time.Minute {
        session.ExpiresAt = time.Now().Add(s.sessionDuration)
        _ = s.store.Set(ctx, sessionID, session, s.sessionDuration)
    }

    return session, nil
}
```

### OAuth Integration

#### Symptoms
- OAuth callback failures
- State mismatch errors
- Token exchange failures

#### Investigation

```bash
# Check OAuth logs
grep -i "oauth" /var/log/linkrift/app.log | tail -50

# Verify redirect URI
echo $GOOGLE_REDIRECT_URI

# Test OAuth endpoint
curl -v "http://localhost:8080/auth/google"
```

#### Solutions

```go
// Implement proper state handling
func (h *OAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
    // Verify state parameter
    state := r.URL.Query().Get("state")

    storedState, err := h.stateStore.Get(r.Context(), state)
    if err != nil {
        http.Error(w, "Invalid state", http.StatusBadRequest)
        return
    }

    // Delete used state
    h.stateStore.Delete(r.Context(), state)

    // Check state expiration
    if time.Since(storedState.CreatedAt) > 10*time.Minute {
        http.Error(w, "State expired", http.StatusBadRequest)
        return
    }

    // Continue with token exchange...
}
```

---

## Debugging with Delve

### Setting Up Delve

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Build with debug symbols
go build -gcflags="all=-N -l" -o linkrift ./cmd/server

# Start with Delve
dlv exec ./linkrift

# Or attach to running process
dlv attach $(pgrep linkrift)
```

### Common Debugging Scenarios

```bash
# Set breakpoint
(dlv) break internal/service/redirect.go:42
(dlv) break main.main

# Conditional breakpoint
(dlv) break internal/service/redirect.go:42 if code == "abc123"

# Continue execution
(dlv) continue

# Step through code
(dlv) next       # Step over
(dlv) step       # Step into
(dlv) stepout    # Step out

# Print variables
(dlv) print code
(dlv) print link

# Print goroutines
(dlv) goroutines
(dlv) goroutine 1  # Switch to goroutine 1

# Print stack trace
(dlv) stack
(dlv) stack -full  # With local variables
```

### Remote Debugging

```bash
# Start server with Delve in headless mode
dlv exec --headless --listen=:2345 --api-version=2 ./linkrift

# Connect from another machine
dlv connect remote-server:2345

# VS Code launch.json for remote debugging
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Remote Debug",
            "type": "go",
            "request": "attach",
            "mode": "remote",
            "remotePath": "/app",
            "port": 2345,
            "host": "remote-server"
        }
    ]
}
```

---

## Profiling with pprof

### CPU Profiling

```bash
# Collect 30-second CPU profile
curl -o cpu.prof 'http://localhost:6060/debug/pprof/profile?seconds=30'

# Analyze
go tool pprof -http=:8080 cpu.prof

# Common commands
go tool pprof cpu.prof
(pprof) top20
(pprof) top -cum  # Cumulative time
(pprof) list functionName
(pprof) web  # Open in browser
```

### Memory Profiling

```bash
# Current heap
curl -o heap.prof http://localhost:6060/debug/pprof/heap

# Allocation tracking
curl -o allocs.prof http://localhost:6060/debug/pprof/allocs

# Analyze
go tool pprof -http=:8080 heap.prof

# Compare two profiles
go tool pprof -base heap1.prof heap2.prof
```

### Goroutine Analysis

```bash
# Get goroutine dump
curl -o goroutine.prof http://localhost:6060/debug/pprof/goroutine

# Text format (human readable)
curl 'http://localhost:6060/debug/pprof/goroutine?debug=2' > goroutines.txt

# Analyze
go tool pprof goroutine.prof
(pprof) top
(pprof) traces
```

---

## Log Analysis

### Finding Relevant Logs

```bash
# Search by request ID
grep "request_id\":\"abc123" /var/log/linkrift/app.log

# Search by time range
grep "2025-01-24T10:" /var/log/linkrift/app.log

# Search errors in last hour
grep '"level":"error"' /var/log/linkrift/app.log | \
    jq 'select(.timestamp > "2025-01-24T09:00:00")'

# Count errors by type
grep '"level":"error"' /var/log/linkrift/app.log | \
    jq -r '.msg' | sort | uniq -c | sort -rn
```

### Log Patterns

```bash
# Find slow requests (>100ms)
grep '"duration_ms"' /var/log/linkrift/app.log | \
    jq 'select(.duration_ms > 100)'

# Find 5xx errors
grep '"status":5' /var/log/linkrift/app.log

# Find failed redirects
grep '"msg":"Redirect failed"' /var/log/linkrift/app.log

# Find rate limiting events
grep '"rate_limit"' /var/log/linkrift/app.log
```

### Correlation IDs

```bash
# Trace request through system
REQUEST_ID="abc123-def456"

# Application logs
grep "$REQUEST_ID" /var/log/linkrift/app.log

# Nginx access logs
grep "$REQUEST_ID" /var/log/nginx/access.log

# Combine with trace ID for distributed tracing
TRACE_ID=$(grep "$REQUEST_ID" /var/log/linkrift/app.log | jq -r '.trace_id' | head -1)
grep "$TRACE_ID" /var/log/linkrift/app.log
```

---

## Recovery Procedures

### Service Recovery

```bash
#!/bin/bash
# scripts/recovery/service-recovery.sh

set -e

echo "=== Service Recovery ==="

# 1. Check service status
systemctl status linkrift || true

# 2. Check recent logs for errors
echo "Recent errors:"
journalctl -u linkrift --since "10 minutes ago" | grep -i error | tail -20

# 3. Restart service
echo "Restarting service..."
systemctl restart linkrift

# 4. Wait for startup
sleep 10

# 5. Verify health
if curl -sf http://localhost:8080/health; then
    echo "Service recovered successfully"
else
    echo "Service still unhealthy, checking detailed status..."
    systemctl status linkrift
    journalctl -u linkrift --since "1 minute ago"
fi
```

### Data Recovery

```bash
#!/bin/bash
# scripts/recovery/data-recovery.sh

# For accidental deletion
DELETED_CODE="abc123"

# Check if link was soft-deleted
psql -U linkrift -d linkrift -c "
SELECT * FROM links
WHERE short_code = '$DELETED_CODE'
AND deleted_at IS NOT NULL;
"

# Restore soft-deleted link
psql -U linkrift -d linkrift -c "
UPDATE links
SET deleted_at = NULL
WHERE short_code = '$DELETED_CODE';
"

# If hard deleted, check backup
pg_restore -t links /backups/postgres/latest.dump | \
    grep "$DELETED_CODE"
```

### Emergency Procedures

```bash
#!/bin/bash
# scripts/emergency/circuit-breaker.sh

# Emergency: Disable redirect service
# Use this when service is causing cascading failures

echo "Enabling circuit breaker..."

# Option 1: Return static response
iptables -A INPUT -p tcp --dport 8080 -j REJECT

# Option 2: Enable maintenance mode
redis-cli SET "maintenance_mode" "true" EX 3600

# Option 3: Update nginx to serve static page
ln -sf /etc/nginx/sites-available/maintenance /etc/nginx/sites-enabled/linkrift
nginx -s reload

echo "Circuit breaker enabled. Service will return maintenance page."
echo "To disable: run scripts/emergency/circuit-breaker-off.sh"
```

```bash
#!/bin/bash
# scripts/emergency/circuit-breaker-off.sh

echo "Disabling circuit breaker..."

# Remove iptables rule
iptables -D INPUT -p tcp --dport 8080 -j REJECT 2>/dev/null || true

# Disable maintenance mode
redis-cli DEL "maintenance_mode"

# Restore nginx config
ln -sf /etc/nginx/sites-available/linkrift /etc/nginx/sites-enabled/linkrift
nginx -s reload

# Verify service
curl -sf http://localhost:8080/health && echo "Service restored"
```

---

## Related Documentation

- [MONITORING_LOGGING.md](./MONITORING_LOGGING.md) - Monitoring and observability
- [MAINTENANCE.md](./MAINTENANCE.md) - Regular maintenance procedures
- [../security/SECURITY.md](../security/SECURITY.md) - Security configuration
