# Linkrift CI/CD Pipeline Guide

**Last Updated: 2025-01-24**

This guide covers the complete CI/CD setup for Linkrift using GitHub Actions, including pipeline stages, Go and frontend CI, Docker builds, and environment deployments.

---

## Table of Contents

1. [Pipeline Overview](#pipeline-overview)
2. [Pipeline Stages](#pipeline-stages)
3. [Go CI Pipeline](#go-ci-pipeline)
4. [Frontend CI Pipeline](#frontend-ci-pipeline)
5. [Docker Image Building](#docker-image-building)
6. [Environment Deployments](#environment-deployments)
7. [Complete Workflow Examples](#complete-workflow-examples)
8. [Secrets and Variables](#secrets-and-variables)
9. [Best Practices](#best-practices)

---

## Pipeline Overview

### CI/CD Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           GitHub Actions CI/CD                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Pull Request                    Main Branch                 Release Tag     │
│       │                              │                            │          │
│       ▼                              ▼                            ▼          │
│  ┌─────────┐                   ┌─────────┐                  ┌─────────┐     │
│  │  Lint   │                   │  Lint   │                  │  Lint   │     │
│  │  Test   │                   │  Test   │                  │  Test   │     │
│  │  Build  │                   │  Build  │                  │  Build  │     │
│  └────┬────┘                   └────┬────┘                  └────┬────┘     │
│       │                              │                            │          │
│       ▼                              ▼                            ▼          │
│  ┌─────────┐                   ┌─────────┐                  ┌─────────┐     │
│  │ Security│                   │ Security│                  │ Security│     │
│  │  Scan   │                   │  Scan   │                  │  Scan   │     │
│  └────┬────┘                   └────┬────┘                  └────┬────┘     │
│       │                              │                            │          │
│       ▼                              ▼                            ▼          │
│  ┌─────────┐                   ┌─────────┐                  ┌─────────┐     │
│  │  Done   │                   │ Docker  │                  │ Docker  │     │
│  └─────────┘                   │  Build  │                  │  Build  │     │
│                                └────┬────┘                  └────┬────┘     │
│                                     │                            │          │
│                                     ▼                            ▼          │
│                                ┌─────────┐                  ┌─────────┐     │
│                                │ Deploy  │                  │ Deploy  │     │
│                                │ Staging │                  │  Prod   │     │
│                                └─────────┘                  └─────────┘     │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Workflow Files Structure

```
.github/
├── workflows/
│   ├── ci.yml                    # Main CI pipeline
│   ├── cd-staging.yml            # Staging deployment
│   ├── cd-production.yml         # Production deployment
│   ├── docker-build.yml          # Docker image builds
│   ├── security-scan.yml         # Security scanning
│   ├── dependency-update.yml     # Dependabot auto-merge
│   └── release.yml               # Release automation
├── actions/
│   ├── setup-go/
│   │   └── action.yml
│   ├── setup-node/
│   │   └── action.yml
│   └── notify/
│       └── action.yml
└── CODEOWNERS
```

---

## Pipeline Stages

### Stage Overview

| Stage | Purpose | Duration | Trigger |
|-------|---------|----------|---------|
| **Lint** | Code quality checks | ~1 min | All pushes |
| **Test** | Unit and integration tests | ~3 min | All pushes |
| **Build** | Compile binaries | ~2 min | All pushes |
| **Security** | Vulnerability scanning | ~2 min | All pushes |
| **Docker** | Build container images | ~5 min | Main/tags |
| **Deploy** | Environment deployment | ~3 min | Main/tags |

### Parallel Execution Strategy

```yaml
# Jobs run in parallel when possible
jobs:
  lint:          # Runs immediately
  test:          # Runs immediately (parallel with lint)
  build:         # Runs immediately (parallel with lint, test)
  security:      # Waits for lint, test, build
  docker:        # Waits for security
  deploy:        # Waits for docker
```

---

## Go CI Pipeline

### Complete Go CI Workflow

```yaml
# .github/workflows/ci-go.yml

name: Go CI

on:
  push:
    branches: [main, develop]
    paths:
      - '**.go'
      - 'go.mod'
      - 'go.sum'
      - '.github/workflows/ci-go.yml'
  pull_request:
    branches: [main, develop]
    paths:
      - '**.go'
      - 'go.mod'
      - 'go.sum'

env:
  GO_VERSION: '1.22'
  GOLANGCI_LINT_VERSION: 'v1.55.2'

jobs:
  # ===========================================
  # Lint Job
  # ===========================================
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: ${{ env.GOLANGCI_LINT_VERSION }}
          args: --timeout=5m

      - name: Check go mod tidy
        run: |
          go mod tidy
          git diff --exit-code go.mod go.sum

      - name: Check formatting
        run: |
          gofmt -l .
          test -z "$(gofmt -l .)"

  # ===========================================
  # Test Job
  # ===========================================
  test:
    name: Test
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: linkrift_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run unit tests
        run: |
          go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
        env:
          LINKRIFT_DB_HOST: localhost
          LINKRIFT_DB_PORT: 5432
          LINKRIFT_DB_USER: test
          LINKRIFT_DB_PASSWORD: test
          LINKRIFT_DB_NAME: linkrift_test
          LINKRIFT_REDIS_URL: redis://localhost:6379/0

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          token: ${{ secrets.CODECOV_TOKEN }}
          files: ./coverage.out
          flags: unittests
          name: codecov-umbrella
          fail_ci_if_error: false

      - name: Run integration tests
        run: |
          go test -v -tags=integration -race ./...
        env:
          LINKRIFT_DB_HOST: localhost
          LINKRIFT_DB_PORT: 5432
          LINKRIFT_DB_USER: test
          LINKRIFT_DB_PASSWORD: test
          LINKRIFT_DB_NAME: linkrift_test
          LINKRIFT_REDIS_URL: redis://localhost:6379/0

  # ===========================================
  # Build Job
  # ===========================================
  build:
    name: Build
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [api, redirect, worker]
        os: [linux]
        arch: [amd64, arm64]
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Get version info
        id: version
        run: |
          echo "version=$(git describe --tags --always --dirty)" >> $GITHUB_OUTPUT
          echo "commit=$(git rev-parse --short HEAD)" >> $GITHUB_OUTPUT
          echo "build_time=$(date -u +%Y-%m-%dT%H:%M:%SZ)" >> $GITHUB_OUTPUT

      - name: Build binary
        run: |
          CGO_ENABLED=0 GOOS=${{ matrix.os }} GOARCH=${{ matrix.arch }} go build \
            -ldflags="-s -w \
              -X main.Version=${{ steps.version.outputs.version }} \
              -X main.Commit=${{ steps.version.outputs.commit }} \
              -X main.BuildTime=${{ steps.version.outputs.build_time }}" \
            -o build/linkrift-${{ matrix.service }}-${{ matrix.os }}-${{ matrix.arch }} \
            ./cmd/${{ matrix.service }}

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: linkrift-${{ matrix.service }}-${{ matrix.os }}-${{ matrix.arch }}
          path: build/linkrift-${{ matrix.service }}-${{ matrix.os }}-${{ matrix.arch }}
          retention-days: 7

  # ===========================================
  # Security Scan Job
  # ===========================================
  security:
    name: Security Scan
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run Gosec
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt sarif -out gosec-results.sarif ./...'

      - name: Upload SARIF file
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: gosec-results.sarif

      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

      - name: Run Nancy (dependency check)
        run: |
          go install github.com/sonatype-nexus-community/nancy@latest
          go list -json -deps ./... | nancy sleuth

  # ===========================================
  # Summary Job
  # ===========================================
  ci-success:
    name: CI Success
    runs-on: ubuntu-latest
    needs: [lint, test, build, security]
    if: always()
    steps:
      - name: Check all jobs
        run: |
          if [[ "${{ needs.lint.result }}" != "success" ]] || \
             [[ "${{ needs.test.result }}" != "success" ]] || \
             [[ "${{ needs.build.result }}" != "success" ]] || \
             [[ "${{ needs.security.result }}" != "success" ]]; then
            echo "One or more jobs failed"
            exit 1
          fi
          echo "All CI jobs passed successfully!"
```

### golangci-lint Configuration

```yaml
# .golangci.yml

run:
  timeout: 5m
  issues-exit-code: 1
  tests: true
  skip-dirs:
    - vendor
    - mocks

output:
  format: colored-line-number
  print-issued-lines: true
  print-linter-name: true

linters:
  enable:
    - bodyclose
    - dogsled
    - dupl
    - errcheck
    - exportloopref
    - exhaustive
    - gochecknoinits
    - goconst
    - gocritic
    - gocyclo
    - gofmt
    - goimports
    - gomnd
    - goprintffuncname
    - gosec
    - gosimple
    - govet
    - ineffassign
    - misspell
    - nakedret
    - noctx
    - nolintlint
    - prealloc
    - revive
    - staticcheck
    - stylecheck
    - typecheck
    - unconvert
    - unparam
    - unused
    - whitespace

linters-settings:
  dupl:
    threshold: 100
  errcheck:
    check-type-assertions: true
    check-blank: true
  gocyclo:
    min-complexity: 15
  govet:
    check-shadowing: true
  misspell:
    locale: US
  goconst:
    min-len: 3
    min-occurrences: 3

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - dupl
        - gomnd
    - path: cmd/
      linters:
        - gochecknoinits
```

---

## Frontend CI Pipeline

### Complete Frontend CI Workflow

```yaml
# .github/workflows/ci-frontend.yml

name: Frontend CI

on:
  push:
    branches: [main, develop]
    paths:
      - 'web/**'
      - '.github/workflows/ci-frontend.yml'
  pull_request:
    branches: [main, develop]
    paths:
      - 'web/**'

env:
  NODE_VERSION: '20'
  PNPM_VERSION: '8'

defaults:
  run:
    working-directory: web

jobs:
  # ===========================================
  # Lint Job
  # ===========================================
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup pnpm
        uses: pnpm/action-setup@v3
        with:
          version: ${{ env.PNPM_VERSION }}

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'pnpm'
          cache-dependency-path: web/pnpm-lock.yaml

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Run ESLint
        run: pnpm lint

      - name: Run Prettier check
        run: pnpm format:check

      - name: Run TypeScript check
        run: pnpm tsc --noEmit

  # ===========================================
  # Test Job
  # ===========================================
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup pnpm
        uses: pnpm/action-setup@v3
        with:
          version: ${{ env.PNPM_VERSION }}

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'pnpm'
          cache-dependency-path: web/pnpm-lock.yaml

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Run unit tests
        run: pnpm test:coverage

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          token: ${{ secrets.CODECOV_TOKEN }}
          files: ./web/coverage/lcov.info
          flags: frontend
          name: frontend-coverage

  # ===========================================
  # Build Job
  # ===========================================
  build:
    name: Build
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup pnpm
        uses: pnpm/action-setup@v3
        with:
          version: ${{ env.PNPM_VERSION }}

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'pnpm'
          cache-dependency-path: web/pnpm-lock.yaml

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Build
        run: pnpm build
        env:
          VITE_API_URL: https://api.linkrift.io
          VITE_APP_VERSION: ${{ github.sha }}

      - name: Upload build artifact
        uses: actions/upload-artifact@v4
        with:
          name: frontend-build
          path: web/dist
          retention-days: 7

      - name: Analyze bundle size
        run: |
          echo "## Bundle Size" >> $GITHUB_STEP_SUMMARY
          echo "" >> $GITHUB_STEP_SUMMARY
          du -sh dist >> $GITHUB_STEP_SUMMARY
          echo "" >> $GITHUB_STEP_SUMMARY
          echo "### JavaScript Files" >> $GITHUB_STEP_SUMMARY
          find dist -name "*.js" -exec du -h {} \; | sort -rh | head -10 >> $GITHUB_STEP_SUMMARY

  # ===========================================
  # E2E Tests Job
  # ===========================================
  e2e:
    name: E2E Tests
    runs-on: ubuntu-latest
    needs: [build]
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup pnpm
        uses: pnpm/action-setup@v3
        with:
          version: ${{ env.PNPM_VERSION }}

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'pnpm'
          cache-dependency-path: web/pnpm-lock.yaml

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Install Playwright
        run: pnpm exec playwright install --with-deps chromium

      - name: Download build artifact
        uses: actions/download-artifact@v4
        with:
          name: frontend-build
          path: web/dist

      - name: Run E2E tests
        run: pnpm test:e2e
        env:
          PLAYWRIGHT_BASE_URL: http://localhost:4173

      - name: Upload test results
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: playwright-report
          path: web/playwright-report
          retention-days: 7

  # ===========================================
  # Lighthouse Job
  # ===========================================
  lighthouse:
    name: Lighthouse
    runs-on: ubuntu-latest
    needs: [build]
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Download build artifact
        uses: actions/download-artifact@v4
        with:
          name: frontend-build
          path: web/dist

      - name: Run Lighthouse
        uses: treosh/lighthouse-ci-action@v11
        with:
          configPath: ./web/lighthouserc.json
          uploadArtifacts: true
          temporaryPublicStorage: true

  # ===========================================
  # Summary Job
  # ===========================================
  frontend-ci-success:
    name: Frontend CI Success
    runs-on: ubuntu-latest
    needs: [lint, test, build, e2e]
    if: always()
    steps:
      - name: Check all jobs
        run: |
          if [[ "${{ needs.lint.result }}" != "success" ]] || \
             [[ "${{ needs.test.result }}" != "success" ]] || \
             [[ "${{ needs.build.result }}" != "success" ]] || \
             [[ "${{ needs.e2e.result }}" != "success" ]]; then
            echo "One or more jobs failed"
            exit 1
          fi
          echo "All frontend CI jobs passed successfully!"
```

### Lighthouse Configuration

```json
// web/lighthouserc.json
{
  "ci": {
    "collect": {
      "staticDistDir": "./dist",
      "url": [
        "http://localhost/",
        "http://localhost/login",
        "http://localhost/dashboard"
      ],
      "numberOfRuns": 3
    },
    "assert": {
      "assertions": {
        "categories:performance": ["warn", { "minScore": 0.9 }],
        "categories:accessibility": ["error", { "minScore": 0.9 }],
        "categories:best-practices": ["warn", { "minScore": 0.9 }],
        "categories:seo": ["warn", { "minScore": 0.9 }],
        "first-contentful-paint": ["warn", { "maxNumericValue": 2000 }],
        "largest-contentful-paint": ["warn", { "maxNumericValue": 2500 }],
        "cumulative-layout-shift": ["warn", { "maxNumericValue": 0.1 }],
        "total-blocking-time": ["warn", { "maxNumericValue": 300 }]
      }
    },
    "upload": {
      "target": "temporary-public-storage"
    }
  }
}
```

---

## Docker Image Building

### Docker Build Workflow

```yaml
# .github/workflows/docker-build.yml

name: Docker Build

on:
  push:
    branches: [main]
    tags: ['v*']
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_PREFIX: ${{ github.repository }}

jobs:
  # ===========================================
  # Build Go Services
  # ===========================================
  build-go-services:
    name: Build ${{ matrix.service }}
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    strategy:
      matrix:
        service: [api, redirect, worker]
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up QEMU
        uses: docker/setup-qemu-action@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Container Registry
        if: github.event_name != 'pull_request'
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-${{ matrix.service }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          file: docker/${{ matrix.service }}/Dockerfile
          platforms: linux/amd64,linux/arm64
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            VERSION=${{ github.ref_name }}
            COMMIT=${{ github.sha }}
            BUILD_TIME=${{ github.event.head_commit.timestamp }}

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-${{ matrix.service }}:${{ steps.meta.outputs.version }}
          format: 'sarif'
          output: 'trivy-results-${{ matrix.service }}.sarif'

      - name: Upload Trivy scan results
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: trivy-results-${{ matrix.service }}.sarif

  # ===========================================
  # Build Frontend
  # ===========================================
  build-frontend:
    name: Build Web
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Container Registry
        if: github.event_name != 'pull_request'
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-web
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          file: docker/web/Dockerfile
          platforms: linux/amd64,linux/arm64
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            VITE_API_URL=https://api.linkrift.io
            VITE_APP_VERSION=${{ github.ref_name }}

  # ===========================================
  # Build Migrations
  # ===========================================
  build-migrations:
    name: Build Migrations
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Container Registry
        if: github.event_name != 'pull_request'
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-migrations
          tags: |
            type=ref,event=branch
            type=semver,pattern={{version}}
            type=sha

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          file: docker/migrations/Dockerfile
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

---

## Environment Deployments

### Staging Deployment Workflow

```yaml
# .github/workflows/cd-staging.yml

name: Deploy to Staging

on:
  push:
    branches: [main]
  workflow_dispatch:

concurrency:
  group: staging-deployment
  cancel-in-progress: false

env:
  REGISTRY: ghcr.io
  IMAGE_PREFIX: ${{ github.repository }}

jobs:
  # ===========================================
  # Deploy to Staging
  # ===========================================
  deploy:
    name: Deploy
    runs-on: ubuntu-latest
    environment:
      name: staging
      url: https://staging.linkrift.io
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Get image tag
        id: image
        run: echo "tag=sha-$(git rev-parse --short HEAD)" >> $GITHUB_OUTPUT

      - name: Run database migrations
        run: |
          aws ecs run-task \
            --cluster linkrift-staging \
            --task-definition linkrift-migrations-staging \
            --overrides '{
              "containerOverrides": [{
                "name": "migrations",
                "environment": [
                  {"name": "MIGRATE_COMMAND", "value": "up"}
                ]
              }]
            }' \
            --network-configuration '{
              "awsvpcConfiguration": {
                "subnets": ["${{ secrets.STAGING_SUBNET }}"],
                "securityGroups": ["${{ secrets.STAGING_SG }}"]
              }
            }'

      - name: Deploy API service
        run: |
          aws ecs update-service \
            --cluster linkrift-staging \
            --service linkrift-api-staging \
            --force-new-deployment \
            --task-definition linkrift-api-staging

      - name: Deploy Redirect service
        run: |
          aws ecs update-service \
            --cluster linkrift-staging \
            --service linkrift-redirect-staging \
            --force-new-deployment \
            --task-definition linkrift-redirect-staging

      - name: Deploy Worker service
        run: |
          aws ecs update-service \
            --cluster linkrift-staging \
            --service linkrift-worker-staging \
            --force-new-deployment \
            --task-definition linkrift-worker-staging

      - name: Deploy Web service
        run: |
          aws ecs update-service \
            --cluster linkrift-staging \
            --service linkrift-web-staging \
            --force-new-deployment \
            --task-definition linkrift-web-staging

      - name: Wait for deployment
        run: |
          aws ecs wait services-stable \
            --cluster linkrift-staging \
            --services linkrift-api-staging linkrift-redirect-staging linkrift-web-staging

      - name: Run smoke tests
        run: |
          # Health checks
          curl -sf https://staging.linkrift.io/health
          curl -sf https://api.staging.linkrift.io/health
          curl -sf https://go.staging.linkrift.io/health

      - name: Notify on success
        if: success()
        uses: slackapi/slack-github-action@v1
        with:
          payload: |
            {
              "text": "Staging deployment successful",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*Staging Deployment Successful* :white_check_mark:\n\n*Commit:* `${{ github.sha }}`\n*Branch:* `${{ github.ref_name }}`\n*URL:* https://staging.linkrift.io"
                  }
                }
              ]
            }
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}

      - name: Notify on failure
        if: failure()
        uses: slackapi/slack-github-action@v1
        with:
          payload: |
            {
              "text": "Staging deployment failed",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*Staging Deployment Failed* :x:\n\n*Commit:* `${{ github.sha }}`\n*Branch:* `${{ github.ref_name }}`\n*Run:* ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}"
                  }
                }
              ]
            }
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

