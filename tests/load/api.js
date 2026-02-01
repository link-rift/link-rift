import http from "k6/http"
import { check, sleep, group } from "k6"
import { Rate, Trend } from "k6/metrics"

// Custom metrics
const apiSuccess = new Rate("api_success")
const apiDuration = new Trend("api_duration", true)

// Configuration
const API_URL = __ENV.API_URL || "http://localhost:8080/api/v1"
const ACCESS_TOKEN = __ENV.ACCESS_TOKEN || ""
const WORKSPACE_ID = __ENV.WORKSPACE_ID || ""

const headers = {
  "Content-Type": "application/json",
  Authorization: `Bearer ${ACCESS_TOKEN}`,
}

export const options = {
  scenarios: {
    // Standard API load
    api_load: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "15s", target: 20 },
        { duration: "30s", target: 50 },
        { duration: "30s", target: 50 },
        { duration: "15s", target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<500"],
    api_success: ["rate>0.95"],
    http_req_failed: ["rate<0.05"],
  },
}

export default function () {
  group("List links", () => {
    const res = http.get(`${API_URL}/workspaces/${WORKSPACE_ID}/links?limit=20`, {
      headers,
      tags: { name: "list_links" },
    })
    const success = res.status === 200
    apiSuccess.add(success)
    apiDuration.add(res.timings.duration)
    check(res, {
      "list links 200": (r) => r.status === 200,
      "list links < 500ms": (r) => r.timings.duration < 500,
    })
  })

  sleep(0.5)

  group("Create link", () => {
    const payload = JSON.stringify({
      url: `https://example.com/load-test-${Date.now()}`,
    })
    const res = http.post(`${API_URL}/workspaces/${WORKSPACE_ID}/links`, payload, {
      headers,
      tags: { name: "create_link" },
    })
    const success = res.status === 201 || res.status === 200
    apiSuccess.add(success)
    apiDuration.add(res.timings.duration)
    check(res, {
      "create link success": (r) => r.status === 201 || r.status === 200,
      "create link < 500ms": (r) => r.timings.duration < 500,
    })

    // Clean up — delete the created link
    if (success) {
      try {
        const body = JSON.parse(res.body)
        if (body.data?.id) {
          http.del(`${API_URL}/workspaces/${WORKSPACE_ID}/links/${body.data.id}`, null, {
            headers,
            tags: { name: "delete_link" },
          })
        }
      } catch {
        // Ignore parse errors
      }
    }
  })

  sleep(0.5)

  group("Get workspaces", () => {
    const res = http.get(`${API_URL}/workspaces`, {
      headers,
      tags: { name: "list_workspaces" },
    })
    apiSuccess.add(res.status === 200)
    apiDuration.add(res.timings.duration)
    check(res, {
      "workspaces 200": (r) => r.status === 200,
      "workspaces < 300ms": (r) => r.timings.duration < 300,
    })
  })

  sleep(0.5)

  group("Get current user", () => {
    const res = http.get(`${API_URL}/auth/me`, {
      headers,
      tags: { name: "get_me" },
    })
    apiSuccess.add(res.status === 200)
    apiDuration.add(res.timings.duration)
    check(res, {
      "auth/me 200": (r) => r.status === 200,
      "auth/me < 200ms": (r) => r.timings.duration < 200,
    })
  })

  sleep(1)
}

export function handleSummary(data) {
  const p50 = data.metrics.http_req_duration?.values?.["p(50)"] || 0
  const p95 = data.metrics.http_req_duration?.values?.["p(95)"] || 0
  const p99 = data.metrics.http_req_duration?.values?.["p(99)"] || 0
  const total = data.metrics.http_reqs?.values?.count || 0
  const rate = data.metrics.http_reqs?.values?.rate || 0

  return {
    stdout: `
=== API Load Test Results ===
Total requests: ${total}
Requests/sec:   ${rate.toFixed(1)}
Latency p50:    ${p50.toFixed(2)}ms
Latency p95:    ${p95.toFixed(2)}ms
Latency p99:    ${p99.toFixed(2)}ms
Success rate:   ${((data.metrics.api_success?.values?.rate || 0) * 100).toFixed(2)}%
=============================
`,
  }
}
