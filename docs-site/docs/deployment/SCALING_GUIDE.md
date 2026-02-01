# Linkrift Scaling Guide

**Last Updated: 2025-01-24**

This guide covers scaling strategies for Linkrift from startup to enterprise scale, including performance baselines, scaling milestones, horizontal scaling for each service, database scaling, and cost optimization.

---

## Table of Contents

1. [Performance Baselines](#performance-baselines)
2. [Scaling Milestones](#scaling-milestones)
3. [Horizontal Scaling](#horizontal-scaling)
4. [Database Scaling](#database-scaling)
5. [Redis Clustering](#redis-clustering)
6. [ClickHouse Scaling](#clickhouse-scaling)
7. [Cost Optimization](#cost-optimization)
8. [Monitoring and Alerting](#monitoring-and-alerting)

---

## Performance Baselines

### Target Performance Metrics

| Metric | Target | Critical Threshold |
|--------|--------|-------------------|
| **Redirect Latency (P50)** | < 0.5ms | < 1ms |
| **Redirect Latency (P99)** | < 1ms | < 5ms |
| **API Latency (P50)** | < 10ms | < 50ms |
| **API Latency (P99)** | < 50ms | < 200ms |
| **Error Rate** | < 0.01% | < 0.1% |
| **Availability** | 99.99% | 99.9% |
| **Cache Hit Rate** | > 99% | > 95% |

### Current Performance Benchmarks

```
┌────────────────────────────────────────────────────────────────────────┐
│                    Redirect Service Benchmark                          │
│                    (1000 concurrent connections)                       │
├────────────────────────────────────────────────────────────────────────┤
│                                                                        │
│  Requests/sec:     125,000                                             │
│  Latency (avg):    0.4ms                                              │
│  Latency (P50):    0.3ms                                              │
│  Latency (P99):    0.8ms                                              │
│  Latency (P99.9):  1.2ms                                              │
│                                                                        │
│  Memory Usage:     ~150MB                                              │
│  CPU Usage:        ~60% (2 cores)                                      │
│                                                                        │
└────────────────────────────────────────────────────────────────────────┘
```

### Benchmark Commands

```bash
#!/bin/bash
# benchmark.sh

# Install hey (HTTP load generator)
go install github.com/rakyll/hey@latest

# Redirect service benchmark
echo "=== Redirect Service Benchmark ==="
hey -n 100000 -c 1000 -m GET http://localhost:8081/abc123

# API service benchmark
echo "=== API Service Benchmark ==="
hey -n 10000 -c 100 -m GET \
    -H "Authorization: Bearer ${API_TOKEN}" \
    http://localhost:8080/v1/links

# Link creation benchmark
echo "=== Link Creation Benchmark ==="
hey -n 1000 -c 50 -m POST \
    -H "Authorization: Bearer ${API_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"url":"https://example.com","title":"Test"}' \
    http://localhost:8080/v1/links
```

### Go Performance Optimizations

```go
// internal/redirect/handler.go
package redirect

import (
    "net/http"
    "sync"
    "time"

    "github.com/redis/go-redis/v9"
)

// Performance-optimized redirect handler
type Handler struct {
    redis     *redis.Client
    cache     *sync.Map      // In-memory L1 cache
    cacheSize int
    cacheTTL  time.Duration
}

type cacheEntry struct {
    url       string
    expiresAt time.Time
}

func NewHandler(redis *redis.Client) *Handler {
    h := &Handler{
        redis:     redis,
        cache:     &sync.Map{},
        cacheSize: 100000,  // 100K entries
        cacheTTL:  time.Minute,
    }

    // Background cache cleanup
    go h.cleanupLoop()

    return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    shortCode := r.URL.Path[1:] // Remove leading /

    // L1 cache lookup (in-memory)
    if entry, ok := h.cache.Load(shortCode); ok {
        ce := entry.(*cacheEntry)
        if time.Now().Before(ce.expiresAt) {
            h.redirect(w, r, ce.url)
            return
        }
        h.cache.Delete(shortCode)
    }

    // L2 cache lookup (Redis)
    ctx := r.Context()
    url, err := h.redis.Get(ctx, "link:"+shortCode).Result()
    if err == nil {
        // Cache in L1
        h.cache.Store(shortCode, &cacheEntry{
            url:       url,
            expiresAt: time.Now().Add(h.cacheTTL),
        })
        h.redirect(w, r, url)
        return
    }

    // Not found
    http.NotFound(w, r)
}

func (h *Handler) redirect(w http.ResponseWriter, r *http.Request, url string) {
    // Use 301 for permanent redirects (cacheable by browsers)
    w.Header().Set("Location", url)
    w.Header().Set("Cache-Control", "public, max-age=3600")
    w.WriteHeader(http.StatusMovedPermanently)
}

func (h *Handler) cleanupLoop() {
    ticker := time.NewTicker(time.Minute)
    for range ticker.C {
        now := time.Now()
        h.cache.Range(func(key, value interface{}) bool {
            if entry := value.(*cacheEntry); now.After(entry.expiresAt) {
                h.cache.Delete(key)
            }
            return true
        })
    }
}
```

---

## Scaling Milestones

### Milestone Overview

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                         Scaling Milestones                                    │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  Stage 1: Startup (10K/day)                                                  │
│  ├── Single server deployment                                                │
│  ├── PostgreSQL single instance                                              │
│  ├── Redis single instance                                                   │
│  └── Monthly cost: ~$100                                                     │
│                                                                               │
│  Stage 2: Growth (100K/day)                                                  │
│  ├── 2-3 API servers                                                         │
│  ├── 2 Redirect servers                                                      │
│  ├── PostgreSQL with read replica                                            │
│  ├── Redis with replication                                                  │
│  └── Monthly cost: ~$500                                                     │
│                                                                               │
│  Stage 3: Scale (1M/day)                                                     │
│  ├── 5 API servers                                                           │
│  ├── 5 Redirect servers                                                      │
│  ├── 3 Worker servers                                                        │
│  ├── PostgreSQL HA cluster                                                   │
│  ├── Redis cluster (3 nodes)                                                 │
│  ├── ClickHouse cluster                                                      │
│  └── Monthly cost: ~$3,000                                                   │
│                                                                               │
│  Stage 4: Enterprise (10M/day)                                               │
│  ├── 15+ API servers (auto-scaling)                                          │
│  ├── 20+ Redirect servers (auto-scaling)                                     │
│  ├── 5+ Worker servers                                                       │
│  ├── PostgreSQL with Citus/sharding                                          │
│  ├── Redis cluster (6+ nodes)                                                │
│  ├── ClickHouse cluster (3+ nodes)                                           │
│  ├── Global CDN                                                              │
│  └── Monthly cost: ~$15,000                                                  │
│                                                                               │
│  Stage 5: Hyperscale (100M/day)                                              │
│  ├── 50+ API servers (multi-region)                                          │
│  ├── 100+ Redirect servers (edge deployment)                                 │
│  ├── 20+ Worker servers                                                      │
│  ├── PostgreSQL sharded cluster                                              │
│  ├── Redis cluster (20+ nodes, multi-region)                                 │
│  ├── ClickHouse cluster (10+ nodes)                                          │
│  ├── Multi-region deployment                                                 │
│  └── Monthly cost: ~$75,000+                                                 │
│                                                                               │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Stage 1: Startup (10K redirects/day)

**Architecture:**
```
┌─────────────┐     ┌─────────────┐
│   Client    │────▶│   Server    │
└─────────────┘     │  (All-in-1) │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
         ┌────────┐  ┌─────────┐  ┌─────────┐
         │Postgres│  │  Redis  │  │ClickHouse│
         └────────┘  └─────────┘  └─────────┘
```

**Infrastructure:**

| Component | Specification | Monthly Cost |
|-----------|--------------|--------------|
| Server | 2 vCPU, 4GB RAM | $40 |
| PostgreSQL | 1 vCPU, 2GB RAM | $30 |
| Redis | 1 vCPU, 1GB RAM | $15 |
| ClickHouse | 1 vCPU, 2GB RAM | $15 |
| **Total** | | **$100** |

**Configuration:**

```yaml
# docker-compose.startup.yml
version: "3.9"

services:
  app:
    image: ghcr.io/linkrift/linkrift-all:latest
    ports:
      - "80:8080"
      - "8081:8081"
    environment:
      - LINKRIFT_MODE=all  # Run all services in one process
      - LINKRIFT_DB_MAX_CONNECTIONS=25
      - LINKRIFT_REDIS_POOL_SIZE=10
    deploy:
      resources:
        limits:
          cpus: "2"
          memory: 4G

  postgres:
    image: postgres:16-alpine
    environment:
      - POSTGRES_SHARED_BUFFERS=512MB
      - POSTGRES_WORK_MEM=16MB
    deploy:
      resources:
        limits:
          cpus: "1"
          memory: 2G

  redis:
    image: redis:7-alpine
    command: redis-server --maxmemory 512mb --maxmemory-policy allkeys-lru
    deploy:
      resources:
        limits:
          cpus: "0.5"
          memory: 1G
```

### Stage 2: Growth (100K redirects/day)

**Architecture:**
```
                    ┌─────────────┐
                    │    Load     │
                    │  Balancer   │
                    └──────┬──────┘
                           │
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
      ┌─────────┐    ┌─────────┐    ┌──────────┐
      │   API   │    │   API   │    │ Redirect │
      │ Server  │    │ Server  │    │ Servers  │
      └────┬────┘    └────┬────┘    └────┬─────┘
           │              │              │
           └──────────────┼──────────────┘
                          │
              ┌───────────┼───────────┐
              ▼           ▼           ▼
         ┌────────┐  ┌─────────┐  ┌─────────┐
         │Postgres│  │  Redis  │  │ClickHouse│
         │Primary │  │ Master  │  │         │
         │+Replica│  │+Replica │  │         │
         └────────┘  └─────────┘  └─────────┘
```

**Infrastructure:**

| Component | Specification | Count | Monthly Cost |
|-----------|--------------|-------|--------------|
| API Server | 2 vCPU, 4GB RAM | 2 | $80 |
| Redirect Server | 2 vCPU, 2GB RAM | 2 | $60 |
| Worker | 2 vCPU, 4GB RAM | 1 | $40 |
| PostgreSQL Primary | 2 vCPU, 4GB RAM | 1 | $60 |
| PostgreSQL Replica | 2 vCPU, 4GB RAM | 1 | $60 |
| Redis Primary | 2 vCPU, 4GB RAM | 1 | $40 |
| Redis Replica | 2 vCPU, 4GB RAM | 1 | $40 |
| ClickHouse | 2 vCPU, 8GB RAM | 1 | $80 |
| Load Balancer | | 1 | $40 |
| **Total** | | | **$500** |

### Stage 3: Scale (1M redirects/day)

**Architecture:**
```
                         ┌─────────────┐
                         │  Cloudflare │
                         │     CDN     │
                         └──────┬──────┘
                                │
                    ┌───────────┴───────────┐
                    ▼                       ▼
              ┌─────────┐             ┌─────────┐
              │   API   │             │Redirect │
              │   LB    │             │   LB    │
              └────┬────┘             └────┬────┘
                   │                       │
         ┌─────────┼─────────┐   ┌─────────┼─────────┐
         ▼         ▼         ▼   ▼         ▼         ▼
      ┌─────┐  ┌─────┐  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐
      │ API │  │ API │  │ API │ │Redir│ │Redir│ │Redir│
      │  1  │  │  2  │  │  3  │ │  1  │ │  2  │ │  3  │
      └──┬──┘  └──┬──┘  └──┬──┘ └──┬──┘ └──┬──┘ └──┬──┘
         └────────┴────────┴──────┴───────┴───────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
         ┌────────┐   ┌─────────┐  ┌──────────┐
         │Postgres│   │  Redis  │  │ClickHouse│
         │   HA   │   │ Cluster │  │  Cluster │
         └────────┘   └─────────┘  └──────────┘
```

**Infrastructure:**

| Component | Specification | Count | Monthly Cost |
|-----------|--------------|-------|--------------|
| API Server | 4 vCPU, 8GB RAM | 5 | $500 |
| Redirect Server | 4 vCPU, 4GB RAM | 5 | $400 |
| Worker | 4 vCPU, 8GB RAM | 3 | $300 |
| PostgreSQL HA | 4 vCPU, 16GB RAM | 3 | $600 |
| Redis Cluster | 4 vCPU, 8GB RAM | 3 | $300 |
| ClickHouse Cluster | 8 vCPU, 32GB RAM | 2 | $600 |
| Load Balancer | | 2 | $100 |
| Cloudflare Pro | | 1 | $200 |
| **Total** | | | **$3,000** |

### Stage 4: Enterprise (10M redirects/day)

**Infrastructure:**

| Component | Specification | Count | Monthly Cost |
|-----------|--------------|-------|--------------|
| API Server (auto-scale) | 4 vCPU, 8GB RAM | 10-20 | $1,500 |
| Redirect Server (auto-scale) | 4 vCPU, 4GB RAM | 15-30 | $2,000 |
| Worker | 8 vCPU, 16GB RAM | 5 | $750 |
| PostgreSQL HA + Citus | 8 vCPU, 32GB RAM | 5 | $2,000 |
| Redis Cluster | 8 vCPU, 16GB RAM | 6 | $1,200 |
| ClickHouse Cluster | 16 vCPU, 64GB RAM | 3 | $2,000 |
| Load Balancer | | 4 | $400 |
| Cloudflare Business | | 1 | $250 |
| Monitoring (Datadog) | | 1 | $500 |
| **Total** | | | **$10,600-15,000** |

### Stage 5: Hyperscale (100M redirects/day)

**Multi-Region Architecture:**

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Global Traffic Manager                       │
│                        (GeoDNS / Cloudflare LB)                     │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│   US-EAST     │       │   EU-WEST     │       │   ASIA-PAC    │
│    Region     │       │    Region     │       │    Region     │
├───────────────┤       ├───────────────┤       ├───────────────┤
│ Redirect x30  │       │ Redirect x30  │       │ Redirect x20  │
│ API x15       │       │ API x15       │       │ API x10       │
│ Worker x8     │       │ Worker x6     │       │ Worker x4     │
├───────────────┤       ├───────────────┤       ├───────────────┤
│ PG Primary    │◀─────▶│ PG Replica    │◀─────▶│ PG Replica    │
│ Redis Cluster │◀─────▶│ Redis Cluster │◀─────▶│ Redis Cluster │
│ CH Cluster    │◀─────▶│ CH Cluster    │◀─────▶│ CH Cluster    │
└───────────────┘       └───────────────┘       └───────────────┘
```

**Infrastructure:**

| Component | Specification | Count | Monthly Cost |
|-----------|--------------|-------|--------------|
| Redirect (global) | 4 vCPU, 4GB RAM | 80 | $8,000 |
| API (global) | 8 vCPU, 16GB RAM | 40 | $8,000 |
| Worker (global) | 8 vCPU, 32GB RAM | 18 | $4,500 |
| PostgreSQL (sharded) | 16 vCPU, 64GB RAM | 12 | $12,000 |
| Redis (global cluster) | 8 vCPU, 32GB RAM | 18 | $7,200 |
| ClickHouse (global) | 32 vCPU, 128GB RAM | 9 | $18,000 |
| Load Balancers | | 9 | $1,800 |
| Cloudflare Enterprise | | 1 | $5,000 |
| Monitoring | | 1 | $2,000 |
| Network/Bandwidth | | - | $8,000 |
| **Total** | | | **$74,500** |

---

## Horizontal Scaling

### API Service Scaling

```yaml
# kubernetes/api-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: linkrift-api
spec:
  replicas: 5
  selector:
    matchLabels:
      app: linkrift-api
  template:
    metadata:
      labels:
        app: linkrift-api
    spec:
      containers:
        - name: api
          image: ghcr.io/linkrift/linkrift-api:latest
          resources:
            requests:
              cpu: "500m"
              memory: "512Mi"
            limits:
              cpu: "2000m"
              memory: "4Gi"
          ports:
            - containerPort: 8080
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 15
            periodSeconds: 10
          env:
            - name: LINKRIFT_DB_MAX_CONNECTIONS
              value: "20"  # Per pod
            - name: GOMAXPROCS
              value: "2"

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: linkrift-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: linkrift-api
  minReplicas: 3
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
    - type: Pods
      pods:
        metric:
          name: http_requests_per_second
        target:
          type: AverageValue
          averageValue: "1000"
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 30
      policies:
        - type: Pods
          value: 4
          periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 10
          periodSeconds: 60
```

### Redirect Service Scaling

```yaml
# kubernetes/redirect-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: linkrift-redirect
spec:
  replicas: 10
  selector:
    matchLabels:
      app: linkrift-redirect
  template:
    metadata:
      labels:
        app: linkrift-redirect
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 100
              podAffinityTerm:
                labelSelector:
                  matchLabels:
                    app: linkrift-redirect
                topologyKey: kubernetes.io/hostname
      containers:
        - name: redirect
          image: ghcr.io/linkrift/linkrift-redirect:latest
          resources:
            requests:
              cpu: "1000m"
              memory: "256Mi"
            limits:
              cpu: "2000m"
              memory: "1Gi"
          ports:
            - containerPort: 8081
          readinessProbe:
            httpGet:
              path: /health
              port: 8081
            initialDelaySeconds: 2
            periodSeconds: 2
            timeoutSeconds: 1
          livenessProbe:
            httpGet:
              path: /health
              port: 8081
            initialDelaySeconds: 5
            periodSeconds: 5
          env:
            - name: LINKRIFT_REDIS_POOL_SIZE
              value: "100"
            - name: GOMAXPROCS
              value: "2"

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: linkrift-redirect-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: linkrift-redirect
  minReplicas: 5
  maxReplicas: 100
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 60
    - type: Pods
      pods:
        metric:
          name: redirect_requests_per_second
        target:
          type: AverageValue
          averageValue: "10000"
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 0  # Scale up immediately
      policies:
        - type: Percent
          value: 100
          periodSeconds: 15
        - type: Pods
          value: 10
          periodSeconds: 15
      selectPolicy: Max
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 10
          periodSeconds: 60
```

### Worker Scaling

```yaml
# kubernetes/worker-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: linkrift-worker
spec:
  replicas: 3
  selector:
    matchLabels:
      app: linkrift-worker
  template:
    metadata:
      labels:
        app: linkrift-worker
    spec:
      containers:
        - name: worker
          image: ghcr.io/linkrift/linkrift-worker:latest
          resources:
            requests:
              cpu: "500m"
              memory: "1Gi"
            limits:
              cpu: "4000m"
              memory: "8Gi"
          env:
            - name: LINKRIFT_WORKER_CONCURRENCY
              value: "20"
            - name: LINKRIFT_WORKER_BATCH_SIZE
              value: "5000"

---
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: linkrift-worker-scaledobject
spec:
  scaleTargetRef:
    name: linkrift-worker
  minReplicaCount: 2
  maxReplicaCount: 20
  triggers:
    - type: redis
      metadata:
        address: redis.linkrift.internal:6379
        listName: linkrift:analytics:queue
        listLength: "10000"
```

---

## Database Scaling

### PostgreSQL Connection Pooling (PgBouncer)

```ini
# pgbouncer.ini

[databases]
linkrift = host=postgres-primary.linkrift.internal port=5432 dbname=linkrift
linkrift_readonly = host=postgres-replica.linkrift.internal port=5432 dbname=linkrift

[pgbouncer]
listen_addr = 0.0.0.0
listen_port = 6432
auth_type = scram-sha-256
auth_file = /etc/pgbouncer/userlist.txt

# Connection pooling
pool_mode = transaction
max_client_conn = 10000
default_pool_size = 50
min_pool_size = 10
reserve_pool_size = 10
reserve_pool_timeout = 5

# Performance
tcp_keepalive = 1
tcp_keepcnt = 3
tcp_keepidle = 60
tcp_keepintvl = 10

# Logging
log_connections = 0
log_disconnections = 0
log_pooler_errors = 1
stats_period = 60

# Memory
pkt_buf = 4096
max_prepared_statements = 0
```

### PostgreSQL Read Replicas

```sql
-- On primary: Configure streaming replication
-- postgresql.conf
wal_level = replica
max_wal_senders = 10
max_replication_slots = 10
hot_standby = on

-- pg_hba.conf (add replication user)
host replication replicator replica-ip/32 scram-sha-256

-- Create replication slot
SELECT pg_create_physical_replication_slot('replica_1');
```

```bash
# On replica: Set up streaming replication
pg_basebackup -h primary-host -D /var/lib/postgresql/data -U replicator -P -R -X stream
```

### PostgreSQL Sharding with Citus

```sql
-- Install Citus extension
CREATE EXTENSION citus;

-- Add worker nodes
SELECT citus_add_node('worker1.linkrift.internal', 5432);
SELECT citus_add_node('worker2.linkrift.internal', 5432);
SELECT citus_add_node('worker3.linkrift.internal', 5432);

-- Distribute links table by user_id
SELECT create_distributed_table('links', 'user_id');

-- Distribute clicks table by link_id
SELECT create_distributed_table('clicks', 'link_id');

-- Reference tables (replicated to all nodes)
SELECT create_reference_table('users');
SELECT create_reference_table('workspaces');

-- Check distribution
SELECT * FROM citus_tables;
SELECT * FROM citus_shards;
```

### Connection Management in Go

```go
// internal/database/pool.go
package database

import (
    "context"
    "database/sql"
    "time"

    _ "github.com/jackc/pgx/v5/stdlib"
)

type PoolConfig struct {
    PrimaryDSN     string
    ReplicaDSN     string
    MaxOpenConns   int
    MaxIdleConns   int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}

type Pool struct {
    primary *sql.DB
    replica *sql.DB
}

func NewPool(cfg PoolConfig) (*Pool, error) {
    primary, err := openDB(cfg.PrimaryDSN, cfg)
    if err != nil {
        return nil, err
    }

    replica, err := openDB(cfg.ReplicaDSN, cfg)
    if err != nil {
        return nil, err
    }

    return &Pool{
        primary: primary,
        replica: replica,
    }, nil
}

func openDB(dsn string, cfg PoolConfig) (*sql.DB, error) {
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, err
    }

    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        return nil, err
    }

    return db, nil
}

// Primary returns the primary (read-write) connection
func (p *Pool) Primary() *sql.DB {
    return p.primary
}

// Replica returns a replica (read-only) connection
func (p *Pool) Replica() *sql.DB {
    return p.replica
}

// ReadWrite executes a function with the primary connection
func (p *Pool) ReadWrite(ctx context.Context, fn func(*sql.Conn) error) error {
    conn, err := p.primary.Conn(ctx)
    if err != nil {
        return err
    }
    defer conn.Close()
    return fn(conn)
}

// ReadOnly executes a function with a replica connection
func (p *Pool) ReadOnly(ctx context.Context, fn func(*sql.Conn) error) error {
    conn, err := p.replica.Conn(ctx)
    if err != nil {
        return err
    }
    defer conn.Close()
    return fn(conn)
}
```

---

## Redis Clustering

### Redis Cluster Configuration

```bash
# Create cluster with 6 nodes (3 masters, 3 replicas)
redis-cli --cluster create \
    redis-1:6379 redis-2:6379 redis-3:6379 \
    redis-4:6379 redis-5:6379 redis-6:6379 \
    --cluster-replicas 1 \
    --cluster-yes
```

### Redis Configuration

```conf
# redis.conf (for cluster mode)

# Cluster configuration
cluster-enabled yes
cluster-config-file nodes.conf
cluster-node-timeout 5000
cluster-announce-ip <node-ip>
cluster-announce-port 6379
cluster-announce-bus-port 16379

# Memory management
maxmemory 8gb
maxmemory-policy allkeys-lru
maxmemory-samples 10

# Persistence (for cache, can disable)
save ""
appendonly no

# Performance
tcp-backlog 65535
tcp-keepalive 300
timeout 0

# Threads
io-threads 4
io-threads-do-reads yes
```

### Go Redis Cluster Client

```go
// internal/cache/redis_cluster.go
package cache

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisCluster struct {
    client *redis.ClusterClient
}

func NewRedisCluster(addrs []string, password string) *RedisCluster {
    client := redis.NewClusterClient(&redis.ClusterOptions{
        Addrs:    addrs,
        Password: password,

        // Pool settings
        PoolSize:        100,
        MinIdleConns:    20,
        PoolTimeout:     time.Second * 4,
        ConnMaxIdleTime: time.Minute * 5,

        // Read settings
        ReadOnly:       true,
        RouteRandomly:  true,

        // Timeouts
        DialTimeout:  time.Second * 5,
        ReadTimeout:  time.Second * 3,
        WriteTimeout: time.Second * 3,

        // Retry
        MaxRetries:      3,
        MinRetryBackoff: time.Millisecond * 8,
        MaxRetryBackoff: time.Millisecond * 512,
    })

    return &RedisCluster{client: client}
}

// GetLink retrieves a link URL from cache
func (r *RedisCluster) GetLink(ctx context.Context, shortCode string) (string, error) {
    return r.client.Get(ctx, "link:"+shortCode).Result()
}

// SetLink caches a link URL
func (r *RedisCluster) SetLink(ctx context.Context, shortCode, url string, ttl time.Duration) error {
    return r.client.Set(ctx, "link:"+shortCode, url, ttl).Err()
}

// IncrementClicks atomically increments click count
func (r *RedisCluster) IncrementClicks(ctx context.Context, shortCode string) (int64, error) {
    return r.client.Incr(ctx, "clicks:"+shortCode).Result()
}

// Pipeline executes multiple commands in a pipeline
func (r *RedisCluster) Pipeline(ctx context.Context, fn func(redis.Pipeliner) error) error {
    _, err := r.client.Pipelined(ctx, fn)
    return err
}

// Health checks cluster health
func (r *RedisCluster) Health(ctx context.Context) error {
    return r.client.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
        return shard.Ping(ctx).Err()
    })
}
```

### Multi-Region Redis

```go
// internal/cache/multi_region.go
package cache

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

type MultiRegionCache struct {
    local  *redis.ClusterClient
    global *redis.ClusterClient
}

func NewMultiRegionCache(localAddrs, globalAddrs []string) *MultiRegionCache {
    return &MultiRegionCache{
        local:  createCluster(localAddrs),
        global: createCluster(globalAddrs),
    }
}

// Get tries local first, then global
func (m *MultiRegionCache) Get(ctx context.Context, key string) (string, error) {
    // Try local cache first
    val, err := m.local.Get(ctx, key).Result()
    if err == nil {
        return val, nil
    }

    // Fall back to global cache
    val, err = m.global.Get(ctx, key).Result()
    if err == nil {
        // Populate local cache
        go m.local.Set(context.Background(), key, val, time.Hour)
        return val, nil
    }

    return "", err
}

// Set writes to both local and global
func (m *MultiRegionCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
    // Write to local
    if err := m.local.Set(ctx, key, value, ttl).Err(); err != nil {
        return err
    }

    // Async write to global
    go m.global.Set(context.Background(), key, value, ttl)

    return nil
}
```

---

## ClickHouse Scaling

### ClickHouse Cluster Configuration

```xml
<!-- /etc/clickhouse-server/config.d/cluster.xml -->
<clickhouse>
    <remote_servers>
        <linkrift_cluster>
            <shard>
                <replica>
                    <host>clickhouse-1</host>
                    <port>9000</port>
                </replica>
                <replica>
                    <host>clickhouse-2</host>
                    <port>9000</port>
                </replica>
            </shard>
            <shard>
                <replica>
                    <host>clickhouse-3</host>
                    <port>9000</port>
                </replica>
                <replica>
                    <host>clickhouse-4</host>
                    <port>9000</port>
                </replica>
            </shard>
            <shard>
                <replica>
                    <host>clickhouse-5</host>
                    <port>9000</port>
                </replica>
                <replica>
                    <host>clickhouse-6</host>
                    <port>9000</port>
                </replica>
            </shard>
        </linkrift_cluster>
    </remote_servers>

    <zookeeper>
        <node>
            <host>zookeeper-1</host>
            <port>2181</port>
        </node>
        <node>
            <host>zookeeper-2</host>
            <port>2181</port>
        </node>
        <node>
            <host>zookeeper-3</host>
            <port>2181</port>
        </node>
    </zookeeper>

    <macros>
        <cluster>linkrift_cluster</cluster>
        <shard>01</shard>  <!-- Set per node -->
        <replica>replica_01</replica>  <!-- Set per node -->
    </macros>
</clickhouse>
```

### Distributed Tables

```sql
-- Create local table on each node
CREATE TABLE linkrift_analytics.clicks_local ON CLUSTER linkrift_cluster
(
    click_id UUID,
    link_id UUID,
    user_id UUID,
    short_code String,
    clicked_at DateTime64(3),
    ip_address IPv4,
    user_agent String,
    referer String,
    country_code LowCardinality(String),
    region String,
    city String,
    device_type LowCardinality(String),
    browser LowCardinality(String),
    os LowCardinality(String),
    is_bot UInt8
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/clicks', '{replica}')
PARTITION BY toYYYYMM(clicked_at)
ORDER BY (link_id, clicked_at)
TTL clicked_at + INTERVAL 2 YEAR
SETTINGS index_granularity = 8192;

-- Create distributed table
CREATE TABLE linkrift_analytics.clicks ON CLUSTER linkrift_cluster
(
    click_id UUID,
    link_id UUID,
    user_id UUID,
    short_code String,
    clicked_at DateTime64(3),
    ip_address IPv4,
    user_agent String,
    referer String,
    country_code LowCardinality(String),
    region String,
    city String,
    device_type LowCardinality(String),
    browser LowCardinality(String),
    os LowCardinality(String),
    is_bot UInt8
)
ENGINE = Distributed(linkrift_cluster, linkrift_analytics, clicks_local, rand());

-- Create materialized view for aggregations
CREATE MATERIALIZED VIEW linkrift_analytics.clicks_hourly_mv ON CLUSTER linkrift_cluster
TO linkrift_analytics.clicks_hourly
AS SELECT
    link_id,
    toStartOfHour(clicked_at) AS hour,
    count() AS click_count,
    uniqExact(ip_address) AS unique_visitors,
    countIf(is_bot = 0) AS human_clicks
FROM linkrift_analytics.clicks_local
GROUP BY link_id, hour;
```

### ClickHouse Performance Tuning

```xml
<!-- /etc/clickhouse-server/config.d/performance.xml -->
<clickhouse>
    <!-- Memory limits -->
    <max_memory_usage>32000000000</max_memory_usage> <!-- 32GB -->
    <max_memory_usage_for_all_queries>48000000000</max_memory_usage_for_all_queries>

    <!-- Merge settings -->
    <max_bytes_to_merge_at_max_space_in_pool>161061273600</max_bytes_to_merge_at_max_space_in_pool>
    <max_bytes_to_merge_at_min_space_in_pool>1048576</max_bytes_to_merge_at_min_space_in_pool>

    <!-- Background tasks -->
    <background_pool_size>16</background_pool_size>
    <background_schedule_pool_size>16</background_schedule_pool_size>

    <!-- Query settings -->
    <max_threads>16</max_threads>
    <max_insert_threads>8</max_insert_threads>

    <!-- Compression -->
    <compression>
        <case>
            <method>zstd</method>
            <level>3</level>
        </case>
    </compression>
</clickhouse>
```

---

## Cost Optimization

### Resource Right-Sizing

```yaml
# Example: Optimize based on actual usage

# Before optimization (over-provisioned)
services:
  api:
    resources:
      requests:
        cpu: "2000m"
        memory: "4Gi"
      limits:
        cpu: "4000m"
        memory: "8Gi"

# After optimization (right-sized based on metrics)
services:
  api:
    resources:
      requests:
        cpu: "500m"      # Based on P95 usage
        memory: "1Gi"    # Based on actual usage + buffer
      limits:
        cpu: "2000m"     # Allow bursting
        memory: "2Gi"    # Prevent OOM
```

### Spot/Preemptible Instances

```yaml
# kubernetes/spot-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: linkrift-worker
spec:
  template:
    spec:
      nodeSelector:
        node.kubernetes.io/lifecycle: spot
      tolerations:
        - key: "kubernetes.io/lifecycle"
          operator: "Equal"
          value: "spot"
          effect: "NoSchedule"
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 1
              preference:
                matchExpressions:
                  - key: node.kubernetes.io/lifecycle
                    operator: In
                    values:
                      - spot
```

### Reserved Instances Strategy

```
┌──────────────────────────────────────────────────────────────────────┐
│                    Instance Purchase Strategy                         │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  Baseline Load (always needed):                                       │
│  └── Reserved Instances (1-3 year) = 60% savings                     │
│      ├── 3x API servers                                              │
│      ├── 5x Redirect servers                                         │
│      ├── 2x Worker servers                                           │
│      └── All database instances                                      │
│                                                                       │
│  Variable Load (peak times):                                          │
│  └── Spot Instances = 70-90% savings                                 │
│      ├── Additional redirect servers                                 │
│      └── Additional worker servers                                   │
│                                                                       │
│  Unpredictable Spikes:                                                │
│  └── On-Demand Instances (pay as you go)                             │
│      └── Auto-scaling overflow                                       │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

### Cost Monitoring

```yaml
# Example: Cost allocation tags
resources:
  labels:
    cost-center: "linkrift"
    environment: "production"
    service: "redirect"
    team: "platform"

# AWS Cost Explorer query
aws ce get-cost-and-usage \
    --time-period Start=2025-01-01,End=2025-01-31 \
    --granularity MONTHLY \
    --metrics "BlendedCost" \
    --group-by Type=TAG,Key=service
```

### Cost Optimization Checklist

| Category | Action | Potential Savings |
|----------|--------|-------------------|
| **Compute** | Right-size instances | 20-40% |
| **Compute** | Use spot instances for workers | 60-80% |
| **Compute** | Reserved instances for baseline | 30-60% |
| **Storage** | Use appropriate storage tiers | 20-50% |
| **Storage** | Implement data lifecycle policies | 30-50% |
| **Network** | Use internal endpoints | 10-20% |
| **Network** | Optimize data transfer | 20-40% |
| **Database** | Use read replicas efficiently | 10-30% |
| **Cache** | Optimize cache hit rates | 20-40% |

---

## Monitoring and Alerting

### Key Metrics Dashboard

```yaml
# grafana/dashboards/scaling.json
{
  "dashboard": {
    "title": "Linkrift Scaling Metrics",
    "panels": [
      {
        "title": "Redirect Latency P99",
        "targets": [
          {
            "expr": "histogram_quantile(0.99, sum(rate(redirect_duration_seconds_bucket[5m])) by (le))"
          }
        ],
        "alert": {
          "conditions": [
            {
              "evaluator": { "params": [0.001], "type": "gt" },
              "reducer": { "type": "avg" }
            }
          ]
        }
      },
      {
        "title": "Requests per Second",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total[1m])) by (service)"
          }
        ]
      },
      {
        "title": "Cache Hit Rate",
        "targets": [
          {
            "expr": "sum(rate(cache_hits_total[5m])) / sum(rate(cache_requests_total[5m]))"
          }
        ]
      },
      {
        "title": "Database Connections",
        "targets": [
          {
            "expr": "sum(pg_stat_activity_count) by (datname)"
          }
        ]
      },
      {
        "title": "Pod Count by Service",
        "targets": [
          {
            "expr": "sum(kube_deployment_status_replicas_available) by (deployment)"
          }
        ]
      }
    ]
  }
}
```

### Scaling Alerts

```yaml
# prometheus/alerts/scaling.yml
groups:
  - name: scaling
    rules:
      - alert: HighRedirectLatency
        expr: |
          histogram_quantile(0.99, sum(rate(redirect_duration_seconds_bucket[5m])) by (le)) > 0.001
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: Redirect P99 latency exceeds 1ms
          description: Consider scaling redirect service

      - alert: HighCPUUtilization
        expr: |
          avg(rate(container_cpu_usage_seconds_total{container!=""}[5m])) by (pod) > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: Pod {{ $labels.pod }} CPU > 80%
          description: Consider horizontal scaling

      - alert: LowCacheHitRate
        expr: |
          sum(rate(cache_hits_total[5m])) / sum(rate(cache_requests_total[5m])) < 0.95
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: Cache hit rate below 95%
          description: Check cache size and eviction policy

      - alert: DatabaseConnectionPoolExhausted
        expr: |
          pg_stat_activity_count / pg_settings_max_connections > 0.8
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: Database connections near limit
          description: Add read replicas or increase pool size

      - alert: HPAMaxedOut
        expr: |
          kube_horizontalpodautoscaler_status_current_replicas == kube_horizontalpodautoscaler_spec_max_replicas
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: HPA at maximum replicas
          description: Consider increasing max replicas
```

### Capacity Planning Queries

```sql
-- ClickHouse: Predict growth
SELECT
    toStartOfDay(clicked_at) AS day,
    count() AS clicks,
    clicks - lagInFrame(clicks) OVER (ORDER BY day) AS growth,
    round(growth / lagInFrame(clicks) OVER (ORDER BY day) * 100, 2) AS growth_pct
FROM linkrift_analytics.clicks
WHERE clicked_at >= now() - INTERVAL 30 DAY
GROUP BY day
ORDER BY day;

-- Estimate storage needs
SELECT
    formatReadableSize(sum(bytes)) AS total_size,
    formatReadableSize(sum(bytes) / count(DISTINCT toYYYYMM(clicked_at))) AS monthly_avg,
    formatReadableSize(sum(bytes) / count(DISTINCT toYYYYMM(clicked_at)) * 12) AS yearly_estimate
FROM system.parts
WHERE database = 'linkrift_analytics';
```

---

## Quick Reference

### Scaling Decision Matrix

| Symptom | Metric | Action |
|---------|--------|--------|
| High redirect latency | P99 > 1ms | Scale redirect pods |
| Cache misses | Hit rate < 95% | Increase Redis memory |
| DB connections exhausted | Usage > 80% | Add read replicas |
| High API latency | P99 > 100ms | Scale API pods |
| Worker queue backlog | Queue > 10K | Scale workers |
| Memory pressure | Usage > 85% | Increase limits or scale |

### Scaling Commands

```bash
# Kubernetes
kubectl scale deployment linkrift-redirect --replicas=10
kubectl autoscale deployment linkrift-api --min=3 --max=20 --cpu-percent=70

# Docker Swarm
docker service scale linkrift_redirect=10

# AWS ECS
aws ecs update-service --cluster linkrift --service redirect --desired-count 10
```

---

*This scaling guide is maintained by the Linkrift Platform Team. For questions or updates, contact platform@linkrift.io*
