# Linkrift Repository Structure

**Last Updated: 2025-01-24**

---

## Table of Contents

- [Overview](#overview)
- [Monorepo Structure](#monorepo-structure)
  - [Directory Layout](#directory-layout)
  - [Package Organization](#package-organization)
  - [Build System](#build-system)
- [Related Repositories](#related-repositories)
  - [link-rift-ee (Enterprise Edition)](#link-rift-ee-enterprise-edition)
  - [link-rift-docs](#link-rift-docs)
  - [SDKs](#sdks)
  - [Other Repositories](#other-repositories)
- [Branch Strategy](#branch-strategy)
  - [Branch Naming](#branch-naming)
  - [Branch Protection](#branch-protection)
  - [Merge Strategy](#merge-strategy)
- [CODEOWNERS and Protected Branches](#codeowners-and-protected-branches)
  - [CODEOWNERS Configuration](#codeowners-configuration)
  - [Branch Protection Rules](#branch-protection-rules)
- [Release Process](#release-process)
  - [Versioning](#versioning)
  - [Release Workflow](#release-workflow)
  - [Changelog Generation](#changelog-generation)

---

## Overview

Linkrift uses a **monorepo architecture** for the main project, with separate repositories for Enterprise Edition features, documentation, and SDKs. This structure enables:

- **Unified development** - All core components in one place
- **Atomic changes** - Cross-cutting changes in single commits
- **Consistent tooling** - Shared linting, testing, and CI/CD
- **Clear boundaries** - Enterprise code physically separated

---

## Monorepo Structure

### Directory Layout

```
link-rift/
├── .github/
│   ├── workflows/           # GitHub Actions CI/CD
│   │   ├── ci.yml
│   │   ├── release.yml
│   │   └── security.yml
│   ├── CODEOWNERS
│   ├── ISSUE_TEMPLATE/
│   └── PULL_REQUEST_TEMPLATE.md
├── api/
│   ├── openapi/             # OpenAPI specifications
│   │   ├── v1.yaml
│   │   └── v2.yaml
│   └── proto/               # Protocol Buffers (gRPC)
│       └── linkrift/
│           ├── links.proto
│           └── analytics.proto
├── cmd/
│   ├── linkrift/            # Main application binary
│   │   └── main.go
│   ├── migrate/             # Database migration tool
│   │   └── main.go
│   └── worker/              # Background job worker
│       └── main.go
├── configs/
│   ├── config.example.yaml
│   ├── config.development.yaml
│   └── config.production.yaml
├── deploy/
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   ├── kubernetes/
│   │   ├── base/
│   │   └── overlays/
│   └── helm/
│       └── linkrift/
├── docs/
│   ├── api/
│   ├── architecture/
│   ├── open-core/
│   └── self-hosting/
├── internal/
│   ├── api/                 # HTTP/gRPC handlers
│   │   ├── handlers/
│   │   ├── middleware/
│   │   └── routes.go
│   ├── config/              # Configuration management
│   ├── database/            # Database layer
│   │   ├── migrations/
│   │   ├── postgres/
│   │   └── sqlite/
│   ├── domain/              # Domain models
│   │   ├── link.go
│   │   ├── user.go
│   │   └── workspace.go
│   ├── license/             # License verification
│   │   ├── keys/
│   │   ├── license.go
│   │   ├── manager.go
│   │   └── verifier.go
│   ├── services/            # Business logic
│   │   ├── analytics/
│   │   ├── links/
│   │   └── users/
│   └── worker/              # Background job processing
│       ├── jobs/
│       └── scheduler.go
├── pkg/                     # Public packages (importable)
│   ├── shortid/             # Short ID generation
│   ├── validator/           # Input validation
│   └── webhook/             # Webhook utilities
├── scripts/
│   ├── build.sh
│   ├── generate.sh
│   └── migrate.sh
├── test/
│   ├── e2e/                 # End-to-end tests
│   ├── integration/         # Integration tests
│   └── fixtures/            # Test data
├── web/                     # Frontend application
│   ├── src/
│   │   ├── components/
│   │   ├── contexts/
│   │   ├── hooks/
│   │   ├── lib/
│   │   ├── pages/
│   │   └── styles/
│   ├── public/
│   ├── package.json
│   └── vite.config.ts
├── .gitignore
├── .golangci.yml
├── CHANGELOG.md
├── CONTRIBUTING.md
├── LICENSE                  # AGPL-3.0
├── Makefile
├── README.md
├── go.mod
├── go.sum
└── turbo.json               # Turborepo configuration
```

### Package Organization

```go
// go.mod
module github.com/link-rift/link-rift

go 1.22

require (
    github.com/go-chi/chi/v5 v5.0.12
    github.com/jackc/pgx/v5 v5.5.3
    github.com/redis/go-redis/v9 v9.4.0
    github.com/rs/zerolog v1.32.0
    // ... other dependencies
)
```

**Package Responsibilities:**

| Package | Purpose | Visibility |
|---------|---------|------------|
| `cmd/` | Application entry points | Internal |
| `internal/` | Private application code | Internal |
| `pkg/` | Reusable libraries | Public |
| `api/` | API specifications | Public |
| `web/` | Frontend application | Internal |

### Build System

```makefile
# Makefile

.PHONY: all build test lint clean

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

all: lint test build

## build: Build all binaries
build: build-server build-worker build-migrate

build-server:
	$(GOBUILD) $(LDFLAGS) -o bin/linkrift ./cmd/linkrift

build-worker:
	$(GOBUILD) $(LDFLAGS) -o bin/linkrift-worker ./cmd/worker

build-migrate:
	$(GOBUILD) $(LDFLAGS) -o bin/linkrift-migrate ./cmd/migrate

## test: Run all tests
test:
	$(GOTEST) -race -coverprofile=coverage.out ./...

## test-integration: Run integration tests
test-integration:
	$(GOTEST) -race -tags=integration ./test/integration/...

## test-e2e: Run end-to-end tests
test-e2e:
	$(GOTEST) -race -tags=e2e ./test/e2e/...

## lint: Run linters
lint:
	golangci-lint run ./...
	cd web && npm run lint

## generate: Generate code (protobuf, mocks, etc.)
generate:
	go generate ./...
	buf generate api/proto

## migrate: Run database migrations
migrate:
	./bin/linkrift-migrate up

## docker-build: Build Docker image
docker-build:
	docker build -t ghcr.io/link-rift/link-rift:$(VERSION) -f deploy/docker/Dockerfile .

## clean: Clean build artifacts
clean:
	rm -rf bin/
	rm -rf web/dist/
	rm -f coverage.out

## help: Show this help
help:
	@grep -E '^##' Makefile | sed 's/## //'
```

---

## Related Repositories

### link-rift-ee (Enterprise Edition)

The Enterprise Edition repository contains proprietary features not available in the Community Edition.

```
link-rift-ee/                    (Private Repository)
├── .github/
│   └── workflows/
├── ee/
│   ├── analytics/               # Advanced analytics
│   │   ├── custom_reports.go
│   │   └── realtime.go
│   ├── auth/                    # Enterprise auth
│   │   ├── saml/
│   │   ├── scim/
│   │   └── oidc/
│   ├── audit/                   # Audit logging
│   │   └── logger.go
│   ├── compliance/              # Compliance features
│   │   ├── gdpr.go
│   │   └── hipaa.go
│   ├── domains/                 # Advanced domain features
│   │   └── white_label.go
│   └── rbac/                    # Advanced RBAC
│       └── policies.go
├── LICENSE                      # Proprietary license
├── README.md
├── go.mod
└── go.sum
```

**Integration with CE:**

```go
// go.mod (link-rift-ee)
module github.com/link-rift/link-rift-ee

go 1.22

require (
    github.com/link-rift/link-rift v1.0.0
)

replace github.com/link-rift/link-rift => ../link-rift
```

```go
// internal/enterprise/loader.go (in main link-rift repo)
//go:build enterprise

package enterprise

import (
    "github.com/link-rift/link-rift-ee/ee/analytics"
    "github.com/link-rift/link-rift-ee/ee/auth/saml"
    "github.com/link-rift/link-rift-ee/ee/audit"
)

func init() {
    // Register enterprise features
    RegisterFeature("advanced_analytics", analytics.New())
    RegisterFeature("saml", saml.New())
    RegisterFeature("audit_logs", audit.New())
}
```

### link-rift-docs

Documentation website repository.

```
link-rift-docs/                  (Public Repository)
├── .github/
│   └── workflows/
│       └── deploy.yml
├── docs/
│   ├── getting-started/
│   ├── guides/
│   ├── api-reference/
│   ├── self-hosting/
│   └── enterprise/
├── blog/
├── src/
│   ├── components/
│   └── pages/
├── static/
├── docusaurus.config.js
└── package.json
```

### SDKs

Official SDK repositories:

| Repository | Language | Status |
|------------|----------|--------|
| `link-rift-js` | JavaScript/TypeScript | Stable |
| `link-rift-python` | Python | Stable |
| `link-rift-go` | Go | Stable |
| `link-rift-ruby` | Ruby | Beta |
| `link-rift-php` | PHP | Beta |
| `link-rift-java` | Java | Planned |

```
link-rift-js/                    (Public Repository)
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── publish.yml
├── src/
│   ├── client.ts
│   ├── links.ts
│   ├── analytics.ts
│   └── types.ts
├── test/
├── examples/
├── LICENSE                      # MIT
├── README.md
├── package.json
└── tsconfig.json
```

### Other Repositories

| Repository | Purpose |
|------------|---------|
| `link-rift-helm` | Helm charts for Kubernetes deployment |
| `link-rift-terraform` | Terraform modules for cloud deployment |
| `link-rift-ansible` | Ansible playbooks for bare metal deployment |
| `link-rift-homebrew` | Homebrew tap for macOS installation |

---

## Branch Strategy

Linkrift follows a **trunk-based development** model with short-lived feature branches.

### Branch Naming

```
main                 # Production-ready code
├── feature/         # New features
│   ├── feature/add-qr-codes
│   └── feature/api-v2
├── fix/             # Bug fixes
│   ├── fix/rate-limit-bypass
│   └── fix/redirect-loop
├── docs/            # Documentation changes
│   └── docs/api-examples
├── refactor/        # Code refactoring
│   └── refactor/database-layer
├── chore/           # Maintenance tasks
│   └── chore/update-dependencies
└── release/         # Release branches
    └── release/v1.2.0
```

**Branch Naming Convention:**

```
<type>/<issue-number>-<short-description>

Examples:
- feature/123-add-custom-domains
- fix/456-prevent-xss-in-slugs
- docs/789-update-api-reference
```

### Branch Protection

```yaml
# Branch protection rules (configured in GitHub)

main:
  # Require pull request reviews
  required_pull_request_reviews:
    required_approving_review_count: 2
    dismiss_stale_reviews: true
    require_code_owner_reviews: true
    require_last_push_approval: true

  # Require status checks
  required_status_checks:
    strict: true
    contexts:
      - "ci / lint"
      - "ci / test"
      - "ci / build"
      - "security / scan"

  # Branch restrictions
  restrictions:
    users: []
    teams:
      - maintainers

  # Additional settings
  enforce_admins: true
  required_linear_history: true
  allow_force_pushes: false
  allow_deletions: false
```

### Merge Strategy

```
# Preferred: Squash and merge for feature branches
git checkout main
git merge --squash feature/123-add-feature
git commit -m "feat: add custom domain support (#123)"

# For release branches: Regular merge
git checkout main
git merge release/v1.2.0 --no-ff
```

---

## CODEOWNERS and Protected Branches

### CODEOWNERS Configuration

```
# .github/CODEOWNERS

# Default owners for everything
* @link-rift/core-team

# API specifications require API team review
/api/ @link-rift/api-team @link-rift/core-team

# Database changes require DBA review
/internal/database/ @link-rift/dba-team @link-rift/core-team
/internal/database/migrations/ @link-rift/dba-team

# Security-sensitive code
/internal/auth/ @link-rift/security-team
/internal/license/ @link-rift/security-team
/pkg/validator/ @link-rift/security-team

# Frontend code
/web/ @link-rift/frontend-team

# Infrastructure and deployment
/deploy/ @link-rift/platform-team
/.github/workflows/ @link-rift/platform-team

# Documentation
/docs/ @link-rift/docs-team
*.md @link-rift/docs-team

# Dependency updates
go.mod @link-rift/core-team @link-rift/security-team
go.sum @link-rift/core-team @link-rift/security-team
/web/package.json @link-rift/frontend-team @link-rift/security-team
/web/package-lock.json @link-rift/frontend-team @link-rift/security-team
```

### Branch Protection Rules

```yaml
# .github/branch-protection.yml (reference configuration)

branches:
  main:
    protection:
      required_pull_request_reviews:
        required_approving_review_count: 2
        dismiss_stale_reviews: true
        require_code_owner_reviews: true
      required_status_checks:
        strict: true
        contexts:
          - "CI / Lint"
          - "CI / Test (ubuntu-latest, 1.22)"
          - "CI / Test (ubuntu-latest, 1.21)"
          - "CI / Build"
          - "Security / Trivy Scan"
          - "Security / CodeQL"
      enforce_admins: true
      required_linear_history: true
      restrictions:
        teams:
          - maintainers

  "release/*":
    protection:
      required_pull_request_reviews:
        required_approving_review_count: 1
        require_code_owner_reviews: true
      required_status_checks:
        strict: true
        contexts:
          - "CI / Lint"
          - "CI / Test (ubuntu-latest, 1.22)"
          - "CI / Build"
```

**GitHub Actions Workflow for Branch Protection:**

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

  test:
    name: Test
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest]
        go-version: ['1.21', '1.22']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      - name: Run tests
        run: go test -race -coverprofile=coverage.out ./...
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Build binaries
        run: make build
      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: binaries
          path: bin/
```

---

## Release Process

### Versioning

Linkrift follows **Semantic Versioning 2.0.0**:

```
MAJOR.MINOR.PATCH

Examples:
- 1.0.0  - Initial stable release
- 1.1.0  - New features, backward compatible
- 1.1.1  - Bug fixes only
- 2.0.0  - Breaking changes
```

**Pre-release versions:**

```
1.2.0-alpha.1    # Early testing
1.2.0-beta.1     # Feature complete, testing
1.2.0-rc.1       # Release candidate
```

### Release Workflow

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write
  packages: write

jobs:
  release:
    name: Create Release
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  docker:
    name: Build and Push Docker Images
    runs-on: ubuntu-latest
    needs: release
    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ghcr.io/link-rift/link-rift
          tags: |
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=semver,pattern={{major}}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          file: deploy/docker/Dockerfile
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          platforms: linux/amd64,linux/arm64

  helm:
    name: Publish Helm Chart
    runs-on: ubuntu-latest
    needs: docker
    steps:
      - uses: actions/checkout@v4

      - name: Install Helm
        uses: azure/setup-helm@v3

      - name: Package Helm chart
        run: |
          helm package deploy/helm/linkrift --version ${{ github.ref_name }}

      - name: Push to Helm registry
        run: |
          helm push linkrift-*.tgz oci://ghcr.io/link-rift/charts
```

**GoReleaser Configuration:**

```yaml
# .goreleaser.yml
project_name: linkrift

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - id: linkrift
    main: ./cmd/linkrift
    binary: linkrift
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

  - id: linkrift-worker
    main: ./cmd/worker
    binary: linkrift-worker
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64

archives:
  - id: default
    builds:
      - linkrift
      - linkrift-worker
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip
    files:
      - LICENSE
      - README.md
      - configs/config.example.yaml

checksum:
  name_template: 'checksums.txt'
  algorithm: sha256

changelog:
  sort: asc
  use: github
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'
      - Merge pull request
      - Merge branch
  groups:
    - title: 'New Features'
      regexp: '^feat'
      order: 0
    - title: 'Bug Fixes'
      regexp: '^fix'
      order: 1
    - title: 'Performance'
      regexp: '^perf'
      order: 2
    - title: 'Security'
      regexp: '^security'
      order: 3

release:
  github:
    owner: link-rift
    name: link-rift
  draft: false
  prerelease: auto
  name_template: "v{{.Version}}"
```

### Changelog Generation

```markdown
<!-- CHANGELOG.md -->
# Changelog

All notable changes to Linkrift will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Feature in progress

## [1.2.0] - 2025-01-15

### Added
- Custom domain support with automatic SSL provisioning
- QR code generation for short links
- Webhook notifications for link events

### Changed
- Improved analytics dashboard performance
- Updated rate limiting algorithm

### Fixed
- Fixed redirect loop when using custom slugs with special characters
- Resolved memory leak in long-running worker processes

### Security
- Updated dependencies to patch CVE-2025-XXXXX

## [1.1.0] - 2024-12-01

### Added
- Team workspaces
- Link expiration settings
- Password-protected links

### Changed
- Migrated to PostgreSQL 16

[Unreleased]: https://github.com/link-rift/link-rift/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/link-rift/link-rift/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/link-rift/link-rift/releases/tag/v1.1.0
```

---

## Related Documentation

- [Open Core Model](./OPEN_CORE_MODEL.md) - Business model and feature tiers
- [License System](./LICENSE_SYSTEM.md) - License verification implementation
- [Contributing Guide](../../CONTRIBUTING.md) - How to contribute
- [API Documentation](../api/README.md) - REST API reference
