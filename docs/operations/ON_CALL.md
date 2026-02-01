# On-Call Procedures

Guide for on-call engineers handling Linkrift production incidents.

---

## On-Call Responsibilities

The on-call engineer is responsible for:
1. Responding to alerts within the SLA (critical: 5 min, warning: 30 min)
2. Triaging and resolving production incidents
3. Escalating when unable to resolve within 30 minutes
4. Documenting actions taken during incidents
5. Conducting post-incident reviews for severity 1-2 incidents

---

## Escalation Policy

| Level | Timeframe | Who | Action |
|-------|-----------|-----|--------|
| L1 | 0-5 min | On-call engineer | Acknowledge alert, begin triage |
| L2 | 15 min | On-call + team lead | Engage additional engineer |
| L3 | 30 min | Engineering manager | Consider war room |
| L4 | 60 min | VP Engineering + CTO | Executive notification |

### Severity Levels

| Severity | Description | Response SLA | Resolution SLA |
|----------|-------------|--------------|----------------|
| SEV1 | Complete outage, all users affected | 5 min | 1 hour |
| SEV2 | Major degradation, >50% users affected | 15 min | 4 hours |
| SEV3 | Minor degradation, <50% users affected | 30 min | 24 hours |
| SEV4 | Low impact, workaround available | 1 hour | 72 hours |

---

## Alert Response Playbooks

### API Service Down

**Alert**: `APIServiceDown`
**Severity**: Critical

1. Check pod status:
   ```bash
   kubectl get pods -l app.kubernetes.io/name=api -n production
   ```

2. Check recent logs:
   ```bash
   kubectl logs -l app.kubernetes.io/name=api -n production --tail=100
   ```

3. Check if pods are crash-looping:
   ```bash
   kubectl describe pod -l app.kubernetes.io/name=api -n production | grep -A5 "State:"
   ```

4. Common causes:
   - **OOMKilled**: Increase memory limits in Helm values
   - **Database connection failure**: Check PostgreSQL connectivity
   - **Config error**: Check ConfigMap and Secrets
   - **Failing health check**: Check `/health` endpoint directly

5. Quick fix — restart:
   ```bash
   kubectl rollout restart deployment/linkrift-api -n production
   ```

---

### Redirect Service Down

**Alert**: `RedirectServiceDown`
**Severity**: Critical (user-facing)

1. Check pod status:
   ```bash
   kubectl get pods -l app.kubernetes.io/name=redirect -n production
   ```

2. Test redirect directly:
   ```bash
   curl -I https://lnkr.ft/health
   ```

3. Check Redis connectivity (redirect depends on Redis cache):
   ```bash
   kubectl exec -it deploy/linkrift-redirect -n production -- redis-cli -h $REDIS_HOST ping
   ```

4. Quick fix — restart:
   ```bash
   kubectl rollout restart deployment/linkrift-redirect -n production
   ```

---

### High Error Rate (5xx)

**Alert**: `HighAPIErrorRate` or `HighRedirectErrorRate`
**Severity**: Critical

1. Identify failing endpoints:
   ```bash
   # In Grafana, check "Request Rate by Status Code" panel
   # Or query Prometheus:
   # sum(rate(linkrift_http_requests_total{status_code=~"5.."}[5m])) by (path)
   ```

2. Check application logs for errors:
   ```bash
   kubectl logs -l app.kubernetes.io/name=api -n production --tail=200 | grep '"level":"error"'
   ```

3. Check database connectivity:
   ```bash
   kubectl exec -it deploy/linkrift-api -n production -- curl localhost:8080/health
   ```

4. Common causes:
   - **Database overloaded**: Check connection pool, slow queries
   - **External service down**: Check Redis, ClickHouse, Meilisearch
   - **Bad deployment**: Roll back to previous version
   - **Traffic spike**: Check if autoscaling is responding

---

### High Latency

**Alert**: `HighAPILatency` or `HighRedirectLatency`
**Severity**: Warning

1. Check current latency in Grafana dashboards

2. Check database query performance:
   ```bash
   # Query Prometheus for slow DB queries:
   # histogram_quantile(0.99, sum(rate(linkrift_db_query_duration_seconds_bucket[5m])) by (le, operation))
   ```

3. Check resource utilization:
   ```bash
   kubectl top pods -l app.kubernetes.io/instance=linkrift -n production
   ```

4. Check if autoscaling is at max:
   ```bash
   kubectl get hpa -n production
   ```

5. Common causes:
   - **Insufficient replicas**: Scale up manually or adjust HPA
   - **Slow database queries**: Check for missing indexes, lock contention
   - **Redis latency**: Check Redis memory usage and eviction policy
   - **Cold cache**: Expected after deployment, should resolve within minutes

