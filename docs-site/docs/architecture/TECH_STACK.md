# Tech Stack

> Last Updated: 2025-01-24

Detailed explanation of technology choices and rationale for Linkrift.

## Table of Contents

- [Overview](#overview)
- [Backend Technologies](#backend-technologies)
- [Frontend Technologies](#frontend-technologies)
- [Data Stores](#data-stores)
- [Infrastructure](#infrastructure)
- [External Services](#external-services)
- [Why Go?](#why-go)
- [Version Compatibility](#version-compatibility)

---

## Overview

| Layer | Technologies |
|-------|-------------|
| **Backend** | Go 1.22, Gin, sqlc, pgx, go-redis |
| **Frontend** | React 18, TypeScript, Vite 5, Tailwind CSS, Shadcn UI |
| **Database** | PostgreSQL 16, Redis 7, ClickHouse 24 |
| **Search** | Meilisearch 1.6 |
| **Infrastructure** | Docker, Kubernetes, NGINX, Cloudflare |
| **Monitoring** | Prometheus, Grafana, Sentry, OpenTelemetry |

---

## Backend Technologies

### Go 1.22

**Role**: Primary backend language

**Why Go over alternatives:**

| Criteria | Go | Node.js | Rust | Java |
|----------|-----|---------|------|------|
| Redirect Latency | ~0.5-2ms | ~5-15ms | ~0.3-1ms | ~3-10ms |
| Memory Usage | 10-30MB | 150-300MB | 5-20MB | 200-500MB |
| Cold Start | ~50ms | ~200-500ms | ~30ms | ~1-3s |
| Concurrency | Excellent | Good | Excellent | Good |
| Developer Velocity | High | Very High | Low | Medium |
| Deployment | Single binary | node_modules | Single binary | JAR + JVM |

**Go advantages for Linkrift:**

1. **Sub-millisecond redirects** — Critical for URL shorteners
2. **Tiny memory footprint** — More instances per server
3. **Single binary deployment** — No runtime dependencies
4. **Fast cold starts** — Better auto-scaling
5. **Built-in concurrency** — Goroutines for parallel processing
6. **Strong typing** — Catch errors at compile time

### Gin

**Role**: HTTP framework

**Why Gin over alternatives:**

| Framework | Performance | Middleware | Documentation | Community |
|-----------|-------------|------------|---------------|-----------|
| **Gin** | Very Fast | Rich | Excellent | Large |
| Fiber | Fastest | Good | Good | Medium |
| Echo | Fast | Rich | Good | Medium |
| Chi | Fast | Minimal | Good | Medium |

**Gin advantages:**
- Battle-tested at scale
- Excellent middleware ecosystem
- JSON validation with binding
- Request/response logging built-in
- Compatible with net/http handlers

### sqlc

**Role**: Type-safe SQL code generation

**Why sqlc over ORMs:**

| Approach | Type Safety | Performance | Flexibility | Learning Curve |
|----------|-------------|-------------|-------------|----------------|
| **sqlc** | Compile-time | Native SQL | Full SQL | Low |
| GORM | Runtime | Abstraction layer | Limited | Medium |
| Ent | Compile-time | Code generation | Schema-first | High |
| Raw SQL | None | Native SQL | Full SQL | N/A |

**sqlc workflow:**

```sql
-- Write SQL queries
-- name: GetLinkByCode :one
SELECT * FROM links WHERE short_code = $1;
```

```bash
# Generate Go code
make sqlc
```

```go
// Use type-safe functions
link, err := queries.GetLinkByCode(ctx, "abc123")
```

### pgx

**Role**: PostgreSQL driver

**Why pgx over lib/pq:**

- Native PostgreSQL protocol (not database/sql)
- Better performance (connection pooling, prepared statements)
- Full PostgreSQL feature support (LISTEN/NOTIFY, COPY)
- Context support throughout

### go-redis

**Role**: Redis client

**Features used:**
- Connection pooling
- Pipelining
- Pub/Sub for real-time features
- Cluster support for scaling

### Additional Go Libraries

| Library | Purpose |
|---------|---------|
| `uber/zap` | Structured logging |
| `spf13/viper` | Configuration management |
| `gorilla/websocket` | WebSocket connections |
| `hibiken/asynq` | Background job queue |
| `go-playground/validator` | Input validation |
| `swaggo/swag` | OpenAPI documentation |
| `golang-migrate/migrate` | Database migrations |

---

## Frontend Technologies

### React 18

**Role**: UI framework

**Why React:**
- Component-based architecture
- Large ecosystem
- Concurrent rendering (React 18)
- Strong TypeScript support

### TypeScript

**Role**: Type safety for JavaScript

**Benefits:**
- Catch errors at compile time
- Better IDE support
- Self-documenting code
- Safer refactoring

### Vite 5

**Role**: Build tool and dev server

**Why Vite over alternatives:**

| Tool | Dev Server Start | HMR Speed | Build Speed | Config |
|------|------------------|-----------|-------------|--------|
| **Vite** | ~300ms | Instant | Fast | Simple |
| Next.js | ~2-5s | Fast | Medium | Complex |
| Create React App | ~10-30s | Slow | Slow | Limited |

**Vite advantages:**
- Native ES modules in development
- Lightning-fast HMR
- Optimized production builds with Rollup
- First-class TypeScript support

### Tailwind CSS

**Role**: Utility-first CSS framework

**Why Tailwind:**
- No custom CSS files
- Consistent design system
- Small production bundle (purged unused styles)
- Easy responsive design

### Shadcn UI

**Role**: Component library

**Why Shadcn over alternatives:**

| Library | Customizable | Bundle Size | Accessibility | Style |
|---------|--------------|-------------|---------------|-------|
| **Shadcn** | Copy/paste | Zero (copied) | Radix-based | Any |
| Material UI | Theme only | Large | Good | Material |
| Chakra UI | Props | Medium | Good | Chakra |
| Ant Design | Theme only | Large | Good | Ant |

**Shadcn advantages:**
- Components copied into codebase (full control)
- Built on Radix primitives (accessible)
- Tailwind-styled (consistent)
- No runtime dependency

### State Management

#### TanStack Query (React Query)

**Role**: Server state management

```typescript
const { data, isLoading } = useQuery({
  queryKey: ['links', filters],
  queryFn: () => api.links.list(filters),
});
```

**Features:**
- Automatic caching
- Background refetching
- Optimistic updates
- Infinite scroll support

#### Zustand

**Role**: Client state management

```typescript
const useUIStore = create<UIState>((set) => ({
  sidebarOpen: true,
  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
}));
```

**Why Zustand over Redux:**
- Minimal boilerplate
- No providers required
- TypeScript-first
- Tiny bundle size

### Additional Frontend Libraries

| Library | Purpose |
|---------|---------|
| `react-router-dom` | Routing |
| `react-hook-form` | Form handling |
| `zod` | Schema validation |
| `recharts` | Charts and graphs |
| `date-fns` | Date formatting |
| `lucide-react` | Icons |

---

## Data Stores

### PostgreSQL 16

**Role**: Primary database

**Why PostgreSQL:**
- ACID compliance
- Advanced features (JSON, full-text search, CTEs)
- Excellent performance
- Strong ecosystem

**Features used:**
- UUID primary keys
- JSONB for flexible data
- Partial indexes for soft deletes
- Table partitioning for clicks

### Redis 7

**Role**: Caching, sessions, queues

**Use cases:**
- Link cache (5-minute TTL)
- Session storage
- Rate limiting (sliding window)
- Job queue (Asynq)
- Real-time counters
- Pub/Sub for WebSocket

### ClickHouse 24

**Role**: Analytics data warehouse

**Why ClickHouse for analytics:**

| Database | Write Speed | Query Speed | Compression | Cost |
|----------|-------------|-------------|-------------|------|
| **ClickHouse** | 1M rows/sec | Sub-second | 10-20x | Low |
| PostgreSQL | 10K rows/sec | Slow at scale | 2-3x | Medium |
| TimescaleDB | 100K rows/sec | Good | 5-10x | Medium |
| BigQuery | Variable | Fast | N/A | Pay per query |

**ClickHouse advantages:**
- Column-oriented storage
- Real-time ingestion
- Excellent compression
- Fast aggregations

### Meilisearch 1.6

**Role**: Full-text search

**Why Meilisearch:**
- Typo-tolerant search
- Instant results
- Easy setup
- RESTful API

---

## Infrastructure

### Docker

**Multi-stage builds for minimal images:**

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api

# Production stage
FROM scratch
COPY --from=builder /api /api
ENTRYPOINT ["/api"]
```

**Image sizes:**
- API: ~15MB
- Redirect: ~10MB
- Worker: ~15MB
- Web: ~20MB (NGINX + static)

### Kubernetes

**Why Kubernetes for production:**
- Horizontal auto-scaling
- Rolling deployments
- Self-healing
- Resource management

### NGINX

**Role**: Reverse proxy, static file serving

**Configuration highlights:**
- Gzip compression
- Caching headers
- Rate limiting backup
- WebSocket upgrade

### Cloudflare

**Role**: CDN, WAF, DNS

**Features used:**
- Edge caching
- DDoS protection
- SSL termination
- DNS management for custom domains

---

## External Services

| Service | Purpose | Alternative |
|---------|---------|-------------|
| **Stripe** | Payments | Paddle, Lemonsqueezy |
| **Resend** | Transactional email | Postmark, SendGrid |
| **Sentry** | Error tracking | Rollbar, Bugsnag |
| **MaxMind** | GeoIP | ip-api, IPinfo |
| **Cloudflare** | CDN/DNS | AWS CloudFront, Fastly |
| **AWS S3** | Object storage | MinIO (self-hosted) |

---

## Why Go?

### Performance Comparison

Benchmark: 10,000 concurrent redirect requests

| Metric | Go (Gin) | Node.js (Express) | Python (FastAPI) |
|--------|----------|-------------------|------------------|
| Requests/sec | 150,000 | 25,000 | 8,000 |
| Latency (p50) | 0.5ms | 3ms | 12ms |
| Latency (p99) | 2ms | 15ms | 50ms |
| Memory | 25MB | 180MB | 120MB |
| CPU | 15% | 45% | 80% |

### Memory Efficiency

```
Go instance:     ~20-30MB per pod
Node.js instance: ~150-300MB per pod

Same 1GB memory limit:
- Go: 33-50 pods
- Node.js: 3-6 pods
```

### Cold Start Times

| Runtime | Cold Start | Impact on Auto-scaling |
|---------|------------|------------------------|
| Go | ~50ms | Excellent |
| Node.js | ~200-500ms | Good |
| Java/Spring | ~1-3s | Poor |
| Python | ~100-300ms | Good |

### Deployment Simplicity

```bash
# Go: Single binary, no dependencies
./linkrift-api

# Node.js: Requires node_modules
npm install && node dist/index.js

# Java: Requires JVM
java -jar linkrift.jar
```

---

## Version Compatibility

### Minimum Versions

| Technology | Minimum | Recommended | Notes |
|------------|---------|-------------|-------|
| Go | 1.22 | 1.22+ | For `slices` package |
| Node.js | 20 LTS | 20 LTS | For pnpm 8 |
| PostgreSQL | 15 | 16 | For better JSON |
| Redis | 6.2 | 7.0 | For streams |
| ClickHouse | 23.3 | 24.1 | LTS versions |
| Docker | 24 | 24+ | For Compose V2 |

### Browser Support

| Browser | Minimum Version |
|---------|-----------------|
| Chrome | 90+ |
| Firefox | 88+ |
| Safari | 14+ |
| Edge | 90+ |

---

## Related Documentation

- [Architecture](ARCHITECTURE.md) — System design
- [Database Schema](DATABASE_SCHEMA.md) — Data model
- [Frontend Architecture](FRONTEND_ARCHITECTURE.md) — React patterns
- [Go Patterns](GO_PATTERNS.md) — Go code patterns
