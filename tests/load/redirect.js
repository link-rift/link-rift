import http from "k6/http"
import { check, sleep } from "k6"
import { Rate, Trend } from "k6/metrics"

// Custom metrics
const redirectSuccess = new Rate("redirect_success")
const redirectDuration = new Trend("redirect_duration", true)

// Configuration
const REDIRECT_URL = __ENV.REDIRECT_URL || "http://localhost:8081"
const SHORT_CODES = (__ENV.SHORT_CODES || "abc123,def456,ghi789").split(",")

export const options = {
  scenarios: {
    // Warm-up: gradual ramp
    warmup: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "10s", target: 10 },
        { duration: "20s", target: 50 },
        { duration: "30s", target: 100 },
        { duration: "20s", target: 50 },
        { duration: "10s", target: 0 },
      ],
      gracefulRampDown: "5s",
    },
    // Sustained load
    sustained: {
      executor: "constant-vus",
      vus: 50,
      duration: "60s",
      startTime: "95s",
    },
    // Spike test
    spike: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "5s", target: 200 },
        { duration: "10s", target: 200 },
        { duration: "5s", target: 0 },
      ],
      startTime: "160s",
    },
  },
  thresholds: {
    // Redirect latency targets
    http_req_duration: ["p(50)<10", "p(95)<50", "p(99)<100"],
    redirect_duration: ["p(50)<5", "p(95)<25", "p(99)<50"],
    redirect_success: ["rate>0.99"],
    http_req_failed: ["rate<0.01"],
  },
}

export default function () {
  const code = SHORT_CODES[Math.floor(Math.random() * SHORT_CODES.length)]
  const url = `${REDIRECT_URL}/${code}`

  const res = http.get(url, {
    redirects: 0, // Don't follow redirects — we're testing the redirect service
    tags: { name: "redirect" },
  })

  const success = res.status === 301 || res.status === 302 || res.status === 307
  redirectSuccess.add(success)
  redirectDuration.add(res.timings.duration)

  check(res, {
    "is redirect": (r) => r.status >= 300 && r.status < 400,
    "has location header": (r) => r.headers["Location"] !== undefined,
    "latency < 50ms": (r) => r.timings.duration < 50,
  })

  sleep(0.1)
}

export function handleSummary(data) {
  const p50 = data.metrics.http_req_duration?.values?.["p(50)"] || 0
  const p95 = data.metrics.http_req_duration?.values?.["p(95)"] || 0
  const p99 = data.metrics.http_req_duration?.values?.["p(99)"] || 0
  const total = data.metrics.http_reqs?.values?.count || 0
  const rate = data.metrics.http_reqs?.values?.rate || 0

  return {
    stdout: `
=== Redirect Service Load Test Results ===
Total requests: ${total}
Requests/sec:   ${rate.toFixed(1)}
Latency p50:    ${p50.toFixed(2)}ms
Latency p95:    ${p95.toFixed(2)}ms
Latency p99:    ${p99.toFixed(2)}ms
Success rate:   ${((data.metrics.redirect_success?.values?.rate || 0) * 100).toFixed(2)}%
==========================================
`,
  }
}
