# Linkrift

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)

<!-- TODO: Enable these badges once GitHub repo and CI/CD are set up
[![Build Status](https://github.com/link-rift/link-rift/actions/workflows/ci.yml/badge.svg)](https://github.com/link-rift/link-rift/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/link-rift/link-rift)](https://goreportcard.com/report/github.com/link-rift/link-rift)
[![Coverage](https://codecov.io/gh/link-rift/link-rift/branch/main/graph/badge.svg)](https://codecov.io/gh/link-rift/link-rift)
-->

**Enterprise-grade URL shortener with analytics, custom domains, and team collaboration.**

Linkrift is an open-source, self-hostable URL shortener platform that combines powerful analytics, smart redirects, team collaboration, and marketing automation. Built with Go for sub-millisecond redirect latency and React for a modern user experience.

<!-- TODO: Add dashboard screenshot once available
![Linkrift Dashboard](docs/assets/dashboard-preview.png)
-->

## Key Features

- **High-Performance Redirects** — Sub-millisecond latency with Go, handling millions of redirects per day
- **Real-Time Analytics** — Geographic, device, and referrer insights with ClickHouse
- **Custom Domains** — Unlimited branded domains with automatic SSL via Cloudflare
- **Smart Redirects** — Device, geo, time-based, and A/B testing rules
- **QR Codes** — Dynamic QR codes with customization and analytics
- **Bio Pages** — Customizable link-in-bio pages with drag-drop builder
- **Team Collaboration** — Workspaces, roles, and audit logs
- **API & Webhooks** — RESTful API with SDKs and real-time webhooks
- **Self-Hostable** — Deploy on your infrastructure with Docker or Kubernetes

## Tech Stack

| Layer | Technologies |
|-------|-------------|
| **Backend** | Go 1.22, Gin, sqlc, pgx |
| **Frontend** | React 18, TypeScript, Vite 5, Tailwind CSS, Shadcn UI |
| **Database** | PostgreSQL 16, Redis 7, ClickHouse |
| **Infrastructure** | Docker, Kubernetes, NGINX, Cloudflare |

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 20+ with pnpm
- Docker & Docker Compose
- Make

### One-Command Setup

```bash
git clone https://github.com/link-rift/link-rift.git
cd link-rift
make docker-up
```

Access the dashboard at `http://localhost:3000` with default credentials:
- Email: `admin@linkrift.local`
- Password: `linkrift123`

### Development Setup

```bash
# Clone repository
git clone https://github.com/link-rift/link-rift.git
cd link-rift

# Copy environment file
cp .env.example .env

# Start infrastructure (PostgreSQL, Redis, ClickHouse)
docker compose up -d postgres redis clickhouse

# Run database migrations
make migrate-up

# Start backend (with hot reload)
make dev-api

# In another terminal, start frontend
make dev-web
```

## Project Structure

```
linkrift/
├── cmd/                    # Application entry points
│   ├── api/                # Main API server
│   ├── redirect/           # Optimized redirect service
│   ├── worker/             # Background job processor
│   └── cli/                # CLI tool
├── internal/               # Private application code
│   ├── handler/            # HTTP handlers (Gin)
│   ├── service/            # Business logic
│   ├── repository/         # Data access layer
│   ├── middleware/         # HTTP middleware
│   └── ee/                 # Enterprise Edition features
├── pkg/                    # Public packages
├── web/                    # React frontend (Vite)
├── migrations/             # Database migrations
├── sqlc/                   # SQL queries and sqlc config
├── deployments/            # Docker Compose and Kubernetes
└── docs/                   # Documentation
```

## Documentation

| Document | Description |
|----------|-------------|
| [Quick Start](docs/getting-started/QUICK_START.md) | Get running in 5 minutes |
| [Setup Guide](docs/getting-started/SETUP_GUIDE.md) | Detailed installation guide |
| [Architecture](docs/architecture/ARCHITECTURE.md) | System design and data flows |
| [API Reference](docs/api/API_DOCUMENTATION.md) | Complete API documentation |
| [Deployment](docs/deployment/DEPLOYMENT_GUIDE.md) | Production deployment guide |
| [Contributing](docs/contributing/CONTRIBUTING.md) | How to contribute |

## Common Commands

```bash
# Development
make dev              # Run all services in development mode
make dev-api          # Run API server with hot reload
make dev-web          # Run Vite dev server

# Build
make build            # Build all binaries
make build-web        # Build frontend for production

# Database
make migrate-up       # Run migrations
make migrate-create   # Create new migration
make sqlc             # Generate sqlc code

# Testing
make test             # Run all tests
make test-cover       # Run tests with coverage
make bench            # Run benchmarks

# Docker
make docker-build     # Build Docker images
make docker-up        # Start all services
make docker-down      # Stop all services

# Linting
make lint             # Run golangci-lint
make fmt              # Format code
```

## Open Core Model

Linkrift uses an open core model:

| Feature | Community (Free) | Enterprise |
|---------|-----------------|------------|
| URL Shortening | ✅ | ✅ |
| Basic Analytics (7 days) | ✅ | ✅ |
| Custom Domains | 1 | Unlimited |
| Team Members | 3 | Unlimited |
| QR Codes | ✅ | ✅ (Branded) |
| SSO/SAML | ❌ | ✅ |
| Advanced Audit Logs | ❌ | ✅ |
| White-Label | ❌ | ✅ |
| Priority Support | ❌ | ✅ |

See [Open Core Model](docs/open-core/OPEN_CORE_MODEL.md) for details.

## Performance

Linkrift is built for speed:

| Metric | Target | Typical |
|--------|--------|---------|
| Redirect Latency (p99) | <5ms | ~1ms |
| Memory per Instance | <50MB | ~20MB |
| Requests per Instance | 50K/sec | 100K+/sec |

See [Performance Optimization](docs/reference/PERFORMANCE_OPTIMIZATION.md) for tuning guides.

## Contributing

We welcome contributions! Please read our [Contributing Guide](docs/contributing/CONTRIBUTING.md) before submitting a pull request.

```bash
# Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/link-rift.git

# Create a feature branch
git checkout -b feature/your-feature

# Make your changes and run tests
make test
make lint

# Submit a pull request
```

## Community

- [GitHub Issues](https://github.com/link-rift/link-rift/issues) — Bug reports and feature requests
- [GitHub Discussions](https://github.com/link-rift/link-rift/discussions) — Questions and ideas
- [Discord](https://discord.gg/linkrift) — Real-time chat

## License

Linkrift is licensed under the [GNU Affero General Public License v3.0](LICENSE).

Enterprise features require a commercial license. Contact sales@linkrift.io for pricing.

---

**Built with Go and React by the Linkrift team**