### Production Deployment Workflow

```yaml
# .github/workflows/cd-production.yml

name: Deploy to Production

on:
  push:
    tags: ['v*']
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to deploy (e.g., v1.0.0)'
        required: true

concurrency:
  group: production-deployment
  cancel-in-progress: false

env:
  REGISTRY: ghcr.io
  IMAGE_PREFIX: ${{ github.repository }}

jobs:
  # ===========================================
  # Pre-deployment Checks
  # ===========================================
  pre-deploy:
    name: Pre-deployment Checks
    runs-on: ubuntu-latest
    outputs:
      version: ${{ steps.version.outputs.version }}
    steps:
      - name: Determine version
        id: version
        run: |
          if [ "${{ github.event_name }}" == "workflow_dispatch" ]; then
            echo "version=${{ github.event.inputs.version }}" >> $GITHUB_OUTPUT
          else
            echo "version=${{ github.ref_name }}" >> $GITHUB_OUTPUT
          fi

      - name: Verify image exists
        run: |
          docker manifest inspect ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-api:${{ steps.version.outputs.version }}
          docker manifest inspect ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-redirect:${{ steps.version.outputs.version }}
          docker manifest inspect ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-worker:${{ steps.version.outputs.version }}
          docker manifest inspect ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-web:${{ steps.version.outputs.version }}

  # ===========================================
  # Deploy to Production
  # ===========================================
  deploy:
    name: Deploy
    runs-on: ubuntu-latest
    needs: [pre-deploy]
    environment:
      name: production
      url: https://linkrift.io
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Create deployment record
        id: deployment
        run: |
          DEPLOYMENT_ID=$(uuidgen)
          echo "id=${DEPLOYMENT_ID}" >> $GITHUB_OUTPUT
          aws dynamodb put-item \
            --table-name linkrift-deployments \
            --item '{
              "id": {"S": "'${DEPLOYMENT_ID}'"},
              "version": {"S": "${{ needs.pre-deploy.outputs.version }}"},
              "timestamp": {"S": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"},
              "status": {"S": "in_progress"},
              "commit": {"S": "${{ github.sha }}"},
              "actor": {"S": "${{ github.actor }}"}
            }'

      - name: Create database backup
        run: |
          aws rds create-db-snapshot \
            --db-instance-identifier linkrift-production \
            --db-snapshot-identifier linkrift-pre-deploy-${{ steps.deployment.outputs.id }}

      - name: Run database migrations
        run: |
          aws ecs run-task \
            --cluster linkrift-production \
            --task-definition linkrift-migrations-production \
            --overrides '{
              "containerOverrides": [{
                "name": "migrations",
                "environment": [
                  {"name": "MIGRATE_COMMAND", "value": "up"}
                ]
              }]
            }' \
            --network-configuration '{
              "awsvpcConfiguration": {
                "subnets": ${{ secrets.PRODUCTION_SUBNETS }},
                "securityGroups": ["${{ secrets.PRODUCTION_SG }}"]
              }
            }'

          # Wait for migration to complete
          sleep 30

      - name: Deploy API (canary)
        run: |
          # Deploy to canary (10% traffic)
          aws ecs update-service \
            --cluster linkrift-production \
            --service linkrift-api-canary \
            --task-definition linkrift-api-production:${{ needs.pre-deploy.outputs.version }}

          aws ecs wait services-stable \
            --cluster linkrift-production \
            --services linkrift-api-canary

      - name: Verify canary health
        run: |
          # Wait for metrics
          sleep 60

          # Check error rate
          ERROR_RATE=$(aws cloudwatch get-metric-statistics \
            --namespace Linkrift/API \
            --metric-name ErrorRate \
            --dimensions Name=Service,Value=api-canary \
            --start-time $(date -u -d '5 minutes ago' +%Y-%m-%dT%H:%M:%SZ) \
            --end-time $(date -u +%Y-%m-%dT%H:%M:%SZ) \
            --period 60 \
            --statistics Average \
            --query 'Datapoints[0].Average' \
            --output text)

          if (( $(echo "$ERROR_RATE > 1" | bc -l) )); then
            echo "Canary error rate too high: ${ERROR_RATE}%"
            exit 1
          fi

      - name: Deploy all services
        run: |
          # Deploy API (full)
          aws ecs update-service \
            --cluster linkrift-production \
            --service linkrift-api-production \
            --task-definition linkrift-api-production:${{ needs.pre-deploy.outputs.version }}

          # Deploy Redirect
          aws ecs update-service \
            --cluster linkrift-production \
            --service linkrift-redirect-production \
            --task-definition linkrift-redirect-production:${{ needs.pre-deploy.outputs.version }}

          # Deploy Worker
          aws ecs update-service \
            --cluster linkrift-production \
            --service linkrift-worker-production \
            --task-definition linkrift-worker-production:${{ needs.pre-deploy.outputs.version }}

          # Deploy Web
          aws ecs update-service \
            --cluster linkrift-production \
            --service linkrift-web-production \
            --task-definition linkrift-web-production:${{ needs.pre-deploy.outputs.version }}

      - name: Wait for deployment
        run: |
          aws ecs wait services-stable \
            --cluster linkrift-production \
            --services \
              linkrift-api-production \
              linkrift-redirect-production \
              linkrift-worker-production \
              linkrift-web-production

      - name: Run smoke tests
        run: |
          # Health checks
          curl -sf https://linkrift.io/health
          curl -sf https://api.linkrift.io/health
          curl -sf https://go.linkrift.io/health

          # Functional tests
          # Create a test link
          RESPONSE=$(curl -sf -X POST https://api.linkrift.io/v1/links \
            -H "Authorization: Bearer ${{ secrets.SMOKE_TEST_TOKEN }}" \
            -H "Content-Type: application/json" \
            -d '{"url": "https://example.com", "title": "Smoke Test"}')

          SHORT_CODE=$(echo $RESPONSE | jq -r '.short_code')

          # Test redirect
          curl -sf -o /dev/null -w "%{http_code}" "https://go.linkrift.io/${SHORT_CODE}" | grep -q "301\|302"

          # Delete test link
          curl -sf -X DELETE "https://api.linkrift.io/v1/links/${SHORT_CODE}" \
            -H "Authorization: Bearer ${{ secrets.SMOKE_TEST_TOKEN }}"

      - name: Update deployment record
        if: always()
        run: |
          STATUS="${{ job.status == 'success' && 'success' || 'failed' }}"
          aws dynamodb update-item \
            --table-name linkrift-deployments \
            --key '{"id": {"S": "${{ steps.deployment.outputs.id }}"}}' \
            --update-expression "SET #status = :status, completed_at = :completed" \
            --expression-attribute-names '{"#status": "status"}' \
            --expression-attribute-values '{
              ":status": {"S": "'${STATUS}'"},
              ":completed": {"S": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}
            }'

      - name: Notify on success
        if: success()
        uses: slackapi/slack-github-action@v1
        with:
          payload: |
            {
              "text": "Production deployment successful",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*Production Deployment Successful* :rocket:\n\n*Version:* `${{ needs.pre-deploy.outputs.version }}`\n*Deployed by:* ${{ github.actor }}\n*URL:* https://linkrift.io"
                  }
                }
              ]
            }
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}

      - name: Notify on failure
        if: failure()
        uses: slackapi/slack-github-action@v1
        with:
          payload: |
            {
              "text": "Production deployment failed - ROLLBACK MAY BE REQUIRED",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*Production Deployment Failed* :rotating_light:\n\n*Version:* `${{ needs.pre-deploy.outputs.version }}`\n*Deployed by:* ${{ github.actor }}\n*Run:* ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}\n\n*Action Required:* Review logs and consider rollback"
                  }
                }
              ]
            }
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}

  # ===========================================
  # Post-deployment Verification
  # ===========================================
  verify:
    name: Verify Deployment
    runs-on: ubuntu-latest
    needs: [deploy]
    steps:
      - name: Wait for traffic
        run: sleep 300  # Wait 5 minutes for traffic

      - name: Check error rates
        run: |
          # Query Prometheus/CloudWatch for error rates
          # Alert if above threshold

      - name: Check latency
        run: |
          # Query Prometheus/CloudWatch for P99 latency
          # Alert if above threshold

      - name: Create GitHub release
        uses: softprops/action-gh-release@v1
        with:
          tag_name: ${{ needs.pre-deploy.outputs.version }}
          name: Release ${{ needs.pre-deploy.outputs.version }}
          generate_release_notes: true
```

