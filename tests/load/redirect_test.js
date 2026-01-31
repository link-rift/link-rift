import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const redirectLatency = new Trend('redirect_latency', true);
const failureRate = new Rate('failures');

// Test configuration
const BASE_URL = __ENV.REDIRECT_URL || 'http://localhost:8081';
const SHORT_CODES = (__ENV.SHORT_CODES || 'test1,test2,test3').split(',');

export const options = {
  scenarios: {
    // Ramp-up test
    ramp_up: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50 },    // Warm up
        { duration: '1m', target: 200 },     // Ramp to 200
        { duration: '1m', target: 500 },     // Ramp to 500
        { duration: '3m', target: 500 },     // Sustain 500
        { duration: '30s', target: 0 },      // Ramp down
      ],
      gracefulRampDown: '10s',
    },
  },
  thresholds: {
    // Target: <1ms p99 redirect latency
    'redirect_latency': ['p(99)<1'],
    // Target: <1% failure rate
    'failures': ['rate<0.01'],
    // HTTP-level thresholds
    'http_req_duration': ['p(95)<5', 'p(99)<10'],
    'http_req_failed': ['rate<0.01'],
  },
};

export default function () {
  // Pick a random short code from the list
  const shortCode = SHORT_CODES[Math.floor(Math.random() * SHORT_CODES.length)];

  const res = http.get(`${BASE_URL}/${shortCode}`, {
    // Don't follow redirects — we're measuring the redirect response time
    redirects: 0,
    tags: { name: 'redirect' },
  });

  const isSuccess = res.status === 301 || res.status === 302 || res.status === 307 || res.status === 308;

  redirectLatency.add(res.timings.duration);
  failureRate.add(!isSuccess);

  check(res, {
    'is redirect': (r) => [301, 302, 307, 308].includes(r.status),
    'has location header': (r) => r.headers['Location'] !== undefined,
    'latency < 1ms': (r) => r.timings.duration < 1,
    'latency < 5ms': (r) => r.timings.duration < 5,
  });

  // Small think-time to simulate real traffic
  sleep(0.01);
}

export function handleSummary(data) {
  const p99 = data.metrics.redirect_latency?.values?.['p(99)'] || 'N/A';
  const p95 = data.metrics.redirect_latency?.values?.['p(95)'] || 'N/A';
  const med = data.metrics.redirect_latency?.values?.['med'] || 'N/A';

  console.log('\n=== Redirect Performance Summary ===');
  console.log(`  Median latency:  ${med}ms`);
  console.log(`  p95 latency:     ${p95}ms`);
  console.log(`  p99 latency:     ${p99}ms`);
  console.log(`  Target:          <1ms p99`);
  console.log(`  Status:          ${p99 < 1 ? 'PASS' : 'FAIL'}`);
  console.log('=====================================\n');

  return {
    stdout: JSON.stringify(data, null, 2),
  };
}
