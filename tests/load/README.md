# Load Tests

Load tests for the Linkrift redirect service using [k6](https://k6.io/).

## Prerequisites

1. Install k6: `brew install k6` (macOS) or see [k6 installation guide](https://k6.io/docs/getting-started/installation/)
2. Running redirect server, PostgreSQL, and Redis

## Seed Test Data

Before running load tests, seed the database with test links:

```sql
-- Connect to the linkrift database
INSERT INTO links (user_id, workspace_id, url, short_code, is_active, total_clicks, unique_clicks)
VALUES
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'https://example.com', 'test1', true, 0, 0),
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'https://example.org', 'test2', true, 0, 0),
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'https://example.net', 'test3', true, 0, 0);
```

## Running

### Basic load test

```bash
k6 run tests/load/redirect_test.js
```

### Custom redirect server URL

```bash
k6 run -e REDIRECT_URL=http://localhost:8081 tests/load/redirect_test.js
```

### Custom short codes

```bash
k6 run -e SHORT_CODES=abc,def,ghi tests/load/redirect_test.js
```

### Quick smoke test (lower load)

```bash
k6 run --vus 10 --duration 30s tests/load/redirect_test.js
```

## Thresholds

| Metric | Target | Description |
|--------|--------|-------------|
| `redirect_latency p(99)` | < 1ms | 99th percentile redirect latency |
| `failures` | < 1% | Error rate |
| `http_req_duration p(95)` | < 5ms | 95th percentile overall HTTP duration |
| `http_req_duration p(99)` | < 10ms | 99th percentile overall HTTP duration |

## Scenarios

The default test ramps up from 50 to 500 virtual users over 2 minutes, sustains 500 VUs for 3 minutes, then ramps down. Total duration: ~6 minutes.
