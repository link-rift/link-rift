# Production Launch Checklist

Pre-launch verification checklist for Linkrift production deployment.

---

## 1. Security Audit

### Application Security
- [ ] All dependencies scanned for vulnerabilities (`govulncheck ./...`)
- [ ] Static analysis clean (`gosec ./...`)
- [ ] No hardcoded secrets in codebase (`grep -r "password\|secret\|token" --include="*.go" | grep -v _test.go`)
- [ ] All API endpoints require authentication (except public routes)
- [ ] Rate limiting enabled on all public endpoints
- [ ] CORS configured with production origins only
- [ ] Security headers configured (CSP, HSTS, X-Frame-Options, X-Content-Type-Options)
- [ ] Input validation on all user-facing endpoints
- [ ] SQL injection prevention verified (sqlc parameterized queries)
- [ ] XSS prevention verified (HTML escaping in redirect templates)

### Infrastructure Security
- [ ] Database credentials rotated from development defaults
- [ ] Redis requires authentication
- [ ] ClickHouse requires authentication
- [ ] S3 bucket is private with IAM-based access
- [ ] Kubernetes NetworkPolicy restricts pod-to-pod traffic
- [ ] Secrets stored in external secret manager (not ConfigMaps)
- [ ] TLS 1.2+ enforced on all connections
- [ ] Admin endpoints not exposed publicly

### Authentication Security
- [ ] PASETO symmetric key is unique and securely generated (32+ bytes)
- [ ] Access token expiry is short (15m default)
- [ ] Refresh token rotation is enabled
- [ ] Session invalidation on password change works
- [ ] Brute-force protection on login endpoint
- [ ] API key hashing uses SHA-256 (not stored in plaintext)

**Reference**: [Security Documentation](../security/SECURITY.md)

---

## 2. Performance Testing

### Baseline Benchmarks
- [ ] API latency P50 < 50ms, P99 < 500ms
- [ ] Redirect latency P50 < 1ms, P99 < 10ms
- [ ] Redirect throughput > 10,000 req/s per instance
- [ ] Database query P99 < 100ms
- [ ] Cache hit rate > 90% for redirects

### Load Test Results
- [ ] Run load test: `k6 run tests/load/api_test.js`
- [ ] Run redirect load test: `k6 run tests/load/redirect_test.js`
- [ ] Verify autoscaling triggers correctly under load
- [ ] Verify graceful degradation under extreme load
- [ ] No memory leaks under sustained load (monitor Go heap)
- [ ] No goroutine leaks under sustained load

### Database Performance
- [ ] All critical queries have execution plans reviewed (`EXPLAIN ANALYZE`)
- [ ] Indexes created for all foreign keys and frequent query patterns
- [ ] Connection pool size tuned (max_open_conns, max_idle_conns)
- [ ] ClickHouse partitioning configured for analytics tables

**Reference**: [Scaling Guide](SCALING_GUIDE.md)

---

## 3. Test Suite

### Backend Tests
```bash
make test          # All unit tests pass
make test-cover    # Coverage > 70%
make bench         # No performance regressions
make lint          # No linting errors
make security      # gosec + govulncheck clean
```

- [ ] All unit tests passing
- [ ] Integration tests passing against test database
- [ ] API endpoint tests covering all routes
- [ ] Auth flow tests (login, register, refresh, logout)
- [ ] Redirect flow tests (resolve, password, expired, disabled)
- [ ] Webhook delivery tests
- [ ] License verification tests

### Frontend Tests
```bash
cd web && npm test       # All tests pass
cd web && npm run build  # Production build succeeds
```

- [ ] Component tests passing
- [ ] Hook tests passing
- [ ] Build produces no warnings

---

## 4. Documentation

- [ ] API documentation available at `/api/docs` (Swagger UI)
- [ ] OpenAPI spec is up to date with all endpoints
- [ ] Docs site builds successfully (`cd docs-site && npm run build`)
- [ ] Architecture documentation reflects current implementation
- [ ] Database schema documentation matches migrations
- [ ] Environment variables documented in deployment guide
- [ ] Runbook procedures tested and verified

**Reference**: [API Documentation](../api/API_DOCUMENTATION.md)

---

## 5. Backup Procedures

### PostgreSQL
```bash
# Verify backup script works
pg_dump -Fc linkrift > /backups/linkrift_$(date +%Y%m%d).dump

# Verify restore works
pg_restore -d linkrift_test /backups/linkrift_test.dump
```

- [ ] Automated daily backups configured (pg_dump or managed service snapshots)
- [ ] Backup retention policy: 7 daily, 4 weekly, 12 monthly
- [ ] Backup restore tested on staging environment
- [ ] Point-in-time recovery (PITR) configured via WAL archiving
- [ ] Backup encryption at rest enabled

### Redis
- [ ] RDB snapshots enabled (every 60s if 1000+ changes)
- [ ] AOF persistence enabled for durability
- [ ] Redis backup restore tested

### ClickHouse
- [ ] Automated backup of analytics data configured
- [ ] Partition-level backup strategy for large tables

