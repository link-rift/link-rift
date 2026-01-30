.PHONY: help dev dev-api dev-redirect dev-worker dev-web \
	build build-api build-redirect build-worker build-cli build-web \
	test test-cover test-race bench \
	lint fmt vet security \
	migrate-up migrate-down migrate-create sqlc \
	docker-up docker-down docker-build \
	clean

# Variables
BINARY_DIR := bin
GO := go
GOFLAGS := -trimpath
LDFLAGS := -s -w
AIR := air
MIGRATE := migrate
SQLC := sqlc

# Build version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS += -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

## help: Show this help message
help:
	@echo "Linkrift Development Commands"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

# ──────────────────────────────────────────────
# Development
# ──────────────────────────────────────────────

## dev: Run all services with hot reload
dev:
	@echo "Starting all services..."
	$(MAKE) -j4 dev-api dev-redirect dev-worker dev-web

## dev-api: Run API server with hot reload (air)
dev-api:
	cd cmd/api && $(AIR)

## dev-redirect: Run redirect service with hot reload
dev-redirect:
	cd cmd/redirect && $(AIR)

## dev-worker: Run worker with hot reload
dev-worker:
	cd cmd/worker && $(AIR)

## dev-web: Run Vite dev server
dev-web:
	cd web && npm run dev

# ──────────────────────────────────────────────
# Build
# ──────────────────────────────────────────────

## build: Build all Go binaries
build: build-api build-redirect build-worker build-cli

## build-api: Build API server
build-api:
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BINARY_DIR)/api ./cmd/api

## build-redirect: Build redirect service
build-redirect:
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BINARY_DIR)/redirect ./cmd/redirect

## build-worker: Build background worker
build-worker:
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BINARY_DIR)/worker ./cmd/worker

## build-cli: Build CLI tool
build-cli:
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BINARY_DIR)/linkrift ./cmd/cli

## build-web: Build frontend for production
build-web:
	cd web && npm run build

# ──────────────────────────────────────────────
# Testing
# ──────────────────────────────────────────────

## test: Run all Go tests
test:
	$(GO) test ./... -v

## test-cover: Run tests with coverage report
test-cover:
	$(GO) test ./... -coverprofile=coverage.out -covermode=atomic
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## test-race: Run tests with race detector
test-race:
	$(GO) test ./... -race -v

## bench: Run benchmark tests
bench:
	$(GO) test ./... -bench=. -benchmem -run=^$

# ──────────────────────────────────────────────
# Code Quality
# ──────────────────────────────────────────────

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format Go code
fmt:
	$(GO) fmt ./...
	goimports -w .

## vet: Run go vet
vet:
	$(GO) vet ./...

## security: Run security scanners
security:
	gosec ./...
	govulncheck ./...

# ──────────────────────────────────────────────
# Database
# ──────────────────────────────────────────────

## migrate-up: Run pending migrations
migrate-up:
	$(MIGRATE) -path migrations/postgres -database "$(DATABASE_URL)" up

## migrate-down: Rollback last migration
migrate-down:
	$(MIGRATE) -path migrations/postgres -database "$(DATABASE_URL)" down 1

## migrate-create: Create new migration (usage: make migrate-create name=<name>)
migrate-create:
	$(MIGRATE) create -ext sql -dir migrations/postgres -seq $(name)

## sqlc: Generate Go code from SQL queries
sqlc:
	$(SQLC) generate

# ──────────────────────────────────────────────
# Docker
# ──────────────────────────────────────────────

## docker-up: Start Docker Compose (development)
docker-up:
	docker compose up -d

## docker-down: Stop Docker Compose
docker-down:
	docker compose down

## docker-build: Build Docker images
docker-build:
	docker compose build

# ──────────────────────────────────────────────
# Cleanup
# ──────────────────────────────────────────────

## clean: Remove build artifacts
clean:
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html