---

## Complete Workflow Examples

### Full CI/CD Pipeline

```yaml
# .github/workflows/ci-cd.yml

name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
    tags: ['v*']
  pull_request:
    branches: [main, develop]

env:
  GO_VERSION: '1.22'
  NODE_VERSION: '20'
  REGISTRY: ghcr.io

jobs:
  # ===========================================
  # Detect Changes
  # ===========================================
  changes:
    name: Detect Changes
    runs-on: ubuntu-latest
    outputs:
      go: ${{ steps.filter.outputs.go }}
      frontend: ${{ steps.filter.outputs.frontend }}
      docker: ${{ steps.filter.outputs.docker }}
    steps:
      - uses: actions/checkout@v4
      - uses: dorny/paths-filter@v3
        id: filter
        with:
          filters: |
            go:
              - '**.go'
              - 'go.mod'
              - 'go.sum'
            frontend:
              - 'web/**'
            docker:
              - 'docker/**'
              - 'Dockerfile*'

  # ===========================================
  # Go CI
  # ===========================================
  go-ci:
    name: Go CI
    needs: [changes]
    if: needs.changes.outputs.go == 'true'
    uses: ./.github/workflows/ci-go.yml
    secrets: inherit

  # ===========================================
  # Frontend CI
  # ===========================================
  frontend-ci:
    name: Frontend CI
    needs: [changes]
    if: needs.changes.outputs.frontend == 'true'
    uses: ./.github/workflows/ci-frontend.yml
    secrets: inherit

  # ===========================================
  # Docker Build
  # ===========================================
  docker:
    name: Docker Build
    needs: [go-ci, frontend-ci]
    if: |
      always() &&
      (needs.go-ci.result == 'success' || needs.go-ci.result == 'skipped') &&
      (needs.frontend-ci.result == 'success' || needs.frontend-ci.result == 'skipped') &&
      (github.ref == 'refs/heads/main' || startsWith(github.ref, 'refs/tags/'))
    uses: ./.github/workflows/docker-build.yml
    secrets: inherit

  # ===========================================
  # Deploy Staging
  # ===========================================
  deploy-staging:
    name: Deploy Staging
    needs: [docker]
    if: github.ref == 'refs/heads/main'
    uses: ./.github/workflows/cd-staging.yml
    secrets: inherit

  # ===========================================
  # Deploy Production
  # ===========================================
  deploy-production:
    name: Deploy Production
    needs: [docker]
    if: startsWith(github.ref, 'refs/tags/v')
    uses: ./.github/workflows/cd-production.yml
    secrets: inherit
```

