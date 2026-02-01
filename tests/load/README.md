# Load Tests

Load tests for Linkrift services using [k6](https://k6.io/).

## Prerequisites

1. Install k6: `brew install k6` (macOS) or see [k6 installation guide](https://k6.io/docs/getting-started/installation/)
2. Running services: API server (port 8080), redirect server (port 8081), PostgreSQL, Redis

## Seed Test Data

Before running load tests, seed the database with test links:

```sql
INSERT INTO links (user_id, workspace_id, url, short_code, is_active, total_clicks, unique_clicks)
VALUES
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'https://example.com', 'test1', true, 0, 0),
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'https://example.org', 'test2', true, 0, 0),
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'https://example.net', 'test3', true, 0, 0);
```

## Test Scripts

| Script | Service | Description |
|--------|---------|-------------|
| `redirect.js` | Redirect (8081) | Multi-scenario: warmup, sustained, spike |
| `api.js` | API (8080) | Mixed read/write API operations |
| `redirect_test.js` | Redirect (8081) | Simple ramp-up stress test |

## Running

### Redirect service load test

```bash
# Full multi-scenario test (warmup → sustained → spike)
k6 run tests/load/redirect.js

# With custom short codes
k6 run -e SHORT_CODES=test1,test2,test3 tests/load/redirect.js

# Custom redirect server URL
k6 run -e REDIRECT_URL=http://localhost:8081 tests/load/redirect.js
```

### API service load test

```bash
# Requires a valid access token and workspace ID
k6 run -e ACCESS_TOKEN=<token> -e WORKSPACE_ID=<id> tests/load/api.js

# Custom API URL
k6 run -e API_URL=http://localhost:8080/api/v1 \
  -e ACCESS_TOKEN=<token> \
  -e WORKSPACE_ID=<id> \
  tests/load/api.js
```

### Simple redirect stress test

```bash
k6 run tests/load/redirect_test.js

# Quick smoke test (lower load)
k6 run --vus 10 --duration 30s tests/load/redirect_test.js
```

## Performance Targets

### Redirect Service

| Metric | Target | Description |
|--------|--------|-------------|
| `http_req_duration p(50)` | < 10ms | Median latency |
| `http_req_duration p(95)` | < 50ms | 95th percentile latency |
| `http_req_duration p(99)` | < 100ms | 99th percentile latency |
| `redirect_success` | > 99% | Success rate |
| `http_req_failed` | < 1% | Error rate |

### API Service

| Metric | Target | Description |
|--------|--------|-------------|
| `http_req_duration p(95)` | < 500ms | 95th percentile latency |
| `api_success` | > 95% | Success rate |
| `http_req_failed` | < 5% | Error rate |

Per-endpoint targets are documented in `baselines.json`.

## Scenarios

### redirect.js

1. **Warmup** — Ramp from 0 to 100 VUs over 90 seconds, then back down
2. **Sustained** — Hold 50 VUs for 60 seconds (starts at t=95s)
3. **Spike** — Burst to 200 VUs for 10 seconds (starts at t=160s)

### api.js

1. **API Load** — Ramp from 0 to 50 VUs over 90 seconds with sustained plateau

### redirect_test.js

1. **Ramp-up** — Ramp from 50 to 500 VUs over 2 minutes, sustain for 3 minutes

## Baselines

Performance baselines from Go benchmark tests are recorded in `baselines.json`. Key benchmarks:

- Resolver cache hit: ~109ns
- L1 cache get: ~41ns
- Bot detector (human): ~56us
- Rule engine match: ~450ns

## Continuous Load Testing

For CI integration, run with JSON output:

```bash
k6 run --out json=results.json tests/load/redirect.js
k6 run --out json=api-results.json \
  -e ACCESS_TOKEN=$TOKEN \
  -e WORKSPACE_ID=$WS_ID \
  tests/load/api.js
```

Compare results against `baselines.json` targets to detect performance regressions.
