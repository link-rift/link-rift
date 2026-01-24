# Linkrift Docker Setup Guide

**Last Updated: 2025-01-24**

This guide covers Docker configuration for Linkrift, including Dockerfiles for each service, multi-stage builds, Docker Compose configurations, volume management, and network setup.

---

## Table of Contents

1. [Overview](#overview)
2. [Dockerfiles](#dockerfiles)
3. [Multi-Stage Go Builds](#multi-stage-go-builds)
4. [Docker Compose Development](#docker-compose-development)
5. [Docker Compose Production](#docker-compose-production)
6. [Volume Management](#volume-management)
7. [Network Configuration](#network-configuration)
8. [Best Practices](#best-practices)

---

## Overview

### Service Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Docker Network: linkrift                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────┐  ┌──────────┐  ┌────────┐  ┌─────┐               │
│  │   API   │  │ Redirect │  │ Worker │  │ Web │               │
│  │  :8080  │  │  :8081   │  │        │  │ :80 │               │
│  └────┬────┘  └────┬─────┘  └───┬────┘  └──┬──┘               │
│       │            │            │          │                   │
│       └────────────┴────────────┴──────────┘                   │
│                         │                                       │
│  ┌──────────────────────┴───────────────────────┐              │
│  │                                               │              │
│  │  ┌──────────┐  ┌─────────┐  ┌────────────┐  │              │
│  │  │ Postgres │  │  Redis  │  │ ClickHouse │  │              │
│  │  │  :5432   │  │  :6379  │  │   :8123    │  │              │
│  │  └──────────┘  └─────────┘  └────────────┘  │              │
│  │                                               │              │
│  │            Infrastructure Services            │              │
│  └───────────────────────────────────────────────┘              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
linkrift/
├── docker/
│   ├── api/
│   │   └── Dockerfile
│   ├── redirect/
│   │   └── Dockerfile
│   ├── worker/
│   │   └── Dockerfile
│   ├── web/
│   │   ├── Dockerfile
│   │   └── nginx.conf
│   └── migrations/
│       └── Dockerfile
├── docker-compose.yml
├── docker-compose.dev.yml
├── docker-compose.prod.yml
└── .dockerignore
```

---

## Dockerfiles

### API Service Dockerfile

```dockerfile
# docker/api/Dockerfile

# ============================================
# Stage 1: Build
# ============================================
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w \
        -X main.Version=${VERSION} \
        -X main.Commit=${COMMIT} \
        -X main.BuildTime=${BUILD_TIME}" \
    -o /build/linkrift-api \
    ./cmd/api

# ============================================
# Stage 2: Runtime
# ============================================
FROM alpine:3.19 AS runtime

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata

# Create non-root user
RUN addgroup -g 1000 linkrift && \
    adduser -u 1000 -G linkrift -s /bin/sh -D linkrift

# Create directories
RUN mkdir -p /app /var/log/linkrift && \
    chown -R linkrift:linkrift /app /var/log/linkrift

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/linkrift-api /app/linkrift-api

# Copy migrations (if bundled)
COPY --from=builder /build/migrations /app/migrations

# Set ownership
RUN chown linkrift:linkrift /app/linkrift-api

# Switch to non-root user
USER linkrift

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run
ENTRYPOINT ["/app/linkrift-api"]
```

### Redirect Service Dockerfile

```dockerfile
# docker/redirect/Dockerfile

# ============================================
# Stage 1: Build
# ============================================
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

# Build with maximum optimization for performance
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w \
        -X main.Version=${VERSION} \
        -X main.Commit=${COMMIT} \
        -X main.BuildTime=${BUILD_TIME}" \
    -o /build/linkrift-redirect \
    ./cmd/redirect

# ============================================
# Stage 2: Runtime (minimal image for performance)
# ============================================
FROM scratch

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary
COPY --from=builder /build/linkrift-redirect /linkrift-redirect

# Expose port
EXPOSE 8081

# Run
ENTRYPOINT ["/linkrift-redirect"]
```

### Worker Service Dockerfile

```dockerfile
# docker/worker/Dockerfile

# ============================================
# Stage 1: Build
# ============================================
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w \
        -X main.Version=${VERSION} \
        -X main.Commit=${COMMIT} \
        -X main.BuildTime=${BUILD_TIME}" \
    -o /build/linkrift-worker \
    ./cmd/worker

# ============================================
# Stage 2: Runtime
# ============================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -g 1000 linkrift && \
    adduser -u 1000 -G linkrift -s /bin/sh -D linkrift

RUN mkdir -p /app /var/log/linkrift && \
    chown -R linkrift:linkrift /app /var/log/linkrift

WORKDIR /app

COPY --from=builder /build/linkrift-worker /app/linkrift-worker

USER linkrift

# Health check via file (worker has no HTTP endpoint)
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD test -f /tmp/worker-healthy || exit 1

ENTRYPOINT ["/app/linkrift-worker"]
```

### Web Frontend Dockerfile

```dockerfile
# docker/web/Dockerfile

# ============================================
# Stage 1: Dependencies
# ============================================
FROM node:20-alpine AS deps

WORKDIR /app

# Copy package files
COPY web/package.json web/pnpm-lock.yaml ./

# Install pnpm and dependencies
RUN npm install -g pnpm && \
    pnpm install --frozen-lockfile

# ============================================
# Stage 2: Build
# ============================================
FROM node:20-alpine AS builder

WORKDIR /app

# Copy dependencies from deps stage
COPY --from=deps /app/node_modules ./node_modules

# Copy source code
COPY web/ .

# Build arguments
ARG VITE_API_URL=https://api.linkrift.io
ARG VITE_APP_VERSION=dev

# Set environment variables for build
ENV VITE_API_URL=${VITE_API_URL}
ENV VITE_APP_VERSION=${VITE_APP_VERSION}

# Install pnpm and build
RUN npm install -g pnpm && \
    pnpm build

# ============================================
# Stage 3: Production
# ============================================
FROM nginx:alpine AS production

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom nginx config
COPY docker/web/nginx.conf /etc/nginx/nginx.conf

# Copy built assets
COPY --from=builder /app/dist /usr/share/nginx/html

# Create non-root user for nginx
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    touch /var/run/nginx.pid && \
    chown nginx:nginx /var/run/nginx.pid

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/health || exit 1

# Run as non-root
USER nginx

CMD ["nginx", "-g", "daemon off;"]
```

### Web Nginx Configuration

```nginx
# docker/web/nginx.conf

user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # Logging format
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;

    # Performance settings
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml application/json
               application/javascript application/xml+rss
               application/atom+xml image/svg+xml;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    server {
        listen 80;
        server_name _;
        root /usr/share/nginx/html;
        index index.html;

        # Health check endpoint
        location /health {
            access_log off;
            return 200 "OK\n";
            add_header Content-Type text/plain;
        }

        # Static assets with long cache
        location /assets/ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }

        # SPA routing
        location / {
            try_files $uri $uri/ /index.html;

            # Don't cache index.html
            location = /index.html {
                add_header Cache-Control "no-cache, no-store, must-revalidate";
                add_header Pragma "no-cache";
                add_header Expires "0";
            }
        }

        # API proxy (for development)
        location /api/ {
            proxy_pass http://api:8080/;
            proxy_http_version 1.1;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

### Migrations Dockerfile

```dockerfile
# docker/migrations/Dockerfile

FROM golang:1.22-alpine AS builder

# Install migrate tool
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

FROM alpine:3.19

RUN apk add --no-cache ca-certificates postgresql-client

COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

COPY migrations /migrations

# Migration script
COPY docker/migrations/migrate.sh /migrate.sh
RUN chmod +x /migrate.sh

ENTRYPOINT ["/migrate.sh"]
```

### Migration Script

```bash
#!/bin/sh
# docker/migrations/migrate.sh

set -e

# Wait for database to be ready
echo "Waiting for database..."
until pg_isready -h "${DB_HOST:-postgres}" -p "${DB_PORT:-5432}" -U "${DB_USER:-linkrift}"; do
    echo "Database not ready, waiting..."
    sleep 2
done

echo "Database is ready!"

# Build connection string
DB_URL="postgres://${DB_USER:-linkrift}:${DB_PASSWORD}@${DB_HOST:-postgres}:${DB_PORT:-5432}/${DB_NAME:-linkrift}?sslmode=${DB_SSL_MODE:-disable}"

# Run migrations
case "${MIGRATE_COMMAND:-up}" in
    up)
        echo "Running migrations up..."
        migrate -path /migrations -database "${DB_URL}" up
        ;;
    down)
        echo "Running migrations down..."
        migrate -path /migrations -database "${DB_URL}" down "${MIGRATE_STEPS:-1}"
        ;;
    version)
        echo "Current migration version:"
        migrate -path /migrations -database "${DB_URL}" version
        ;;
    force)
        echo "Forcing migration version to ${MIGRATE_VERSION}..."
        migrate -path /migrations -database "${DB_URL}" force "${MIGRATE_VERSION}"
        ;;
    *)
        echo "Unknown command: ${MIGRATE_COMMAND}"
        exit 1
        ;;
esac

echo "Migration complete!"
```

---

## Multi-Stage Go Builds

### Optimized Multi-Stage Build

```dockerfile
# Dockerfile.optimized

# ============================================
# Stage 1: Module Cache
# ============================================
FROM golang:1.22-alpine AS modules

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

# ============================================
# Stage 2: Build
# ============================================
FROM golang:1.22-alpine AS builder

# Install UPX for binary compression (optional)
RUN apk add --no-cache git ca-certificates upx

WORKDIR /build

# Copy cached modules
COPY --from=modules /go/pkg /go/pkg

# Copy source
COPY . .

# Build arguments
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
ARG SERVICE=api

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w \
        -X main.Version=${VERSION} \
        -X main.Commit=${COMMIT} \
        -X main.BuildTime=${BUILD_TIME}" \
    -trimpath \
    -o /build/linkrift-${SERVICE} \
    ./cmd/${SERVICE}

# Compress binary (reduces size by ~60%)
RUN upx --best --lzma /build/linkrift-${SERVICE} || true

# ============================================
# Stage 3: Security Scan (optional)
# ============================================
FROM aquasec/trivy:latest AS scanner

COPY --from=builder /build/linkrift-* /scan/
RUN trivy rootfs --no-progress --severity HIGH,CRITICAL /scan/

# ============================================
# Stage 4: Runtime
# ============================================
FROM gcr.io/distroless/static-debian12 AS runtime

# Copy binary
COPY --from=builder /build/linkrift-* /app/

# Non-root user (distroless default)
USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/linkrift-api"]
```

### Build Script for All Services

```bash
#!/bin/bash
# build-docker.sh

set -euo pipefail

# Configuration
REGISTRY="${DOCKER_REGISTRY:-ghcr.io/linkrift}"
VERSION="${VERSION:-$(git describe --tags --always --dirty)}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD)}"
BUILD_TIME="${BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

SERVICES=("api" "redirect" "worker" "web")

echo "=== Building Linkrift Docker Images ==="
echo "Registry: ${REGISTRY}"
echo "Version: ${VERSION}"
echo "Commit: ${COMMIT}"
echo ""

# Build each service
for service in "${SERVICES[@]}"; do
    echo "Building ${service}..."

    if [[ "${service}" == "web" ]]; then
        docker build \
            --file docker/web/Dockerfile \
            --build-arg VITE_APP_VERSION="${VERSION}" \
            --tag "${REGISTRY}/linkrift-${service}:${VERSION}" \
            --tag "${REGISTRY}/linkrift-${service}:latest" \
            .
    else
        docker build \
            --file docker/${service}/Dockerfile \
            --build-arg VERSION="${VERSION}" \
            --build-arg COMMIT="${COMMIT}" \
            --build-arg BUILD_TIME="${BUILD_TIME}" \
            --tag "${REGISTRY}/linkrift-${service}:${VERSION}" \
            --tag "${REGISTRY}/linkrift-${service}:latest" \
            .
    fi

    echo "${service} built successfully!"
    echo ""
done

# Build migrations
echo "Building migrations..."
docker build \
    --file docker/migrations/Dockerfile \
    --tag "${REGISTRY}/linkrift-migrations:${VERSION}" \
    --tag "${REGISTRY}/linkrift-migrations:latest" \
    .

echo ""
echo "=== All Images Built ==="
docker images | grep linkrift
```

---

## Docker Compose Development

```yaml
# docker-compose.dev.yml

version: "3.9"

services:
  # ===========================================
  # Application Services
  # ===========================================
  api:
    build:
      context: .
      dockerfile: docker/api/Dockerfile
      target: builder
    command: go run ./cmd/api
    volumes:
      - .:/build
      - go-cache:/go/pkg
    ports:
      - "8080:8080"
    environment:
      - LINKRIFT_ENV=development
      - LINKRIFT_API_PORT=8080
      - LINKRIFT_DB_HOST=postgres
      - LINKRIFT_DB_PORT=5432
      - LINKRIFT_DB_NAME=linkrift
      - LINKRIFT_DB_USER=linkrift
      - LINKRIFT_DB_PASSWORD=linkrift_dev
      - LINKRIFT_DB_SSL_MODE=disable
      - LINKRIFT_REDIS_URL=redis://redis:6379/0
      - LINKRIFT_JWT_SECRET=dev-secret-change-in-production
      - LINKRIFT_LOG_LEVEL=debug
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - linkrift

  redirect:
    build:
      context: .
      dockerfile: docker/redirect/Dockerfile
      target: builder
    command: go run ./cmd/redirect
    volumes:
      - .:/build
      - go-cache:/go/pkg
    ports:
      - "8081:8081"
    environment:
      - LINKRIFT_ENV=development
      - LINKRIFT_REDIRECT_PORT=8081
      - LINKRIFT_DB_HOST=postgres
      - LINKRIFT_DB_PORT=5432
      - LINKRIFT_DB_NAME=linkrift
      - LINKRIFT_DB_USER=linkrift
      - LINKRIFT_DB_PASSWORD=linkrift_dev
      - LINKRIFT_DB_SSL_MODE=disable
      - LINKRIFT_REDIS_URL=redis://redis:6379/0
      - LINKRIFT_LOG_LEVEL=debug
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - linkrift

  worker:
    build:
      context: .
      dockerfile: docker/worker/Dockerfile
      target: builder
    command: go run ./cmd/worker
    volumes:
      - .:/build
      - go-cache:/go/pkg
    environment:
      - LINKRIFT_ENV=development
      - LINKRIFT_DB_HOST=postgres
      - LINKRIFT_DB_PORT=5432
      - LINKRIFT_DB_NAME=linkrift
      - LINKRIFT_DB_USER=linkrift
      - LINKRIFT_DB_PASSWORD=linkrift_dev
      - LINKRIFT_DB_SSL_MODE=disable
      - LINKRIFT_REDIS_URL=redis://redis:6379/1
      - LINKRIFT_CLICKHOUSE_HOST=clickhouse
      - LINKRIFT_CLICKHOUSE_PORT=9000
      - LINKRIFT_CLICKHOUSE_DATABASE=linkrift_analytics
      - LINKRIFT_LOG_LEVEL=debug
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      clickhouse:
        condition: service_healthy
    networks:
      - linkrift

  web:
    build:
      context: .
      dockerfile: docker/web/Dockerfile
      target: deps
    command: sh -c "cd /app && pnpm dev --host"
    volumes:
      - ./web:/app
      - web-node-modules:/app/node_modules
    ports:
      - "5173:5173"
    environment:
      - VITE_API_URL=http://localhost:8080
    networks:
      - linkrift

  # ===========================================
  # Infrastructure Services
  # ===========================================
  postgres:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=linkrift
      - POSTGRES_PASSWORD=linkrift_dev
      - POSTGRES_DB=linkrift
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init.sql:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U linkrift"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - linkrift

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - linkrift

  clickhouse:
    image: clickhouse/clickhouse-server:24-alpine
    ports:
      - "8123:8123"
      - "9000:9000"
    environment:
      - CLICKHOUSE_DB=linkrift_analytics
      - CLICKHOUSE_USER=linkrift
      - CLICKHOUSE_PASSWORD=linkrift_dev
    volumes:
      - clickhouse-data:/var/lib/clickhouse
      - ./scripts/init-clickhouse.sql:/docker-entrypoint-initdb.d/init.sql:ro
    healthcheck:
      test: ["CMD", "clickhouse-client", "--query", "SELECT 1"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - linkrift

  # ===========================================
  # Development Tools
  # ===========================================
  mailhog:
    image: mailhog/mailhog
    ports:
      - "1025:1025"  # SMTP
      - "8025:8025"  # Web UI
    networks:
      - linkrift

  adminer:
    image: adminer
    ports:
      - "8082:8080"
    environment:
      - ADMINER_DEFAULT_SERVER=postgres
    networks:
      - linkrift

  redis-commander:
    image: rediscommander/redis-commander
    ports:
      - "8083:8081"
    environment:
      - REDIS_HOSTS=local:redis:6379
    networks:
      - linkrift

networks:
  linkrift:
    driver: bridge

volumes:
  postgres-data:
  redis-data:
  clickhouse-data:
  go-cache:
  web-node-modules:
```

### Development Helper Script

```bash
#!/bin/bash
# dev.sh - Development environment helper

set -euo pipefail

COMPOSE_FILE="docker-compose.dev.yml"

case "${1:-help}" in
    up)
        echo "Starting development environment..."
        docker compose -f ${COMPOSE_FILE} up -d
        echo ""
        echo "Services:"
        echo "  API:        http://localhost:8080"
        echo "  Redirect:   http://localhost:8081"
        echo "  Web:        http://localhost:5173"
        echo "  Adminer:    http://localhost:8082"
        echo "  Redis UI:   http://localhost:8083"
        echo "  MailHog:    http://localhost:8025"
        ;;
    down)
        echo "Stopping development environment..."
        docker compose -f ${COMPOSE_FILE} down
        ;;
    logs)
        docker compose -f ${COMPOSE_FILE} logs -f "${2:-}"
        ;;
    restart)
        docker compose -f ${COMPOSE_FILE} restart "${2:-}"
        ;;
    rebuild)
        echo "Rebuilding services..."
        docker compose -f ${COMPOSE_FILE} build --no-cache "${2:-}"
        docker compose -f ${COMPOSE_FILE} up -d "${2:-}"
        ;;
    shell)
        service="${2:-api}"
        docker compose -f ${COMPOSE_FILE} exec "${service}" sh
        ;;
    migrate)
        echo "Running migrations..."
        docker compose -f ${COMPOSE_FILE} run --rm \
            -e DB_HOST=postgres \
            -e DB_USER=linkrift \
            -e DB_PASSWORD=linkrift_dev \
            -e DB_NAME=linkrift \
            -e MIGRATE_COMMAND="${2:-up}" \
            migrations
        ;;
    test)
        echo "Running tests..."
        docker compose -f ${COMPOSE_FILE} exec api go test -v ./...
        ;;
    clean)
        echo "Cleaning up..."
        docker compose -f ${COMPOSE_FILE} down -v --remove-orphans
        docker system prune -f
        ;;
    *)
        echo "Usage: $0 {up|down|logs|restart|rebuild|shell|migrate|test|clean}"
        exit 1
        ;;
esac
```

---

## Docker Compose Production

```yaml
# docker-compose.prod.yml

version: "3.9"

services:
  # ===========================================
  # Application Services
  # ===========================================
  api:
    image: ${REGISTRY:-ghcr.io/linkrift}/linkrift-api:${VERSION:-latest}
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 10s
        failure_action: rollback
        order: start-first
      rollback_config:
        parallelism: 1
        delay: 10s
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
        window: 120s
      resources:
        limits:
          cpus: "2"
          memory: 4G
        reservations:
          cpus: "0.5"
          memory: 512M
    environment:
      - LINKRIFT_ENV=production
      - LINKRIFT_API_PORT=8080
      - LINKRIFT_DB_HOST=${DB_HOST}
      - LINKRIFT_DB_PORT=${DB_PORT:-5432}
      - LINKRIFT_DB_NAME=${DB_NAME}
      - LINKRIFT_DB_USER=${DB_USER}
      - LINKRIFT_DB_PASSWORD=${DB_PASSWORD}
      - LINKRIFT_DB_SSL_MODE=require
      - LINKRIFT_DB_MAX_CONNECTIONS=100
      - LINKRIFT_REDIS_URL=${REDIS_URL}
      - LINKRIFT_JWT_SECRET=${JWT_SECRET}
      - LINKRIFT_LOG_LEVEL=info
      - LINKRIFT_LOG_FORMAT=json
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 10s
    networks:
      - linkrift-internal
      - linkrift-public
    logging:
      driver: json-file
      options:
        max-size: "100m"
        max-file: "5"

  redirect:
    image: ${REGISTRY:-ghcr.io/linkrift}/linkrift-redirect:${VERSION:-latest}
    deploy:
      replicas: 5
      update_config:
        parallelism: 1
        delay: 5s
        failure_action: rollback
        order: start-first
      restart_policy:
        condition: on-failure
        delay: 2s
        max_attempts: 5
        window: 60s
      resources:
        limits:
          cpus: "2"
          memory: 2G
        reservations:
          cpus: "1"
          memory: 256M
    environment:
      - LINKRIFT_ENV=production
      - LINKRIFT_REDIRECT_PORT=8081
      - LINKRIFT_DB_HOST=${DB_HOST}
      - LINKRIFT_DB_PORT=${DB_PORT:-5432}
      - LINKRIFT_DB_NAME=${DB_NAME}
      - LINKRIFT_DB_USER=${DB_USER_READONLY}
      - LINKRIFT_DB_PASSWORD=${DB_PASSWORD_READONLY}
      - LINKRIFT_DB_SSL_MODE=require
      - LINKRIFT_REDIS_URL=${REDIS_URL}
      - LINKRIFT_LOG_LEVEL=warn
      - LINKRIFT_LOG_FORMAT=json
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8081/health"]
      interval: 5s
      timeout: 2s
      retries: 3
      start_period: 5s
    networks:
      - linkrift-internal
      - linkrift-public
    logging:
      driver: json-file
      options:
        max-size: "50m"
        max-file: "3"

  worker:
    image: ${REGISTRY:-ghcr.io/linkrift}/linkrift-worker:${VERSION:-latest}
    deploy:
      replicas: 2
      update_config:
        parallelism: 1
        delay: 30s
        failure_action: rollback
      restart_policy:
        condition: on-failure
        delay: 10s
        max_attempts: 3
        window: 300s
      resources:
        limits:
          cpus: "2"
          memory: 4G
        reservations:
          cpus: "0.25"
          memory: 256M
    environment:
      - LINKRIFT_ENV=production
      - LINKRIFT_DB_HOST=${DB_HOST}
      - LINKRIFT_DB_PORT=${DB_PORT:-5432}
      - LINKRIFT_DB_NAME=${DB_NAME}
      - LINKRIFT_DB_USER=${DB_USER}
      - LINKRIFT_DB_PASSWORD=${DB_PASSWORD}
      - LINKRIFT_DB_SSL_MODE=require
      - LINKRIFT_REDIS_URL=${REDIS_URL_WORKER}
      - LINKRIFT_CLICKHOUSE_HOST=${CLICKHOUSE_HOST}
      - LINKRIFT_CLICKHOUSE_PORT=${CLICKHOUSE_PORT:-9000}
      - LINKRIFT_CLICKHOUSE_DATABASE=${CLICKHOUSE_DATABASE}
      - LINKRIFT_CLICKHOUSE_USER=${CLICKHOUSE_USER}
      - LINKRIFT_CLICKHOUSE_PASSWORD=${CLICKHOUSE_PASSWORD}
      - LINKRIFT_WORKER_CONCURRENCY=10
      - LINKRIFT_WORKER_BATCH_SIZE=1000
      - LINKRIFT_LOG_LEVEL=info
      - LINKRIFT_LOG_FORMAT=json
    networks:
      - linkrift-internal
    logging:
      driver: json-file
      options:
        max-size: "100m"
        max-file: "5"

  web:
    image: ${REGISTRY:-ghcr.io/linkrift}/linkrift-web:${VERSION:-latest}
    deploy:
      replicas: 2
      update_config:
        parallelism: 1
        delay: 10s
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
      resources:
        limits:
          cpus: "0.5"
          memory: 256M
        reservations:
          cpus: "0.1"
          memory: 64M
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost/health"]
      interval: 30s
      timeout: 5s
      retries: 3
    networks:
      - linkrift-public
    logging:
      driver: json-file
      options:
        max-size: "50m"
        max-file: "3"

  # ===========================================
  # Load Balancer
  # ===========================================
  traefik:
    image: traefik:v3.0
    command:
      - "--api.dashboard=true"
      - "--providers.docker=true"
      - "--providers.docker.swarmMode=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.letsencrypt.acme.tlschallenge=true"
      - "--certificatesresolvers.letsencrypt.acme.email=${ACME_EMAIL}"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
      - "--metrics.prometheus=true"
      - "--accesslog=true"
      - "--accesslog.format=json"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - traefik-certs:/letsencrypt
    deploy:
      placement:
        constraints:
          - node.role == manager
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.dashboard.rule=Host(`traefik.linkrift.io`)"
        - "traefik.http.routers.dashboard.service=api@internal"
        - "traefik.http.routers.dashboard.middlewares=auth"
        - "traefik.http.middlewares.auth.basicauth.users=${TRAEFIK_AUTH}"
    networks:
      - linkrift-public
    logging:
      driver: json-file
      options:
        max-size: "100m"
        max-file: "5"

networks:
  linkrift-internal:
    driver: overlay
    internal: true
  linkrift-public:
    driver: overlay

volumes:
  traefik-certs:
```

### Production Environment File

```bash
# .env.production

# Registry
REGISTRY=ghcr.io/linkrift
VERSION=1.0.0

# Database
DB_HOST=db.linkrift.internal
DB_PORT=5432
DB_NAME=linkrift
DB_USER=linkrift_api
DB_PASSWORD=<secure-password>
DB_USER_READONLY=linkrift_readonly
DB_PASSWORD_READONLY=<secure-password>

# Redis
REDIS_URL=rediss://:${REDIS_PASSWORD}@redis.linkrift.internal:6379/0
REDIS_URL_WORKER=rediss://:${REDIS_PASSWORD}@redis.linkrift.internal:6379/1
REDIS_PASSWORD=<secure-password>

# ClickHouse
CLICKHOUSE_HOST=clickhouse.linkrift.internal
CLICKHOUSE_PORT=9000
CLICKHOUSE_DATABASE=linkrift_analytics
CLICKHOUSE_USER=linkrift
CLICKHOUSE_PASSWORD=<secure-password>

# JWT
JWT_SECRET=<secure-256-bit-secret>

# SSL/TLS
ACME_EMAIL=admin@linkrift.io

# Traefik
TRAEFIK_AUTH=admin:$apr1$...
```

---

## Volume Management

### Volume Types and Usage

```yaml
# Volume configuration examples

volumes:
  # Named volume for persistent data
  postgres-data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /data/postgres

  # NFS volume for shared storage
  shared-uploads:
    driver: local
    driver_opts:
      type: nfs
      o: addr=nfs.linkrift.internal,rw
      device: ":/exports/linkrift/uploads"

  # tmpfs for high-performance temp storage
  redis-temp:
    driver: local
    driver_opts:
      type: tmpfs
      device: tmpfs
      o: size=1g,uid=999

  # AWS EBS volume
  clickhouse-data:
    driver: rexray/ebs
    driver_opts:
      size: 500
      volumetype: gp3
      iops: 3000
```

### Backup Script for Volumes

```bash
#!/bin/bash
# backup-volumes.sh

set -euo pipefail

BACKUP_DIR="/backups/linkrift/$(date +%Y%m%d)"
mkdir -p "${BACKUP_DIR}"

echo "=== Volume Backup ==="
echo "Backup directory: ${BACKUP_DIR}"
echo ""

# Backup PostgreSQL
echo "Backing up PostgreSQL..."
docker exec linkrift-postgres pg_dump -U linkrift -Fc linkrift > "${BACKUP_DIR}/postgres.dump"
echo "PostgreSQL backup: $(du -h ${BACKUP_DIR}/postgres.dump | cut -f1)"

# Backup Redis
echo "Backing up Redis..."
docker exec linkrift-redis redis-cli BGSAVE
sleep 5
docker cp linkrift-redis:/data/dump.rdb "${BACKUP_DIR}/redis.rdb"
echo "Redis backup: $(du -h ${BACKUP_DIR}/redis.rdb | cut -f1)"

# Backup ClickHouse
echo "Backing up ClickHouse..."
docker exec linkrift-clickhouse clickhouse-client \
    --query "BACKUP DATABASE linkrift_analytics TO Disk('backups', 'backup_$(date +%Y%m%d)')"
docker cp linkrift-clickhouse:/var/lib/clickhouse/backups "${BACKUP_DIR}/clickhouse"
echo "ClickHouse backup: $(du -sh ${BACKUP_DIR}/clickhouse | cut -f1)"

# Compress backups
echo ""
echo "Compressing backups..."
tar -czf "${BACKUP_DIR}.tar.gz" -C "$(dirname ${BACKUP_DIR})" "$(basename ${BACKUP_DIR})"
rm -rf "${BACKUP_DIR}"

echo ""
echo "Backup complete: ${BACKUP_DIR}.tar.gz"
echo "Size: $(du -h ${BACKUP_DIR}.tar.gz | cut -f1)"

# Upload to S3 (optional)
if [[ -n "${S3_BUCKET:-}" ]]; then
    echo ""
    echo "Uploading to S3..."
    aws s3 cp "${BACKUP_DIR}.tar.gz" "s3://${S3_BUCKET}/backups/"
    echo "Upload complete!"
fi
```

---

## Network Configuration

### Network Architecture

```yaml
# Network definitions

networks:
  # Public-facing network (load balancer, web)
  linkrift-public:
    driver: overlay
    driver_opts:
      encrypted: "true"
    ipam:
      config:
        - subnet: 10.10.0.0/24

  # Internal services network
  linkrift-internal:
    driver: overlay
    internal: true
    driver_opts:
      encrypted: "true"
    ipam:
      config:
        - subnet: 10.20.0.0/24

  # Database network (isolated)
  linkrift-data:
    driver: overlay
    internal: true
    driver_opts:
      encrypted: "true"
    ipam:
      config:
        - subnet: 10.30.0.0/24
```

### Service Network Configuration

```yaml
services:
  api:
    networks:
      linkrift-public:
        aliases:
          - api
      linkrift-internal:
        aliases:
          - api-internal
      linkrift-data:

  redirect:
    networks:
      linkrift-public:
        aliases:
          - redirect
      linkrift-internal:

  worker:
    networks:
      linkrift-internal:
      linkrift-data:

  postgres:
    networks:
      linkrift-data:
        aliases:
          - postgres
          - db

  redis:
    networks:
      linkrift-internal:
        aliases:
          - redis
          - cache
```

### Network Security with IPTables

```bash
#!/bin/bash
# network-security.sh

# Allow established connections
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# Allow Docker networks
iptables -A INPUT -i docker0 -j ACCEPT
iptables -A INPUT -i br-+ -j ACCEPT

# Allow specific ports
iptables -A INPUT -p tcp --dport 22 -j ACCEPT    # SSH
iptables -A INPUT -p tcp --dport 80 -j ACCEPT    # HTTP
iptables -A INPUT -p tcp --dport 443 -j ACCEPT   # HTTPS

# Block all other incoming traffic
iptables -A INPUT -j DROP

# Save rules
iptables-save > /etc/iptables/rules.v4
```

---

## Best Practices

### 1. Image Optimization

```dockerfile
# Use specific versions, not latest
FROM golang:1.22.0-alpine3.19

# Combine RUN commands to reduce layers
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# Use .dockerignore
```

**.dockerignore**:

```
.git
.github
.gitignore
*.md
!README.md
docker-compose*.yml
.env*
coverage/
dist/
node_modules/
.vscode/
.idea/
*.log
*.test
*_test.go
```

### 2. Security Best Practices

```dockerfile
# Run as non-root user
RUN addgroup -g 1000 app && adduser -u 1000 -G app -s /bin/sh -D app
USER app

# Use read-only filesystem
# In docker-compose.yml:
# read_only: true
# tmpfs:
#   - /tmp

# Scan images for vulnerabilities
# docker scout cves linkrift-api:latest
```

### 3. Health Check Patterns

```dockerfile
# HTTP health check
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# TCP health check
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
    CMD nc -z localhost 8080 || exit 1

# Custom script health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD /app/healthcheck.sh || exit 1
```

### 4. Logging Configuration

```yaml
services:
  api:
    logging:
      driver: json-file
      options:
        max-size: "100m"
        max-file: "5"
        labels: "service,environment"
        env: "LINKRIFT_ENV"
    labels:
      - "service=api"
      - "environment=production"
```

### 5. Resource Management

```yaml
services:
  api:
    deploy:
      resources:
        limits:
          cpus: "2"
          memory: 4G
        reservations:
          cpus: "0.5"
          memory: 512M
    # Kernel parameter tuning
    sysctls:
      net.core.somaxconn: 65535
      net.ipv4.tcp_syncookies: 1
    ulimits:
      nofile:
        soft: 65535
        hard: 65535
```

---

## Quick Reference Commands

```bash
# Build all images
docker compose -f docker-compose.dev.yml build

# Start development environment
docker compose -f docker-compose.dev.yml up -d

# View logs
docker compose -f docker-compose.dev.yml logs -f api

# Execute command in container
docker compose -f docker-compose.dev.yml exec api sh

# Run migrations
docker compose -f docker-compose.dev.yml run --rm migrations up

# Stop and remove containers
docker compose -f docker-compose.dev.yml down

# Remove volumes (data loss!)
docker compose -f docker-compose.dev.yml down -v

# Production deployment
docker stack deploy -c docker-compose.prod.yml linkrift

# Scale service
docker service scale linkrift_api=5

# Update service
docker service update --image ghcr.io/linkrift/linkrift-api:v1.1.0 linkrift_api

# Rollback service
docker service rollback linkrift_api
```

---

*This Docker setup guide is maintained by the Linkrift Platform Team. For questions or updates, contact platform@linkrift.io*
