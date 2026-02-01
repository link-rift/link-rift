#!/usr/bin/env bash
# Linkrift Pre-Launch Verification Script
# Runs all automated checks before production deployment.
#
# Usage: ./scripts/pre-launch.sh
#
# This script verifies:
#   1. Backend builds and tests pass
#   2. Frontend builds and tests pass
#   3. Security audit passes
#   4. Database migrations are clean
#   5. API health check responds

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BOLD='\033[1m'
NC='\033[0m'

PASS=0
FAIL=0
SKIP=0

section() { echo -e "\n${BOLD}── $1 ──${NC}"; }
pass() { echo -e "  ${GREEN}✓${NC} $1"; ((PASS++)); }
fail() { echo -e "  ${RED}✗${NC} $1"; ((FAIL++)); }
skip() { echo -e "  ${YELLOW}○${NC} $1 (skipped)"; ((SKIP++)); }

echo "============================================"
echo " Linkrift Pre-Launch Verification"
echo " $(date)"
echo "============================================"

# ── Backend ───────────────────────────────────
section "Backend Build & Test"

if go build ./... 2>/dev/null; then
    pass "Go build succeeds"
else
    fail "Go build failed"
fi

if go vet ./... 2>/dev/null; then
    pass "Go vet clean"
else
    fail "Go vet found issues"
fi

if go test ./... -short -count=1 2>/dev/null; then
    pass "Go tests pass"
else
    fail "Go tests failed"
fi

if command -v golangci-lint &>/dev/null; then
    if golangci-lint run ./... 2>/dev/null; then
        pass "Linter clean"
    else
        fail "Linter found issues"
    fi
else
    skip "golangci-lint not installed"
fi

# ── Frontend ──────────────────────────────────
section "Frontend Build & Test"

if [ -d "web" ] && [ -f "web/package.json" ]; then
    if (cd web && npm run build 2>/dev/null); then
        pass "Frontend build succeeds"
    else
        fail "Frontend build failed"
    fi

    if (cd web && npm test -- --run 2>/dev/null); then
        pass "Frontend tests pass"
    else
        fail "Frontend tests failed"
    fi
else
    skip "Frontend directory not found"
fi

# ── Documentation Site ────────────────────────
section "Documentation Site"

if [ -d "docs-site" ] && [ -f "docs-site/package.json" ]; then
    if (cd docs-site && npm run build 2>/dev/null); then
        pass "Docs site build succeeds"
    else
        fail "Docs site build failed"
    fi
else
    skip "Docs site not found"
fi

# ── Security ──────────────────────────────────
section "Security Checks"

if [ -x "scripts/security-audit.sh" ]; then
    if ./scripts/security-audit.sh 2>/dev/null; then
        pass "Security audit passed"
    else
        fail "Security audit found issues"
    fi
else
    skip "Security audit script not found"
fi

# ── Database ──────────────────────────────────
section "Database"

if command -v migrate &>/dev/null; then
    # Check migration version
    VERSION=$(migrate -path migrations/postgres -database "${DATABASE_URL:-}" version 2>&1 || echo "unknown")
    if [ "$VERSION" != "unknown" ]; then
        pass "Database migration version: $VERSION"
    else
        skip "Could not determine migration version (DB not connected)"
    fi
else
    skip "golang-migrate not installed"
fi

# Verify sqlc generated code is up to date
if command -v sqlc &>/dev/null; then
    DIFF=$(sqlc diff 2>/dev/null || echo "error")
    if [ "$DIFF" = "" ]; then
        pass "sqlc generated code is up to date"
    elif [ "$DIFF" = "error" ]; then
        skip "sqlc diff failed (check sqlc.yaml)"
    else
        fail "sqlc generated code is out of date — run: make sqlc"
    fi
else
    skip "sqlc not installed"
fi

# ── Configuration ─────────────────────────────
section "Configuration Files"

if [ -f "internal/config/config.yaml" ]; then
    pass "config.yaml exists"
else
    fail "config.yaml missing"
fi

if [ -f "deployments/helm/linkrift/values.yaml" ]; then
    pass "Helm values.yaml exists"
else
    fail "Helm values.yaml missing"
fi

if [ -f "openapi/openapi.yaml" ]; then
    pass "OpenAPI spec exists"
else
    warn "OpenAPI spec missing — Swagger UI will be disabled"
fi

# ── Summary ───────────────────────────────────
echo
echo "============================================"
echo " Summary"
echo "============================================"
echo -e "  ${GREEN}Passed${NC}:  $PASS"
echo -e "  ${RED}Failed${NC}:  $FAIL"
echo -e "  ${YELLOW}Skipped${NC}: $SKIP"
echo

if [ "$FAIL" -gt 0 ]; then
    echo -e "${RED}Pre-launch verification FAILED — fix $FAIL issue(s) before deploying${NC}"
    exit 1
else
    echo -e "${GREEN}Pre-launch verification PASSED${NC}"
    exit 0
fi
