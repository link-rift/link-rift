# Setup Guide

> Last Updated: 2025-01-24

Complete installation guide for Linkrift development environment.

## Table of Contents

- [Prerequisites](#prerequisites)
- [System Requirements](#system-requirements)
- [Installation](#installation)
- [Database Setup](#database-setup)
- [Environment Configuration](#environment-configuration)
- [External Services](#external-services)
- [Running the Application](#running-the-application)
- [Verification](#verification)
- [Common Issues](#common-issues)

---

## Prerequisites

### Required Tools

| Tool | Version | Installation |
|------|---------|--------------|
| Go | 1.22+ | [go.dev/dl](https://go.dev/dl/) |
| Node.js | 20+ | [nodejs.org](https://nodejs.org/) |
| pnpm | 8+ | `npm install -g pnpm` |
| Docker | 24+ | [docker.com](https://www.docker.com/) |
| Docker Compose | 2.20+ | Included with Docker Desktop |
| Make | Any | Pre-installed on macOS/Linux |
| Git | 2.40+ | [git-scm.com](https://git-scm.com/) |

### Verify Installations

```bash
go version          # go version go1.22.0 or higher
node --version      # v20.0.0 or higher
pnpm --version      # 8.0.0 or higher
docker --version    # Docker version 24.0.0 or higher
docker compose version  # Docker Compose version v2.20.0 or higher
make --version      # GNU Make 3.81 or higher
git --version       # git version 2.40.0 or higher
```

### Development Tools (Recommended)

```bash
# Install Go development tools
go install github.com/air-verse/air@latest           # Hot reload
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest  # SQL code generation
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest  # Migrations
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest  # Linter

# Verify tools
air -v
sqlc version
migrate -version
golangci-lint --version
```

### IDE Setup (VS Code)

Install recommended extensions:

```json
{
  "recommendations": [
    "golang.go",
    "bradlc.vscode-tailwindcss",
    "esbenp.prettier-vscode",
    "dbaeumer.vscode-eslint",
    "ms-azuretools.vscode-docker"
  ]
}
```

Go extension settings (`.vscode/settings.json`):

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "gopls": {
    "formatting.gofumpt": true
  }
}
```

---

## System Requirements

### Minimum

| Resource | Requirement |
|----------|-------------|
| CPU | 2 cores |
| RAM | 4 GB |
| Disk | 10 GB |
| OS | macOS 12+, Ubuntu 20.04+, Windows 11 (WSL2) |

### Recommended

| Resource | Requirement |
|----------|-------------|
| CPU | 4+ cores |
| RAM | 8+ GB |
| Disk | 20+ GB SSD |
| OS | macOS 13+, Ubuntu 22.04+ |

---

## Installation

### 1. Clone Repository

```bash
git clone https://github.com/link-rift/link-rift.git
cd link-rift
```

### 2. Install Go Dependencies

```bash
go mod download
go mod verify
```

### 3. Install Frontend Dependencies

```bash
cd web
pnpm install
cd ..
```

### 4. Copy Environment File

```bash
cp .env.example .env
```

---

## Database Setup

### Start Infrastructure Services

```bash
# Start PostgreSQL, Redis, ClickHouse, Meilisearch
docker compose up -d postgres redis clickhouse meilisearch

# Verify services are running
docker compose ps
```

### PostgreSQL Setup

```bash
# Run migrations
make migrate-up

# Verify migrations
make migrate-status

# Seed development data (optional)
make seed
```

### Generate sqlc Code

```bash
# Generate Go code from SQL queries
make sqlc

# Verify generated files
ls internal/repository/sqlc/
```

### Redis Verification

```bash
# Connect to Redis
docker compose exec redis redis-cli ping
# Should return: PONG
```

### ClickHouse Verification

```bash
# Connect to ClickHouse
docker compose exec clickhouse clickhouse-client --query "SELECT 1"
# Should return: 1
```

---

## Environment Configuration

### Core Settings

```bash
# .env

# Application
APP_ENV=development
APP_DEBUG=true
APP_URL=http://localhost:3000
API_URL=http://localhost:8080

# Server
API_PORT=8080
REDIRECT_PORT=8081
WEB_PORT=3000

# Database
DATABASE_URL=postgres://linkrift:linkrift@localhost:5432/linkrift?sslmode=disable
DATABASE_MAX_CONNECTIONS=25
DATABASE_MAX_IDLE_CONNECTIONS=5

# Redis
REDIS_URL=redis://localhost:6379/0
REDIS_POOL_SIZE=10

# ClickHouse
CLICKHOUSE_URL=http://localhost:8123
CLICKHOUSE_DATABASE=linkrift

# Meilisearch
MEILISEARCH_URL=http://localhost:7700
MEILISEARCH_API_KEY=masterKey

# Security
JWT_SECRET=your-super-secret-key-change-in-production
PASETO_KEY=your-32-byte-paseto-key-here123
ENCRYPTION_KEY=your-32-byte-encryption-key-here
```

### External Services (Optional for Development)

```bash
# Stripe (billing)
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_PRO=price_...
STRIPE_PRICE_BUSINESS=price_...

# Cloudflare (custom domains)
CLOUDFLARE_API_TOKEN=...
CLOUDFLARE_ZONE_ID=...

# AWS S3 (file storage)
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...
AWS_REGION=us-east-1
AWS_S3_BUCKET=linkrift-uploads

# Email (Resend)
RESEND_API_KEY=re_...
EMAIL_FROM=noreply@linkrift.io

# MaxMind GeoIP
MAXMIND_LICENSE_KEY=...
MAXMIND_ACCOUNT_ID=...

# OAuth Providers
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
GITHUB_CLIENT_ID=...
GITHUB_CLIENT_SECRET=...

# Monitoring
SENTRY_DSN=...
```

---

## External Services

### Stripe (Required for Billing)

1. Create account at [stripe.com](https://stripe.com)
2. Get test API keys from Dashboard → Developers → API keys
3. Create products and prices in Stripe Dashboard
4. Set up webhook endpoint: `https://your-domain.com/api/v1/webhooks/stripe`

### Cloudflare (Required for Custom Domains)

1. Create account at [cloudflare.com](https://cloudflare.com)
2. Create API token with Zone:DNS:Edit permissions
3. Note your Zone ID from the domain overview page

### AWS S3 (Required for File Uploads)

1. Create S3 bucket with appropriate CORS settings
2. Create IAM user with S3 access
3. Configure bucket policy for public read access to uploads

### MaxMind GeoIP (Required for Geographic Analytics)

1. Create account at [maxmind.com](https://www.maxmind.com/)
2. Generate license key
3. Download GeoLite2-City database

---

## Running the Application

### Development Mode (with Hot Reload)

```bash
# Terminal 1: API Server
make dev-api

# Terminal 2: Redirect Service
make dev-redirect

# Terminal 3: Worker
make dev-worker

# Terminal 4: Frontend
make dev-web
```

### All-in-One Docker

```bash
make docker-up
```

### Access Points

| Service | URL |
|---------|-----|
| Web Dashboard | http://localhost:3000 |
| API | http://localhost:8080 |
| Redirect Service | http://localhost:8081 |
| API Docs (Swagger) | http://localhost:8080/swagger/index.html |
| ClickHouse UI | http://localhost:8123/play |
| Meilisearch UI | http://localhost:7700 |

---

## Verification

### Health Checks

```bash
# API health
curl http://localhost:8080/health

# Redirect service health
curl http://localhost:8081/health

# Expected response:
# {"status":"healthy","version":"0.1.0","timestamp":"..."}
```

### Create Test Link

```bash
# Login and get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@linkrift.local","password":"linkrift123"}' \
  | jq -r '.access_token')

# Create link
curl -X POST http://localhost:8080/api/v1/links \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"url":"https://example.com","custom_code":"test123"}'

# Test redirect
curl -I http://localhost:8081/test123
# Should return: HTTP/1.1 301 Moved Permanently
```

### Run Tests

```bash
# All tests
make test

# With coverage
make test-cover

# Specific package
go test ./internal/service/...
```

---

## Common Issues

### Port Already in Use

```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>

# Or use different ports
API_PORT=9080 make dev-api
```

### Database Connection Failed

```bash
# Check if PostgreSQL is running
docker compose ps postgres

# View logs
docker compose logs postgres

# Restart PostgreSQL
docker compose restart postgres

# Reset database
docker compose down -v
docker compose up -d postgres
make migrate-up
```

### Go Module Issues

```bash
# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download

# Verify modules
go mod verify
```

### Frontend Build Errors

```bash
# Clear node_modules
cd web
rm -rf node_modules
pnpm install

# Clear Vite cache
rm -rf node_modules/.vite
pnpm dev
```

### Hot Reload Not Working

```bash
# Ensure air is installed
go install github.com/air-verse/air@latest

# Check air config exists
cat .air.toml

# Run air directly
air -c .air.toml
```

### ClickHouse Connection Issues

```bash
# Check ClickHouse status
docker compose logs clickhouse

# Connect manually
docker compose exec clickhouse clickhouse-client

# Verify database exists
docker compose exec clickhouse clickhouse-client --query "SHOW DATABASES"
```

---

## Next Steps

- [Development Guide](DEVELOPMENT_GUIDE.md) — Development workflow and patterns
- [Architecture](../architecture/ARCHITECTURE.md) — System design overview
- [Database Schema](../architecture/DATABASE_SCHEMA.md) — Database structure
- [API Documentation](../api/API_DOCUMENTATION.md) — API reference

---

**Need help?** Open an issue on [GitHub](https://github.com/link-rift/link-rift/issues).
