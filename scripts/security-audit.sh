#!/usr/bin/env bash
# Linkrift Security Audit Script
# Run before production deployments to verify security posture.
#
# Usage: ./scripts/security-audit.sh
#
# Prerequisites:
#   - Go toolchain installed
#   - gosec: go install github.com/securego/gosec/v2/cmd/gosec@latest
#   - govulncheck: go install golang.org/x/vuln/cmd/govulncheck@latest

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

PASS=0
FAIL=0
WARN=0

pass() { echo -e "  ${GREEN}PASS${NC} $1"; ((PASS++)); }
fail() { echo -e "  ${RED}FAIL${NC} $1"; ((FAIL++)); }
warn() { echo -e "  ${YELLOW}WARN${NC} $1"; ((WARN++)); }

echo "============================================"
echo " Linkrift Security Audit"
echo " $(date)"
echo "============================================"
echo

# ── 1. Dependency Vulnerability Scan ──────────
echo "1. Dependency Vulnerability Scan"
echo "---"

if command -v govulncheck &>/dev/null; then
    if govulncheck ./... 2>/dev/null; then
        pass "No known vulnerabilities in dependencies"
    else
        fail "Vulnerabilities found — run: govulncheck ./..."
    fi
else
    warn "govulncheck not installed — install: go install golang.org/x/vuln/cmd/govulncheck@latest"
fi
echo

# ── 2. Static Analysis (gosec) ────────────────
echo "2. Static Security Analysis"
echo "---"

if command -v gosec &>/dev/null; then
    GOSEC_OUTPUT=$(gosec -quiet -fmt text ./... 2>&1 || true)
    ISSUE_COUNT=$(echo "$GOSEC_OUTPUT" | grep -c "^\[" || true)
    if [ "$ISSUE_COUNT" -eq 0 ]; then
        pass "No security issues found by gosec"
    else
        fail "gosec found $ISSUE_COUNT issues — run: gosec ./..."
    fi
else
    warn "gosec not installed — install: go install github.com/securego/gosec/v2/cmd/gosec@latest"
fi
echo

# ── 3. Hardcoded Secrets Check ────────────────
echo "3. Hardcoded Secrets Check"
echo "---"

SECRETS_FOUND=0
# Check for common secret patterns in Go source files (excluding tests and vendor)
for pattern in "password\s*=\s*\"[^\"]\+" "secret\s*=\s*\"[^\"]\+" "api_key\s*=\s*\"[^\"]\+" "token\s*=\s*\"[a-zA-Z0-9]\{20,\}\""; do
    MATCHES=$(grep -rl "$pattern" --include="*.go" --exclude-dir=vendor --exclude="*_test.go" . 2>/dev/null || true)
    if [ -n "$MATCHES" ]; then
        ((SECRETS_FOUND++))
    fi
done

if [ "$SECRETS_FOUND" -eq 0 ]; then
    pass "No hardcoded secrets detected in Go source"
else
    warn "Potential hardcoded secrets found — review manually"
fi

# Check for .env files that shouldn't be committed
if [ -f .env ]; then
    fail ".env file exists in repository root — should be in .gitignore"
else
    pass "No .env file in repository root"
fi
echo

# ── 4. SQL Injection Prevention ───────────────
echo "4. SQL Injection Prevention"
echo "---"

# Check for raw SQL string concatenation (common injection vector)
RAW_SQL=$(grep -rn "fmt.Sprintf.*SELECT\|fmt.Sprintf.*INSERT\|fmt.Sprintf.*UPDATE\|fmt.Sprintf.*DELETE" --include="*.go" --exclude-dir=vendor . 2>/dev/null || true)
if [ -z "$RAW_SQL" ]; then
    pass "No raw SQL string formatting detected (using sqlc)"
else
    fail "Raw SQL string formatting found — potential SQL injection:"
    echo "$RAW_SQL" | head -5
fi
echo

# ── 5. Authentication Configuration ──────────
echo "5. Authentication Configuration"
echo "---"

# Verify PASETO token usage (not JWT)
if grep -rq "jwt" --include="*.go" --exclude-dir=vendor . 2>/dev/null; then
    warn "JWT references found — Linkrift should use PASETO tokens"
else
    pass "No JWT usage — PASETO tokens in use"
fi

# Verify password hashing uses Argon2id
if grep -rq "argon2id\|crypto.HashPassword" --include="*.go" --exclude-dir=vendor . 2>/dev/null; then
    pass "Argon2id password hashing detected"
else
    warn "Could not confirm Argon2id usage for password hashing"
fi
echo

# ── 6. CORS Configuration ────────────────────
echo "6. CORS Configuration"
echo "---"

if grep -rq 'AllowOrigins.*\*\|AllowAllOrigins.*true' --include="*.go" --exclude-dir=vendor . 2>/dev/null; then
    fail "Wildcard CORS origin detected — restrict to specific domains"
else
    pass "CORS not using wildcard origins"
fi
echo

# ── 7. TLS Configuration ─────────────────────
echo "7. TLS Configuration"
echo "---"

# Check Helm values for TLS
if [ -f deployments/helm/linkrift/values.yaml ]; then
    if grep -q "tls:" deployments/helm/linkrift/values.yaml; then
        pass "TLS configured in Helm values"
    else
        warn "No TLS configuration found in Helm values"
    fi
fi

# Check for SSL redirect in ingress
if grep -rq "ssl-redirect.*true" deployments/helm/linkrift/templates/ 2>/dev/null; then
    pass "SSL redirect enabled in ingress"
else
    warn "SSL redirect not confirmed in ingress annotations"
fi
echo

# ── 8. Docker Image Security ─────────────────
echo "8. Docker Image Security"
echo "---"

if [ -f docker/Dockerfile.api ]; then
    # Check for non-root user
    if grep -q "USER" docker/Dockerfile.api; then
        pass "API Dockerfile uses non-root user"
    else
        warn "API Dockerfile may run as root — add USER directive"
    fi

    # Check for distroless / scratch / alpine base
    if grep -q "scratch\|distroless\|alpine" docker/Dockerfile.api; then
        pass "API Dockerfile uses minimal base image"
    else
        warn "API Dockerfile may not use minimal base image"
    fi
fi
echo

# ── Summary ───────────────────────────────────
echo "============================================"
echo " Summary"
echo "============================================"
echo -e "  ${GREEN}Passed${NC}: $PASS"
echo -e "  ${RED}Failed${NC}: $FAIL"
echo -e "  ${YELLOW}Warnings${NC}: $WARN"
echo

if [ "$FAIL" -gt 0 ]; then
    echo -e "${RED}Security audit FAILED — fix $FAIL issue(s) before deploying${NC}"
    exit 1
elif [ "$WARN" -gt 0 ]; then
    echo -e "${YELLOW}Security audit passed with $WARN warning(s)${NC}"
    exit 0
else
    echo -e "${GREEN}Security audit PASSED${NC}"
    exit 0
fi