---

## Secrets and Variables

### Required Secrets

```yaml
# Repository Secrets (Settings -> Secrets and variables -> Actions)

# Docker Registry
GITHUB_TOKEN                    # Auto-provided by GitHub

# AWS
AWS_ACCESS_KEY_ID               # IAM user access key
AWS_SECRET_ACCESS_KEY           # IAM user secret key

# Database
DB_HOST                         # Database hostname
DB_PASSWORD                     # Database password
DB_PASSWORD_READONLY            # Read-only database password

# Redis
REDIS_URL                       # Redis connection URL

# Application
JWT_SECRET                      # JWT signing secret
SMOKE_TEST_TOKEN                # API token for smoke tests

# Notifications
SLACK_WEBHOOK_URL               # Slack incoming webhook

# Coverage
CODECOV_TOKEN                   # Codecov upload token

# Staging Environment
STAGING_SUBNET                  # Staging VPC subnet
STAGING_SG                      # Staging security group

# Production Environment
PRODUCTION_SUBNETS              # Production VPC subnets (JSON array)
PRODUCTION_SG                   # Production security group
```

### Environment Variables

```yaml
# Repository Variables (Settings -> Secrets and variables -> Actions -> Variables)

GO_VERSION: '1.22'
NODE_VERSION: '20'
PNPM_VERSION: '8'
REGISTRY: 'ghcr.io'
```