---

### Database Issues

**Alert**: `HighDBConnectionUsage` or `SlowDBQueries` or `PostgreSQLDown`
**Severity**: Critical (if PostgreSQL down), Warning (otherwise)

1. Check database status:
   ```bash
   # If using managed DB (RDS), check AWS console
   # If self-managed:
   kubectl exec -it postgres-0 -n production -- psql -U linkrift -c "SELECT count(*) FROM pg_stat_activity;"
   ```

2. Check active connections and long-running queries:
   ```sql
   SELECT pid, now() - pg_stat_activity.query_start AS duration, query, state
   FROM pg_stat_activity
   WHERE state != 'idle'
   ORDER BY duration DESC
   LIMIT 10;
   ```

3. Kill long-running queries if necessary:
   ```sql
   SELECT pg_terminate_backend(pid) FROM pg_stat_activity
   WHERE duration > interval '5 minutes' AND state != 'idle';
   ```

4. If connection pool exhausted:
   - Scale down API replicas temporarily
   - Restart API pods to release connections
   - Increase `max_open_conns` in config

---

### Redis Down

**Alert**: `RedisDown`
**Severity**: Critical

1. The redirect service will fall back to database queries (slower but functional)

2. Check Redis status:
   ```bash
   redis-cli -h $REDIS_HOST ping
   redis-cli -h $REDIS_HOST info memory
   ```

3. If OOM:
   ```bash
   redis-cli -h $REDIS_HOST config set maxmemory-policy allkeys-lru
   ```

4. Restart Redis if unresponsive:
   ```bash
   kubectl rollout restart statefulset/redis -n production
   ```

---

## Useful Commands

### Service Status
```bash
# All pods
kubectl get pods -n production

# Deployment status
kubectl rollout status deployment/linkrift-api -n production

# Recent events
kubectl get events -n production --sort-by=.metadata.creationTimestamp | tail -20
```

### Quick Rollback
```bash
# View revision history
kubectl rollout history deployment/linkrift-api -n production

# Rollback to previous version
kubectl rollout undo deployment/linkrift-api -n production

# Rollback to specific revision
kubectl rollout undo deployment/linkrift-api -n production --to-revision=3
```

### Scale
```bash
# Manual scale
kubectl scale deployment/linkrift-api --replicas=5 -n production

# Check HPA status
kubectl get hpa -n production
```

### Logs
```bash
# Last 100 lines
kubectl logs -l app.kubernetes.io/name=api -n production --tail=100

# Follow logs
kubectl logs -l app.kubernetes.io/name=api -n production -f

# Logs from previous container (after crash)
kubectl logs -l app.kubernetes.io/name=api -n production --previous
```

### Dashboards

| Dashboard | URL | Purpose |
|-----------|-----|---------|
| Grafana Overview | `https://grafana.linkrift.io/d/linkrift-overview` | System overview |
| Redirect Performance | `https://grafana.linkrift.io/d/linkrift-redirect` | Redirect latency |
| Prometheus Alerts | `https://prometheus.linkrift.io/alerts` | Active alerts |
| Alertmanager | `https://alertmanager.linkrift.io` | Alert routing |

---

## Post-Incident Review

After any SEV1 or SEV2 incident, conduct a blameless post-incident review within 48 hours.

### Template

```markdown
# Post-Incident Review: [Title]

**Date**: YYYY-MM-DD
**Duration**: HH:MM - HH:MM (X minutes)
**Severity**: SEV1/SEV2
**Impact**: [Number of users affected, services affected]

## Timeline
- HH:MM — Alert triggered
- HH:MM — On-call acknowledged
- HH:MM — Root cause identified
- HH:MM — Fix deployed
- HH:MM — Monitoring confirmed resolution

## Root Cause
[Description of the root cause]

## Resolution
[What was done to resolve the incident]

## Action Items
- [ ] [Prevention measure 1]
- [ ] [Prevention measure 2]
- [ ] [Detection improvement]

## Lessons Learned
- [What went well]
- [What could be improved]
```

---

## Communication Templates

### Internal (Slack)
```
:rotating_light: [SEV1] API service experiencing elevated error rates
Impact: Users may see errors when accessing the dashboard
Status: Investigating
On-call: @engineer
```

### Status Page Update
```
Investigating - We are currently investigating reports of elevated
error rates on the Linkrift platform. Some users may experience
intermittent errors. Our team is actively working on a resolution.
```

### Resolution
```
Resolved - The issue causing elevated error rates has been resolved.
The root cause was [brief description]. All services are operating
normally. We will publish a detailed post-incident review.
```