### S3
- [ ] Versioning enabled on upload bucket
- [ ] Cross-region replication configured for critical assets

**Reference**: [Maintenance Guide](../operations/MAINTENANCE.md)

---

## 6. Rollback Procedures

### Application Rollback
```bash
# Kubernetes — rollback to previous revision
kubectl rollout undo deployment/linkrift-api -n production
kubectl rollout undo deployment/linkrift-redirect -n production
kubectl rollout undo deployment/linkrift-worker -n production
kubectl rollout undo deployment/linkrift-web -n production

# Verify rollback
kubectl rollout status deployment/linkrift-api -n production
```

### Database Rollback
```bash
# Rollback last migration
migrate -path migrations/postgres -database "$DATABASE_URL" down 1

# Verify schema version
migrate -path migrations/postgres -database "$DATABASE_URL" version
```

- [ ] Rollback procedure documented and tested on staging
- [ ] Database migration rollback tested (up and down)
- [ ] Blue-green or canary deployment strategy configured
- [ ] Previous container image tagged and available in registry
- [ ] Rollback can be completed within 5 minutes
- [ ] Rollback does not cause data loss

**Reference**: [Deployment Guide](DEPLOYMENT_GUIDE.md)

---

## 7. SSL/TLS Certificates

- [ ] TLS certificates provisioned for all domains
- [ ] cert-manager installed in Kubernetes cluster
- [ ] ClusterIssuer configured for Let's Encrypt production
- [ ] Certificate auto-renewal verified (check expiry > 30 days)
- [ ] HSTS header configured with max-age >= 31536000
- [ ] SSL redirect enabled (HTTP → HTTPS)
- [ ] TLS 1.2 minimum version enforced
- [ ] Strong cipher suites configured

### Verification
```bash
# Check certificate
openssl s_client -connect app.linkrift.io:443 -servername app.linkrift.io < /dev/null 2>/dev/null | openssl x509 -noout -dates

# Check TLS version
nmap --script ssl-enum-ciphers -p 443 app.linkrift.io
```

---

## 8. DNS Configuration

- [ ] A/AAAA records pointing to load balancer / ingress IP
- [ ] `app.linkrift.io` → Application (API + Web)
- [ ] `lnkr.ft` → Redirect service
- [ ] CNAME records for custom domains pointing to redirect service
- [ ] MX records for email delivery (if applicable)
- [ ] SPF, DKIM, DMARC records for email authentication
- [ ] DNS propagation verified (`dig +short app.linkrift.io`)
- [ ] TTL set appropriately (300s for A records, 3600s for MX)
- [ ] Cloudflare proxy enabled for DDoS protection
- [ ] DNS failover configured (if multi-region)

### Verification
```bash
# Verify DNS resolution
dig +short app.linkrift.io
dig +short lnkr.ft

# Verify SSL via DNS
curl -I https://app.linkrift.io/health
curl -I https://lnkr.ft/health
```

---

## 9. Monitoring Configuration

- [ ] Prometheus scraping all services (API, Redirect, Worker)
- [ ] Grafana dashboards imported (Overview, Redirect Performance)
- [ ] Alerting rules configured and tested
- [ ] Alert notification channels configured (Slack / PagerDuty)
- [ ] Log aggregation active (Promtail → Loki)
- [ ] Uptime monitoring configured (external service)
- [ ] Custom metrics emitting correctly (`/metrics` endpoint accessible)

### Verification
```bash
# Check Prometheus targets
curl -s http://prometheus:9090/api/v1/targets | jq '.data.activeTargets[].health'

# Check alert rules
curl -s http://prometheus:9090/api/v1/rules | jq '.data.groups[].rules[].name'

# Verify metrics endpoint
curl -s http://api:9090/metrics | head -20
curl -s http://redirect:9091/metrics | head -20
```

**Reference**: [Monitoring & Logging](../operations/MONITORING_LOGGING.md)

---

## 10. On-Call Procedures

- [ ] On-call rotation schedule established
- [ ] Escalation policy documented
- [ ] Runbook for common incidents available
- [ ] PagerDuty / OpsGenie integration configured
- [ ] War room communication channel designated (Slack)
- [ ] Post-incident review template created
- [ ] All team members have access to production dashboards
- [ ] Emergency contact list maintained

**Reference**: [On-Call Procedures](../operations/ON_CALL.md)

---

## Final Verification

Run the complete verification script:

```bash
# Backend
make test && make lint && make build && make security

# Frontend
cd web && npm test && npm run build

# Docs site
cd docs-site && npm run build

# Verify health endpoints
curl -sf http://localhost:8080/health
curl -sf http://localhost:8081/health
```

### Sign-Off

| Area | Owner | Status | Date |
|------|-------|--------|------|
| Security Audit | | | |
| Performance Testing | | | |
| Test Suite | | | |
| Documentation | | | |
| Backup Procedures | | | |
| Rollback Procedures | | | |
| SSL/TLS | | | |
| DNS | | | |
| Monitoring | | | |
| On-Call | | | |

**Launch approved by**: _______________
**Date**: _______________
