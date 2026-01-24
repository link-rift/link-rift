# Architecture

> Last Updated: 2025-01-24

High-level system architecture and design of Linkrift.

## Table of Contents

- [System Overview](#system-overview)
- [High-Level Architecture](#high-level-architecture)
- [Service Architecture](#service-architecture)
- [Data Flow](#data-flow)
- [Caching Strategy](#caching-strategy)
- [Security Architecture](#security-architecture)
- [Scalability Considerations](#scalability-considerations)

---

## System Overview

Linkrift is designed as a collection of specialized services optimized for their specific workloads:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              LINKRIFT ARCHITECTURE                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                   │
│  │   Browser   │     │   Mobile    │     │   API       │                   │
│  │   Clients   │     │   Apps      │     │   Clients   │                   │
│  └──────┬──────┘     └──────┬──────┘     └──────┬──────┘                   │
│         │                   │                   │                          │
│         └───────────────────┼───────────────────┘                          │
│                             │                                              │
│                             ▼                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                         CLOUDFLARE (CDN/WAF)                         │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                             │                                              │
│         ┌───────────────────┼───────────────────┐                          │
│         │                   │                   │                          │
│         ▼                   ▼                   ▼                          │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                   │
│  │   NGINX     │     │   NGINX     │     │   NGINX     │                   │
│  │  (Web/API)  │     │ (Redirect)  │     │  (Static)   │                   │
│  └──────┬──────┘     └──────┬──────┘     └──────┬──────┘                   │
│         │                   │                   │                          │
│         ▼                   ▼                   ▼                          │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                   │
│  │   API       │     │  Redirect   │     │   React     │                   │
│  │  Service    │     │  Service    │     │   SPA       │                   │
│  │   (Go)      │     │   (Go)      │     │  (Vite)     │                   │
│  └──────┬──────┘     └──────┬──────┘     └─────────────┘                   │
│         │                   │                                              │
│         └───────────────────┼───────────────────────────┐                  │
│                             │                           │                  │
│         ┌───────────────────┼───────────────────┐       │                  │
│         │                   │                   │       │                  │
│         ▼                   ▼                   ▼       ▼                  │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────────┐               │
│  │  PostgreSQL │     │    Redis    │     │   ClickHouse    │               │
│  │  (Primary)  │     │   (Cache)   │     │  (Analytics)    │               │
│  └─────────────┘     └─────────────┘     └─────────────────┘               │
│                                                                             │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                   │
│  │   Worker    │     │  Scheduler  │     │ Meilisearch │                   │
│  │  (Asynq)    │     │   (Cron)    │     │  (Search)   │                   │
│  └─────────────┘     └─────────────┘     └─────────────┘                   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## High-Level Architecture

### Services

| Service | Port | Responsibility |
|---------|------|----------------|
| **API** | 8080 | REST API, authentication, business logic |
| **Redirect** | 8081 | High-performance URL redirects (<1ms) |
| **Worker** | - | Background job processing (Asynq) |
| **Scheduler** | - | Scheduled tasks (analytics rollup, cleanup) |
| **Web** | 3000 | React SPA dashboard |

### Data Stores

| Store | Purpose | Data |
|-------|---------|------|
| **PostgreSQL** | Primary database | Users, links, workspaces, domains |
| **Redis** | Caching, sessions, queues | Link cache, rate limits, job queue |
| **ClickHouse** | Analytics warehouse | Click events, aggregated metrics |
| **Meilisearch** | Full-text search | Link search, workspace search |
| **S3** | Object storage | QR codes, exports, uploads |

---

## Service Architecture

### API Service

The main API service handles all business logic:

```
┌────────────────────────────────────────────────────────────────┐
│                         API SERVICE                            │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    HTTP Layer (Gin)                      │  │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────────┐ │  │
│  │  │ Router  │ │ Middlew │ │ Handler │ │ Request/Response│ │  │
│  │  └────┬────┘ └────┬────┘ └────┬────┘ └────────┬────────┘ │  │
│  └───────┼──────────┼──────────┼────────────────┼───────────┘  │
│          │          │          │                │              │
│  ┌───────┴──────────┴──────────┴────────────────┴───────────┐  │
│  │                    Service Layer                         │  │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐           │  │
│  │  │ AuthSvc    │ │ LinkSvc    │ │ AnalyticsSvc│          │  │
│  │  └─────┬──────┘ └─────┬──────┘ └─────┬──────┘           │  │
│  │        │              │              │                   │  │
│  │  ┌─────┴──────┐ ┌─────┴──────┐ ┌─────┴──────┐           │  │
│  │  │ DomainSvc  │ │ QRSvc      │ │ BillingSvc │           │  │
│  │  └────────────┘ └────────────┘ └────────────┘           │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              │                                 │
│  ┌───────────────────────────┴──────────────────────────────┐  │
│  │                   Repository Layer                       │  │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐           │  │
│  │  │ UserRepo   │ │ LinkRepo   │ │ ClickRepo  │           │  │
│  │  └─────┬──────┘ └─────┬──────┘ └─────┬──────┘           │  │
│  └────────┼──────────────┼──────────────┼───────────────────┘  │
│           │              │              │                      │
│           ▼              ▼              ▼                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │  PostgreSQL │  │    Redis    │  │  ClickHouse │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

### Redirect Service

Optimized for sub-millisecond latency:

```
┌────────────────────────────────────────────────────────────────┐
│                      REDIRECT SERVICE                          │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  Request: GET /:code                                           │
│           │                                                    │
│           ▼                                                    │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ 1. Check Redis Cache                                    │   │
│  │    Key: link:{code}                                     │   │
│  │    TTL: 5 minutes                                       │   │
│  └────────────────────┬────────────────────────────────────┘   │
│                       │                                        │
│           ┌───────────┴───────────┐                            │
│           │ Cache Hit?            │                            │
│           └───────────┬───────────┘                            │
│                       │                                        │
│         ┌─────────────┼─────────────┐                          │
│         │ Yes         │         No  │                          │
│         ▼             │             ▼                          │
│  ┌──────────────┐     │     ┌──────────────┐                   │
│  │ Use Cached   │     │     │ Query        │                   │
│  │ Link Data    │     │     │ PostgreSQL   │                   │
│  └──────┬───────┘     │     └──────┬───────┘                   │
│         │             │            │                           │
│         │             │     ┌──────┴───────┐                   │
│         │             │     │ Cache Result │                   │
│         │             │     └──────┬───────┘                   │
│         │             │            │                           │
│         └─────────────┼────────────┘                           │
│                       │                                        │
│                       ▼                                        │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ 2. Evaluate Redirect Rules                              │   │
│  │    - Device targeting                                   │   │
│  │    - Geo targeting                                      │   │
│  │    - Time-based rules                                   │   │
│  │    - A/B testing                                        │   │
│  └────────────────────┬────────────────────────────────────┘   │
│                       │                                        │
│                       ▼                                        │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ 3. Async: Track Click                                   │   │
│  │    - Push to Redis queue                                │   │
│  │    - Worker processes to ClickHouse                     │   │
│  └────────────────────┬────────────────────────────────────┘   │
│                       │                                        │
│                       ▼                                        │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ 4. Return 301/302 Redirect                              │   │
│  │    Location: {destination_url}                          │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

### Worker Service

Processes background jobs using Asynq:

```
┌────────────────────────────────────────────────────────────────┐
│                       WORKER SERVICE                           │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    ASYNQ SERVER                         │   │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐          │   │
│  │  │ Concurrency│ │   Queues   │ │  Handlers  │          │   │
│  │  │    Pool    │ │  Priority  │ │  Registry  │          │   │
│  │  └────────────┘ └────────────┘ └────────────┘          │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                 │
│         ┌────────────────────┼────────────────────┐            │
│         │                    │                    │            │
│         ▼                    ▼                    ▼            │
│  ┌────────────┐       ┌────────────┐       ┌────────────┐      │
│  │   Click    │       │   Email    │       │  Webhook   │      │
│  │ Processor  │       │   Sender   │       │  Delivery  │      │
│  └─────┬──────┘       └─────┬──────┘       └─────┬──────┘      │
│        │                    │                    │             │
│        ▼                    ▼                    ▼             │
│  ┌────────────┐       ┌────────────┐       ┌────────────┐      │
│  │ClickHouse │       │   Resend   │       │  HTTP POST │      │
│  │   Insert   │       │    API     │       │  + Retry   │      │
│  └────────────┘       └────────────┘       └────────────┘      │
│                                                                │
│  Job Types:                                                    │
│  ├── click:process      - Process click events                │
│  ├── email:send         - Send transactional emails           │
│  ├── webhook:deliver    - Deliver webhook payloads            │
│  ├── qr:generate        - Generate QR code images             │
│  ├── analytics:rollup   - Aggregate analytics data            │
│  ├── link:preview       - Fetch link metadata                 │
│  └── export:generate    - Generate data exports               │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

---

## Data Flow

### Link Creation Flow

```
┌────────────────────────────────────────────────────────────────┐
│                    LINK CREATION FLOW                          │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  Client                    API                    Database     │
│    │                        │                        │         │
│    │  POST /api/v1/links    │                        │         │
│    │──────────────────────▶│                        │         │
│    │                        │                        │         │
│    │                        │ 1. Validate request    │         │
│    │                        │ 2. Check permissions   │         │
│    │                        │ 3. Generate short code │         │
│    │                        │                        │         │
│    │                        │ INSERT INTO links      │         │
│    │                        │───────────────────────▶│         │
│    │                        │                        │         │
│    │                        │◀───────────────────────│         │
│    │                        │                        │         │
│    │                        │ 4. Cache link          │         │
│    │                        │──────▶ Redis           │         │
│    │                        │                        │         │
│    │                        │ 5. Enqueue jobs        │         │
│    │                        │  - link:preview        │         │
│    │                        │  - webhook:deliver     │         │
│    │                        │                        │         │
│    │  201 Created + Link    │                        │         │
│    │◀──────────────────────│                        │         │
│    │                        │                        │         │
└────────────────────────────────────────────────────────────────┘
```

### Click Tracking Flow

```
┌────────────────────────────────────────────────────────────────┐
│                    CLICK TRACKING FLOW                         │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  Browser        Redirect       Redis        Worker   ClickHouse│
│    │               │             │            │           │    │
│    │ GET /abc123   │             │            │           │    │
│    │──────────────▶│             │            │           │    │
│    │               │             │            │           │    │
│    │               │ GET link:abc│            │           │    │
│    │               │────────────▶│            │           │    │
│    │               │◀────────────│            │           │    │
│    │               │             │            │           │    │
│    │               │ LPUSH clicks│            │           │    │
│    │               │────────────▶│            │           │    │
│    │               │             │            │           │    │
│    │ 301 Redirect  │             │            │           │    │
│    │◀──────────────│             │            │           │    │
│    │               │             │            │           │    │
│    │               │             │ BRPOP      │           │    │
│    │               │             │◀───────────│           │    │
│    │               │             │            │           │    │
│    │               │             │            │ GeoIP     │    │
│    │               │             │            │ Device    │    │
│    │               │             │            │ Enrich    │    │
│    │               │             │            │           │    │
│    │               │             │            │ INSERT    │    │
│    │               │             │            │──────────▶│    │
│    │               │             │            │           │    │
└────────────────────────────────────────────────────────────────┘
```

### Analytics Query Flow

```
┌────────────────────────────────────────────────────────────────┐
│                   ANALYTICS QUERY FLOW                         │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  Dashboard          API              Cache         ClickHouse  │
│    │                 │                 │               │       │
│    │ GET /analytics  │                 │               │       │
│    │────────────────▶│                 │               │       │
│    │                 │                 │               │       │
│    │                 │ Check cache     │               │       │
│    │                 │────────────────▶│               │       │
│    │                 │◀────────────────│               │       │
│    │                 │                 │               │       │
│    │                 │ [Cache Miss]    │               │       │
│    │                 │                 │               │       │
│    │                 │ SELECT ... FROM clicks          │       │
│    │                 │────────────────────────────────▶│       │
│    │                 │◀────────────────────────────────│       │
│    │                 │                 │               │       │
│    │                 │ Cache result    │               │       │
│    │                 │ TTL: 1 min      │               │       │
│    │                 │────────────────▶│               │       │
│    │                 │                 │               │       │
│    │ Analytics Data  │                 │               │       │
│    │◀────────────────│                 │               │       │
│    │                 │                 │               │       │
└────────────────────────────────────────────────────────────────┘
```

---

## Caching Strategy

### Multi-Layer Cache

```
┌─────────────────────────────────────────────────────────────┐
│                    CACHING LAYERS                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Layer 1: Cloudflare Edge (30s - 5min)                     │
│  ├── Static assets (JS, CSS, images)                       │
│  ├── Public pages (pricing, docs)                          │
│  └── Bio page content                                       │
│                                                             │
│  Layer 2: Redis (5min - 24h)                               │
│  ├── Link data (key: link:{code})                          │
│  ├── User sessions (key: session:{token})                  │
│  ├── Rate limits (key: rate:{ip}:{endpoint})               │
│  ├── Analytics cache (key: analytics:{link_id}:{period})   │
│  └── Domain verification (key: domain:{domain})            │
│                                                             │
│  Layer 3: PostgreSQL                                        │
│  └── Source of truth for all data                          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Redis Key Patterns

| Pattern | TTL | Purpose |
|---------|-----|---------|
| `link:{code}` | 5 min | Cached link data |
| `link:rules:{code}` | 5 min | Redirect rules |
| `session:{token}` | 24 hours | User sessions |
| `rate:{ip}:{endpoint}` | 1 min | Rate limiting |
| `analytics:{id}:{period}` | 1 min | Analytics cache |
| `domain:{domain}` | 1 hour | Domain verification status |
| `user:{id}` | 15 min | User profile cache |

### Cache Invalidation

```go
// Service layer handles cache invalidation
func (s *linkService) Update(ctx context.Context, id string, input UpdateLinkInput) (*models.Link, error) {
    link, err := s.repo.Update(ctx, id, input)
    if err != nil {
        return nil, err
    }

    // Invalidate cache
    s.cache.Delete(ctx, fmt.Sprintf("link:%s", link.ShortCode))
    s.cache.Delete(ctx, fmt.Sprintf("link:rules:%s", link.ShortCode))

    return link, nil
}
```

---

## Security Architecture

### Authentication Flow

```
┌─────────────────────────────────────────────────────────────┐
│                  AUTHENTICATION FLOW                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Login Request                                           │
│     POST /api/v1/auth/login                                │
│     { email, password }                                     │
│                                                             │
│  2. Verify Credentials                                      │
│     - Hash password with Argon2id                          │
│     - Compare with stored hash                             │
│                                                             │
│  3. Generate Tokens                                         │
│     - Access token (PASETO, 15 min)                        │
│     - Refresh token (opaque, 7 days)                       │
│                                                             │
│  4. Store Session                                           │
│     - Redis: session:{refresh_token} -> user_id            │
│     - Set httpOnly cookie for refresh token                │
│                                                             │
│  5. Return Response                                         │
│     { access_token, expires_at }                           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Request Authentication

```
┌─────────────────────────────────────────────────────────────┐
│              AUTHENTICATED REQUEST FLOW                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Request Headers:                                           │
│  ├── Authorization: Bearer {access_token}                  │
│  └── Cookie: refresh_token={refresh_token}                 │
│                                                             │
│  Middleware Chain:                                          │
│  1. CORS                                                    │
│  2. Rate Limiting                                           │
│  3. Request Logging                                         │
│  4. Auth Middleware                                         │
│     ├── Extract token from header                          │
│     ├── Verify PASETO signature                            │
│     ├── Check expiration                                   │
│     ├── Load user from cache/DB                            │
│     └── Set user in context                                │
│  5. License Middleware (for EE features)                   │
│  6. Permission Middleware                                   │
│  7. Handler                                                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Scalability Considerations

### Horizontal Scaling

```
┌─────────────────────────────────────────────────────────────┐
│                   SCALING TOPOLOGY                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Load Balancer (Cloudflare / AWS ALB)                      │
│         │                                                   │
│         ├─────────────────┬─────────────────┐              │
│         │                 │                 │              │
│         ▼                 ▼                 ▼              │
│  ┌──────────┐      ┌──────────┐      ┌──────────┐         │
│  │ API Pod  │      │ API Pod  │      │ API Pod  │         │
│  │   #1     │      │   #2     │      │   #3     │         │
│  └──────────┘      └──────────┘      └──────────┘         │
│                                                             │
│  ┌──────────┐      ┌──────────┐      ┌──────────┐         │
│  │Redirect  │      │Redirect  │      │Redirect  │         │
│  │ Pod #1   │      │ Pod #2   │      │ Pod #3   │         │
│  └──────────┘      └──────────┘      └──────────┘         │
│                                                             │
│  ┌──────────┐      ┌──────────┐                            │
│  │Worker #1 │      │Worker #2 │  (Scale based on queue)   │
│  └──────────┘      └──────────┘                            │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ PostgreSQL Primary │ Read Replica │ Read Replica    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │     Redis Cluster (3 masters, 3 replicas)           │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │     ClickHouse Cluster (3 shards, 2 replicas)       │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Scaling Milestones

| Redirects/Day | API Pods | Redirect Pods | Workers | PostgreSQL | Redis |
|---------------|----------|---------------|---------|------------|-------|
| 10K | 1 | 1 | 1 | Single | Single |
| 100K | 2 | 2 | 2 | Single + Replica | Single |
| 1M | 3 | 4 | 3 | Primary + 2 Replicas | 3-node Cluster |
| 10M | 5 | 10 | 5 | Cluster | 6-node Cluster |
| 100M | 10 | 25 | 10 | Sharded | 9-node Cluster |

---

## Related Documentation

- [Tech Stack](TECH_STACK.md) — Technology choices
- [Database Schema](DATABASE_SCHEMA.md) — Data model
- [Go Patterns](GO_PATTERNS.md) — Go-specific patterns
- [Scaling Guide](../deployment/SCALING_GUIDE.md) — Detailed scaling
- [Performance Optimization](../reference/PERFORMANCE_OPTIMIZATION.md) — Performance tuning