### Environment Configuration

```yaml
# Staging Environment
name: staging
url: https://staging.linkrift.io
protection_rules:
  - required_reviewers: 0
  - wait_timer: 0

# Production Environment
name: production
url: https://linkrift.io
protection_rules:
  - required_reviewers: 1
  - wait_timer: 5  # 5 minute delay
  - prevent_self_review: true
```

---

## Best Practices

### 1. Caching Strategies

```yaml
# Go module caching
- uses: actions/setup-go@v5
  with:
    go-version: ${{ env.GO_VERSION }}
    cache: true

# Node modules caching
- uses: actions/setup-node@v4
  with:
    node-version: ${{ env.NODE_VERSION }}
    cache: 'pnpm'
    cache-dependency-path: web/pnpm-lock.yaml

# Docker layer caching
- uses: docker/build-push-action@v5
  with:
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

### 2. Parallel Execution

```yaml
# Run independent jobs in parallel
jobs:
  lint:
    # Runs immediately
  test:
    # Runs immediately (parallel with lint)
  build:
    # Runs immediately (parallel with lint, test)
  security:
    needs: [lint, test]  # Waits for both
```

### 3. Fail Fast

```yaml
strategy:
  fail-fast: true  # Cancel all jobs if one fails
  matrix:
    service: [api, redirect, worker]
