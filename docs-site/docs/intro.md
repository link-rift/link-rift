---
sidebar_position: 1
slug: /
---

# Linkrift Documentation

Welcome to the Linkrift documentation. Linkrift is an enterprise-grade URL shortener with analytics, custom domains, and team collaboration.

## Quick Navigation

### Getting Started

- **[Quick Start](getting-started/QUICK_START.md)** — Get up and running in minutes
- **[Setup Guide](getting-started/SETUP_GUIDE.md)** — Detailed setup instructions
- **[Development Guide](getting-started/DEVELOPMENT_GUIDE.md)** — Local development workflow

### Architecture

- **[Architecture Overview](architecture/ARCHITECTURE.md)** — System design and components
- **[Database Schema](architecture/DATABASE_SCHEMA.md)** — Tables, indexes, and relationships
- **[Go Patterns](architecture/GO_PATTERNS.md)** — Backend conventions and patterns
- **[Frontend Architecture](architecture/FRONTEND_ARCHITECTURE.md)** — React, TypeScript, and state management

### API

- **[API Documentation](api/API_DOCUMENTATION.md)** — REST API reference
- **[SDK Documentation](api/SDK_DOCUMENTATION.md)** — Client SDK usage
- **[CLI Tool](api/CLI_TOOL.md)** — Command-line interface

### Features

- **[Authentication](features/AUTHENTICATION.md)** — Auth flows, sessions, PASETO tokens
- **[Analytics Pipeline](features/ANALYTICS_PIPELINE.md)** — Click tracking and reporting
- **[Custom Domains](features/CUSTOM_DOMAINS.md)** — Branded short links
- **[QR Codes](features/QR_CODES.md)** — QR code generation and customization
- **[Bio Pages](features/BIO_PAGES.md)** — Link-in-bio pages
- **[Webhooks](features/WEBHOOKS.md)** — Event notifications
- **[Team Collaboration](features/TEAM_COLLABORATION.md)** — Workspaces, roles, and permissions

## Tech Stack

| Layer | Technologies |
|-------|-------------|
| Backend | Go 1.22, Gin, sqlc, pgx, go-redis, uber/zap |
| Frontend | React 18, TypeScript, Vite 5, Tailwind CSS, Shadcn UI |
| Database | PostgreSQL 16, Redis 7, ClickHouse (analytics) |
| Infrastructure | Docker, Kubernetes, NGINX, Cloudflare |

## Open Core Model

Linkrift is open-source (AGPL-3.0) with enterprise features available under a commercial license:

- **Community Edition** — Free, self-hostable with core features
- **Enterprise Edition** — SSO/SAML, audit logs, white-label, SCIM provisioning
