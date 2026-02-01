# Quick Start

> Last Updated: 2025-01-24

Get Linkrift running in under 5 minutes.

## TL;DR

```bash
git clone https://github.com/link-rift/link-rift.git
cd link-rift
make docker-up
```

Open `http://localhost:3000` — Login with `admin@linkrift.local` / `linkrift123`

---

## Prerequisites

| Tool | Version | Check |
|------|---------|-------|
| Docker | 24+ | `docker --version` |
| Docker Compose | 2.20+ | `docker compose version` |
| Make | Any | `make --version` |

## Option 1: Docker Compose (Recommended)

```bash
# Clone
git clone https://github.com/link-rift/link-rift.git
cd link-rift

# Start all services
make docker-up

# Wait for health checks (~30 seconds)
docker compose ps
```

**Services started:**
- API: `http://localhost:8080`
- Redirect: `http://localhost:8081`
- Web UI: `http://localhost:3000`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- ClickHouse: `localhost:8123`

## Option 2: Local Development

```bash
# Clone
git clone https://github.com/link-rift/link-rift.git
cd link-rift

# Environment
cp .env.example .env

# Start databases only
docker compose up -d postgres redis clickhouse

# Install Go dependencies
go mod download

# Run migrations
make migrate-up

# Start API (terminal 1)
make dev-api

# Start frontend (terminal 2)
cd web && pnpm install && pnpm dev
```

## Default Credentials

| Service | Username/Email | Password |
|---------|---------------|----------|
| Admin | `admin@linkrift.local` | `linkrift123` |
| PostgreSQL | `linkrift` | `linkrift` |
| Redis | - | - |
| ClickHouse | `default` | - |

> ⚠️ **Warning**: Change default credentials before deploying to production.

## First Steps

1. **Create a short link**
   - Click "New Link" in the dashboard
   - Enter destination URL
   - Optionally customize the short code

2. **View analytics**
   - Click on any link to see real-time stats
   - Geographic, device, and referrer breakdowns

3. **Add a custom domain**
   - Go to Settings → Domains
   - Add your domain and follow DNS instructions

## Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# Create a link via API
curl -X POST http://localhost:8080/api/v1/links \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"url": "https://example.com"}'

# Test redirect
curl -I http://localhost:8081/abc123
```

## Stopping Services

```bash
# Stop all containers
make docker-down

# Stop and remove volumes (⚠️ deletes data)
docker compose down -v
```

## Next Steps

- [Setup Guide](SETUP_GUIDE.md) — Detailed installation instructions
- [Development Guide](DEVELOPMENT_GUIDE.md) — Development workflow
- [Architecture](../architecture/ARCHITECTURE.md) — System design overview
- [API Documentation](../api/API_DOCUMENTATION.md) — API reference

## Troubleshooting

### Port conflicts

```bash
# Check what's using port 8080
lsof -i :8080

# Use different ports
API_PORT=9080 WEB_PORT=4000 make docker-up
```

### Database connection issues

```bash
# Check database logs
docker compose logs postgres

# Reset database
docker compose down -v
make docker-up
```

### Frontend not loading

```bash
# Check web container logs
docker compose logs web

# Rebuild frontend
docker compose build web
docker compose up -d web
```

---

**Need help?** Open an issue on [GitHub](https://github.com/link-rift/link-rift/issues).