```

### 4. Timeout Configuration

```yaml
jobs:
  build:
    timeout-minutes: 15
    steps:
      - name: Long running step
        timeout-minutes: 10
        run: make build
```

### 5. Conditional Execution

```yaml
# Skip on certain conditions
if: |
  !contains(github.event.head_commit.message, '[skip ci]') &&
  !contains(github.event.head_commit.message, '[ci skip]')

# Run only on specific paths
on:
  push:
    paths:
      - '**.go'
      - '!**_test.go'  # Exclude test files
```

### 6. Reusable Workflows

```yaml
# .github/workflows/reusable-deploy.yml
on:
  workflow_call:
    inputs:
      environment:
        required: true
        type: string
    secrets:
      AWS_ACCESS_KEY_ID:
        required: true

# Usage
jobs:
  deploy:
    uses: ./.github/workflows/reusable-deploy.yml
    with:
      environment: staging
    secrets: inherit
```

### 7. Status Checks

```yaml
# Branch protection rules
required_status_checks:
  strict: true
  contexts:
    - "Go CI / Lint"
    - "Go CI / Test"
    - "Go CI / Build"
    - "Go CI / Security Scan"
    - "Frontend CI / Lint"
    - "Frontend CI / Test"
    - "Frontend CI / Build"
```

---

*This CI/CD guide is maintained by the Linkrift Platform Team. For questions or updates, contact platform@linkrift.io*
