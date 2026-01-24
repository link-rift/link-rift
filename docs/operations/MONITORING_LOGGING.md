# Monitoring and Logging

**Last Updated: 2025-01-24**

This document covers Linkrift's monitoring and logging infrastructure, including structured logging, metrics collection, distributed tracing, and alerting.

---

## Table of Contents

- [Overview](#overview)
- [Structured Logging with Zap](#structured-logging-with-zap)
  - [Logger Configuration](#logger-configuration)
  - [Log Levels](#log-levels)
  - [Contextual Logging](#contextual-logging)
  - [Request Logging Middleware](#request-logging-middleware)
- [Prometheus Metrics](#prometheus-metrics)
  - [Metrics Configuration](#metrics-configuration)
  - [Key Metrics](#key-metrics)
  - [Custom Metrics](#custom-metrics)
  - [Metrics Endpoint](#metrics-endpoint)
- [Grafana Dashboards](#grafana-dashboards)
  - [Dashboard Overview](#dashboard-overview)
  - [Redirect Performance Dashboard](#redirect-performance-dashboard)
  - [System Health Dashboard](#system-health-dashboard)
- [Key Performance Indicators](#key-performance-indicators)
  - [Redirect Latency](#redirect-latency)
  - [Throughput Metrics](#throughput-metrics)
  - [Error Rates](#error-rates)
- [Alerting Rules](#alerting-rules)
  - [Prometheus Alerting](#prometheus-alerting)
  - [Alert Severity Levels](#alert-severity-levels)
  - [Notification Channels](#notification-channels)
- [OpenTelemetry Tracing](#opentelemetry-tracing)
  - [Tracing Configuration](#tracing-configuration)
  - [Span Instrumentation](#span-instrumentation)
  - [Trace Correlation](#trace-correlation)
- [Profiling with pprof](#profiling-with-pprof)
  - [Enabling pprof](#enabling-pprof)
  - [Available Profiles](#available-profiles)
  - [Profiling Examples](#profiling-examples)
- [Log Aggregation](#log-aggregation)
- [Best Practices](#best-practices)

---

## Overview

Linkrift employs a comprehensive observability stack:

| Component | Technology | Purpose |
|-----------|------------|---------|
| Logging | uber/zap | Structured, high-performance logging |
| Metrics | Prometheus | Time-series metrics collection |
| Visualization | Grafana | Dashboards and alerting |
| Tracing | OpenTelemetry | Distributed request tracing |
| Profiling | pprof | Runtime performance profiling |

---

## Structured Logging with Zap

### Logger Configuration

```go
// internal/logger/logger.go
package logger

import (
    "os"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Init(env string) error {
    var config zap.Config

    if env == "production" {
        config = zap.NewProductionConfig()
        config.EncoderConfig.TimeKey = "timestamp"
        config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
        config.EncoderConfig.StacktraceKey = "stacktrace"
    } else {
        config = zap.NewDevelopmentConfig()
        config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    }

    // Configure output
    config.OutputPaths = []string{"stdout"}
    config.ErrorOutputPaths = []string{"stderr"}

    // Set initial fields
    config.InitialFields = map[string]interface{}{
        "service": "linkrift",
        "version": os.Getenv("APP_VERSION"),
    }

    var err error
    Log, err = config.Build(
        zap.AddCaller(),
        zap.AddStacktrace(zapcore.ErrorLevel),
    )
    if err != nil {
        return err
    }

    return nil
}

// Sync flushes any buffered log entries
func Sync() {
    if Log != nil {
        _ = Log.Sync()
    }
}
```

### Log Levels

```go
// Usage examples for different log levels
package main

import (
    "go.uber.org/zap"
    "linkrift/internal/logger"
)

func exampleLogging() {
    // Debug - detailed information for debugging
    logger.Log.Debug("Processing redirect request",
        zap.String("short_code", "abc123"),
        zap.String("user_agent", "Mozilla/5.0..."),
    )

    // Info - general operational information
    logger.Log.Info("Link created successfully",
        zap.String("short_code", "abc123"),
        zap.String("user_id", "user_456"),
        zap.String("original_url", "https://example.com"),
    )

    // Warn - potentially harmful situations
    logger.Log.Warn("Rate limit approaching threshold",
        zap.String("ip", "192.168.1.100"),
        zap.Int("requests", 95),
        zap.Int("limit", 100),
    )

    // Error - error events that might still allow the application to continue
    logger.Log.Error("Failed to resolve short link",
        zap.String("short_code", "abc123"),
        zap.Error(err),
    )

    // Fatal - severe errors that cause premature termination
    logger.Log.Fatal("Database connection failed",
        zap.Error(err),
    )
}
```

### Contextual Logging

```go
// internal/logger/context.go
package logger

import (
    "context"

    "go.uber.org/zap"
)

type ctxKey struct{}

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, l *zap.Logger) context.Context {
    return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext retrieves the logger from context
func FromContext(ctx context.Context) *zap.Logger {
    if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
        return l
    }
    return Log
}

// WithFields creates a child logger with additional fields
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
    l := FromContext(ctx).With(fields...)
    return WithLogger(ctx, l)
}
```

### Request Logging Middleware

```go
// internal/middleware/logging.go
package middleware

import (
    "net/http"
    "time"

    "github.com/go-chi/chi/v5/middleware"
    "go.uber.org/zap"
    "linkrift/internal/logger"
)

func RequestLogger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        requestID := middleware.GetReqID(r.Context())

        // Create request-scoped logger
        reqLogger := logger.Log.With(
            zap.String("request_id", requestID),
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
            zap.String("remote_addr", r.RemoteAddr),
            zap.String("user_agent", r.UserAgent()),
        )

        // Wrap response writer to capture status code
        ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

        // Add logger to context
        ctx := logger.WithLogger(r.Context(), reqLogger)

        // Process request
        next.ServeHTTP(ww, r.WithContext(ctx))

        // Log completion
        duration := time.Since(start)
        reqLogger.Info("Request completed",
            zap.Int("status", ww.Status()),
            zap.Int("bytes", ww.BytesWritten()),
            zap.Duration("duration", duration),
            zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
        )
    })
}
```

---

## Prometheus Metrics

### Metrics Configuration

```go
// internal/metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP metrics
    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "linkrift_http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "linkrift_http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
        },
        []string{"method", "path"},
    )

    // Redirect metrics
    RedirectsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "linkrift_redirects_total",
            Help: "Total number of redirects performed",
        },
        []string{"status", "cached"},
    )

    RedirectLatency = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "linkrift_redirect_latency_seconds",
            Help:    "Redirect resolution latency in seconds",
            Buckets: []float64{.0001, .0005, .001, .005, .01, .025, .05, .1},
        },
    )

    // Link metrics
    LinksCreated = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "linkrift_links_created_total",
            Help: "Total number of links created",
        },
    )

    LinksActive = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "linkrift_links_active",
            Help: "Number of active links",
        },
    )

    // Cache metrics
    CacheHits = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "linkrift_cache_hits_total",
            Help: "Total number of cache hits",
        },
    )

    CacheMisses = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "linkrift_cache_misses_total",
            Help: "Total number of cache misses",
        },
    )

    // Database metrics
    DBQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "linkrift_db_query_duration_seconds",
            Help:    "Database query duration in seconds",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
        },
        []string{"query"},
    )

    DBConnectionsActive = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "linkrift_db_connections_active",
            Help: "Number of active database connections",
        },
    )
)
```

### Key Metrics

```go
// internal/metrics/collectors.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Throughput metrics
    RequestsPerSecond = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "linkrift_requests_per_second",
            Help: "Current requests per second",
        },
    )

    // Error metrics
    ErrorsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "linkrift_errors_total",
            Help: "Total number of errors by type",
        },
        []string{"type", "component"},
    )

    // Rate limiting metrics
    RateLimitHits = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "linkrift_rate_limit_hits_total",
            Help: "Total number of rate limit hits",
        },
        []string{"endpoint"},
    )

    // Custom domain metrics
    CustomDomainResolutions = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "linkrift_custom_domain_resolutions_total",
            Help: "Total custom domain resolutions",
        },
        []string{"domain", "status"},
    )

    // Click analytics metrics
    ClicksProcessed = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "linkrift_clicks_processed_total",
            Help: "Total clicks processed for analytics",
        },
    )

    ClicksQueueSize = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "linkrift_clicks_queue_size",
            Help: "Current size of clicks processing queue",
        },
    )
)
```

### Custom Metrics

```go
// internal/metrics/business.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Business metrics
    ActiveUsers = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "linkrift_active_users",
            Help: "Number of active users in the last 24 hours",
        },
    )

    APIKeyUsage = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "linkrift_api_key_usage_total",
            Help: "API key usage by key ID",
        },
        []string{"key_id", "endpoint"},
    )

    LinksByPlan = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "linkrift_links_by_plan",
            Help: "Number of links by subscription plan",
        },
        []string{"plan"},
    )
)

// RecordBusinessMetrics updates business-related metrics
func RecordBusinessMetrics(activeUserCount, freeLinks, proLinks, enterpriseLinks int64) {
    ActiveUsers.Set(float64(activeUserCount))
    LinksByPlan.WithLabelValues("free").Set(float64(freeLinks))
    LinksByPlan.WithLabelValues("pro").Set(float64(proLinks))
    LinksByPlan.WithLabelValues("enterprise").Set(float64(enterpriseLinks))
}
```

### Metrics Endpoint

```go
// internal/server/metrics.go
package server

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *Server) metricsRoutes() chi.Router {
    r := chi.NewRouter()

    // Prometheus metrics endpoint
    r.Handle("/metrics", promhttp.Handler())

    // Health check endpoints
    r.Get("/health", s.healthCheck)
    r.Get("/ready", s.readinessCheck)

    return r
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func (s *Server) readinessCheck(w http.ResponseWriter, r *http.Request) {
    // Check database connectivity
    if err := s.db.Ping(); err != nil {
        http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
        return
    }

    // Check Redis connectivity
    if err := s.redis.Ping(r.Context()).Err(); err != nil {
        http.Error(w, "Cache unavailable", http.StatusServiceUnavailable)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Ready"))
}
```

---

## Grafana Dashboards

### Dashboard Overview

```json
{
  "dashboard": {
    "title": "Linkrift Overview",
    "uid": "linkrift-overview",
    "tags": ["linkrift", "overview"],
    "timezone": "browser",
    "panels": [
      {
        "title": "Request Rate",
        "type": "stat",
        "gridPos": { "h": 4, "w": 6, "x": 0, "y": 0 },
        "targets": [
          {
            "expr": "sum(rate(linkrift_http_requests_total[5m]))",
            "legendFormat": "Requests/s"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "stat",
        "gridPos": { "h": 4, "w": 6, "x": 6, "y": 0 },
        "targets": [
          {
            "expr": "sum(rate(linkrift_http_requests_total{status=~\"5..\"}[5m])) / sum(rate(linkrift_http_requests_total[5m])) * 100",
            "legendFormat": "Error %"
          }
        ]
      },
      {
        "title": "P99 Latency",
        "type": "stat",
        "gridPos": { "h": 4, "w": 6, "x": 12, "y": 0 },
        "targets": [
          {
            "expr": "histogram_quantile(0.99, sum(rate(linkrift_http_request_duration_seconds_bucket[5m])) by (le))",
            "legendFormat": "P99"
          }
        ]
      },
      {
        "title": "Active Links",
        "type": "stat",
        "gridPos": { "h": 4, "w": 6, "x": 18, "y": 0 },
        "targets": [
          {
            "expr": "linkrift_links_active",
            "legendFormat": "Links"
          }
        ]
      }
    ]
  }
}
```

### Redirect Performance Dashboard

```json
{
  "dashboard": {
    "title": "Linkrift Redirect Performance",
    "uid": "linkrift-redirects",
    "panels": [
      {
        "title": "Redirect Latency Distribution",
        "type": "heatmap",
        "gridPos": { "h": 8, "w": 12, "x": 0, "y": 0 },
        "targets": [
          {
            "expr": "sum(increase(linkrift_redirect_latency_seconds_bucket[1m])) by (le)",
            "format": "heatmap"
          }
        ]
      },
      {
        "title": "Redirect Latency Percentiles",
        "type": "timeseries",
        "gridPos": { "h": 8, "w": 12, "x": 12, "y": 0 },
        "targets": [
          {
            "expr": "histogram_quantile(0.50, sum(rate(linkrift_redirect_latency_seconds_bucket[5m])) by (le))",
            "legendFormat": "P50"
          },
          {
            "expr": "histogram_quantile(0.95, sum(rate(linkrift_redirect_latency_seconds_bucket[5m])) by (le))",
            "legendFormat": "P95"
          },
          {
            "expr": "histogram_quantile(0.99, sum(rate(linkrift_redirect_latency_seconds_bucket[5m])) by (le))",
            "legendFormat": "P99"
          }
        ]
      },
      {
        "title": "Cache Hit Rate",
        "type": "gauge",
        "gridPos": { "h": 6, "w": 8, "x": 0, "y": 8 },
        "targets": [
          {
            "expr": "sum(rate(linkrift_cache_hits_total[5m])) / (sum(rate(linkrift_cache_hits_total[5m])) + sum(rate(linkrift_cache_misses_total[5m]))) * 100",
            "legendFormat": "Hit Rate %"
          }
        ]
      },
      {
        "title": "Redirects by Status",
        "type": "piechart",
        "gridPos": { "h": 6, "w": 8, "x": 8, "y": 8 },
        "targets": [
          {
            "expr": "sum(increase(linkrift_redirects_total[1h])) by (status)",
            "legendFormat": "{{status}}"
          }
        ]
      },
      {
        "title": "Throughput",
        "type": "timeseries",
        "gridPos": { "h": 6, "w": 8, "x": 16, "y": 8 },
        "targets": [
          {
            "expr": "sum(rate(linkrift_redirects_total[1m]))",
            "legendFormat": "Redirects/s"
          }
        ]
      }
    ]
  }
}
```

### System Health Dashboard

```json
{
  "dashboard": {
    "title": "Linkrift System Health",
    "uid": "linkrift-system",
    "panels": [
      {
        "title": "Go Routines",
        "type": "timeseries",
        "targets": [
          {
            "expr": "go_goroutines{job=\"linkrift\"}",
            "legendFormat": "Goroutines"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "type": "timeseries",
        "targets": [
          {
            "expr": "go_memstats_heap_alloc_bytes{job=\"linkrift\"}",
            "legendFormat": "Heap Alloc"
          },
          {
            "expr": "go_memstats_heap_sys_bytes{job=\"linkrift\"}",
            "legendFormat": "Heap Sys"
          }
        ]
      },
      {
        "title": "GC Pause Duration",
        "type": "timeseries",
        "targets": [
          {
            "expr": "rate(go_gc_duration_seconds_sum{job=\"linkrift\"}[5m])",
            "legendFormat": "GC Duration"
          }
        ]
      },
      {
        "title": "Database Connections",
        "type": "timeseries",
        "targets": [
          {
            "expr": "linkrift_db_connections_active",
            "legendFormat": "Active Connections"
          }
        ]
      }
    ]
  }
}
```

---

## Key Performance Indicators

### Redirect Latency

```go
// internal/service/redirect.go
package service

import (
    "context"
    "time"

    "linkrift/internal/metrics"
)

func (s *RedirectService) Resolve(ctx context.Context, shortCode string) (string, error) {
    start := time.Now()
    defer func() {
        metrics.RedirectLatency.Observe(time.Since(start).Seconds())
    }()

    // Check cache first
    if url, err := s.cache.Get(ctx, shortCode); err == nil {
        metrics.CacheHits.Inc()
        metrics.RedirectsTotal.WithLabelValues("success", "true").Inc()
        return url, nil
    }
    metrics.CacheMisses.Inc()

    // Query database
    link, err := s.repo.GetByShortCode(ctx, shortCode)
    if err != nil {
        metrics.RedirectsTotal.WithLabelValues("not_found", "false").Inc()
        return "", err
    }

    // Cache the result
    _ = s.cache.Set(ctx, shortCode, link.OriginalURL, 24*time.Hour)

    metrics.RedirectsTotal.WithLabelValues("success", "false").Inc()
    return link.OriginalURL, nil
}
```

### Throughput Metrics

```go
// internal/metrics/throughput.go
package metrics

import (
    "sync"
    "time"
)

type ThroughputTracker struct {
    mu       sync.RWMutex
    requests []time.Time
    window   time.Duration
}

func NewThroughputTracker(window time.Duration) *ThroughputTracker {
    t := &ThroughputTracker{
        requests: make([]time.Time, 0, 1000),
        window:   window,
    }
    go t.cleanup()
    return t
}

func (t *ThroughputTracker) Record() {
    t.mu.Lock()
    t.requests = append(t.requests, time.Now())
    t.mu.Unlock()
}

func (t *ThroughputTracker) Rate() float64 {
    t.mu.RLock()
    defer t.mu.RUnlock()

    cutoff := time.Now().Add(-t.window)
    count := 0
    for _, ts := range t.requests {
        if ts.After(cutoff) {
            count++
        }
    }
    return float64(count) / t.window.Seconds()
}

func (t *ThroughputTracker) cleanup() {
    ticker := time.NewTicker(time.Minute)
    for range ticker.C {
        t.mu.Lock()
        cutoff := time.Now().Add(-t.window)
        filtered := t.requests[:0]
        for _, ts := range t.requests {
            if ts.After(cutoff) {
                filtered = append(filtered, ts)
            }
        }
        t.requests = filtered
        t.mu.Unlock()
    }
}
```

### Error Rates

```go
// internal/middleware/errors.go
package middleware

import (
    "net/http"

    "linkrift/internal/metrics"
)

func ErrorTracker(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                metrics.ErrorsTotal.WithLabelValues("panic", "http").Inc()
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// RecordError records an error with its type and component
func RecordError(errType, component string) {
    metrics.ErrorsTotal.WithLabelValues(errType, component).Inc()
}
```

---

## Alerting Rules

### Prometheus Alerting

```yaml
# prometheus/alerts/linkrift.yml
groups:
  - name: linkrift-alerts
    rules:
      # High error rate
      - alert: HighErrorRate
        expr: |
          sum(rate(linkrift_http_requests_total{status=~"5.."}[5m]))
          / sum(rate(linkrift_http_requests_total[5m])) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }} over the last 5 minutes"

      # High redirect latency
      - alert: HighRedirectLatency
        expr: |
          histogram_quantile(0.99, sum(rate(linkrift_redirect_latency_seconds_bucket[5m])) by (le)) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High redirect latency"
          description: "P99 redirect latency is {{ $value | humanizeDuration }}"

      # Low cache hit rate
      - alert: LowCacheHitRate
        expr: |
          sum(rate(linkrift_cache_hits_total[5m]))
          / (sum(rate(linkrift_cache_hits_total[5m])) + sum(rate(linkrift_cache_misses_total[5m]))) < 0.8
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Low cache hit rate"
          description: "Cache hit rate is {{ $value | humanizePercentage }}"

      # Database connection saturation
      - alert: DatabaseConnectionsSaturated
        expr: linkrift_db_connections_active > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Database connections nearing limit"
          description: "Active connections: {{ $value }}"

      # High memory usage
      - alert: HighMemoryUsage
        expr: |
          go_memstats_heap_alloc_bytes{job="linkrift"}
          / go_memstats_heap_sys_bytes{job="linkrift"} > 0.9
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Heap usage is at {{ $value | humanizePercentage }}"

      # Service down
      - alert: ServiceDown
        expr: up{job="linkrift"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Linkrift service is down"
          description: "Service has been down for more than 1 minute"

      # High rate limiting
      - alert: HighRateLimiting
        expr: sum(rate(linkrift_rate_limit_hits_total[5m])) > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High rate of rate-limited requests"
          description: "{{ $value }} requests/s are being rate limited"
```

### Alert Severity Levels

| Severity | Response Time | Example Conditions |
|----------|---------------|-------------------|
| Critical | Immediate | Service down, >10% error rate |
| Warning | 15 minutes | High latency, low cache hit rate |
| Info | Next business day | Unusual traffic patterns |

### Notification Channels

```yaml
# alertmanager/config.yml
global:
  resolve_timeout: 5m
  slack_api_url: '${SLACK_WEBHOOK_URL}'

route:
  group_by: ['alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  receiver: 'default'
  routes:
    - match:
        severity: critical
      receiver: 'pagerduty-critical'
      continue: true
    - match:
        severity: warning
      receiver: 'slack-warnings'

receivers:
  - name: 'default'
    slack_configs:
      - channel: '#linkrift-alerts'
        title: '{{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'

  - name: 'pagerduty-critical'
    pagerduty_configs:
      - service_key: '${PAGERDUTY_SERVICE_KEY}'
        severity: critical

  - name: 'slack-warnings'
    slack_configs:
      - channel: '#linkrift-warnings'
        send_resolved: true
```

---

## OpenTelemetry Tracing

### Tracing Configuration

```go
// internal/tracing/tracing.go
package tracing

import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func InitTracer(ctx context.Context, serviceName, otlpEndpoint string) (*sdktrace.TracerProvider, error) {
    // Create OTLP exporter
    client := otlptracegrpc.NewClient(
        otlptracegrpc.WithEndpoint(otlpEndpoint),
        otlptracegrpc.WithInsecure(),
    )

    exporter, err := otlptrace.New(ctx, client)
    if err != nil {
        return nil, err
    }

    // Create resource
    res, err := resource.Merge(
        resource.Default(),
        resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName(serviceName),
            semconv.ServiceVersion("1.0.0"),
            attribute.String("environment", "production"),
        ),
    )
    if err != nil {
        return nil, err
    }

    // Create tracer provider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1)), // Sample 10%
    )

    // Set global tracer provider
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    return tp, nil
}
```

### Span Instrumentation

```go
// internal/tracing/spans.go
package tracing

import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("linkrift")

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
    return tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

// Example usage in redirect service
func (s *RedirectService) ResolveWithTracing(ctx context.Context, shortCode string) (string, error) {
    ctx, span := StartSpan(ctx, "redirect.resolve",
        attribute.String("short_code", shortCode),
    )
    defer span.End()

    // Check cache
    ctx, cacheSpan := StartSpan(ctx, "cache.get")
    url, err := s.cache.Get(ctx, shortCode)
    cacheSpan.End()

    if err == nil {
        span.SetAttributes(attribute.Bool("cache_hit", true))
        return url, nil
    }
    span.SetAttributes(attribute.Bool("cache_hit", false))

    // Query database
    ctx, dbSpan := StartSpan(ctx, "db.query")
    link, err := s.repo.GetByShortCode(ctx, shortCode)
    if err != nil {
        dbSpan.RecordError(err)
        dbSpan.SetStatus(codes.Error, err.Error())
        dbSpan.End()
        return "", err
    }
    dbSpan.End()

    span.SetAttributes(attribute.String("original_url", link.OriginalURL))
    return link.OriginalURL, nil
}
```

### Trace Correlation

```go
// internal/middleware/tracing.go
package middleware

import (
    "net/http"

    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
    "go.opentelemetry.io/otel/trace"
    "go.uber.org/zap"
    "linkrift/internal/logger"
)

// TracingMiddleware adds OpenTelemetry tracing to HTTP handlers
func TracingMiddleware(next http.Handler) http.Handler {
    return otelhttp.NewHandler(next, "http-request",
        otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
            return r.Method + " " + r.URL.Path
        }),
    )
}

// CorrelateTraceWithLogs adds trace and span IDs to the logger
func CorrelateTraceWithLogs(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        span := trace.SpanFromContext(ctx)

        if span.SpanContext().IsValid() {
            traceID := span.SpanContext().TraceID().String()
            spanID := span.SpanContext().SpanID().String()

            reqLogger := logger.FromContext(ctx).With(
                zap.String("trace_id", traceID),
                zap.String("span_id", spanID),
            )
            ctx = logger.WithLogger(ctx, reqLogger)
        }

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

## Profiling with pprof

### Enabling pprof

```go
// internal/server/pprof.go
package server

import (
    "net/http"
    "net/http/pprof"

    "github.com/go-chi/chi/v5"
)

func (s *Server) pprofRoutes() chi.Router {
    r := chi.NewRouter()

    // Require authentication for pprof endpoints
    r.Use(s.authMiddleware)

    r.HandleFunc("/", pprof.Index)
    r.HandleFunc("/cmdline", pprof.Cmdline)
    r.HandleFunc("/profile", pprof.Profile)
    r.HandleFunc("/symbol", pprof.Symbol)
    r.HandleFunc("/trace", pprof.Trace)

    r.Handle("/goroutine", pprof.Handler("goroutine"))
    r.Handle("/heap", pprof.Handler("heap"))
    r.Handle("/threadcreate", pprof.Handler("threadcreate"))
    r.Handle("/block", pprof.Handler("block"))
    r.Handle("/mutex", pprof.Handler("mutex"))
    r.Handle("/allocs", pprof.Handler("allocs"))

    return r
}

// Mount pprof routes on internal port only
func (s *Server) mountInternalRoutes() {
    s.internalRouter.Mount("/debug/pprof", s.pprofRoutes())
    s.internalRouter.Mount("/", s.metricsRoutes())
}
```

### Available Profiles

| Profile | Description | Use Case |
|---------|-------------|----------|
| `goroutine` | Stack traces of all current goroutines | Goroutine leaks |
| `heap` | Memory allocations of live objects | Memory leaks |
| `allocs` | Memory allocations (cumulative) | Memory allocation patterns |
| `threadcreate` | Stack traces leading to creation of new OS threads | Thread creation issues |
| `block` | Stack traces of goroutines blocked on synchronization | Contention issues |
| `mutex` | Stack traces of mutex contention | Lock contention |
| `profile` | CPU profile | CPU performance |
| `trace` | Execution trace | Latency analysis |

### Profiling Examples

```bash
# CPU profiling (30 seconds)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine analysis
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Block profiling (requires runtime.SetBlockProfileRate)
go tool pprof http://localhost:6060/debug/pprof/block

# Interactive analysis
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap

# Download and analyze offline
curl -o heap.prof http://localhost:6060/debug/pprof/heap
go tool pprof -http=:8080 heap.prof

# Trace analysis
curl -o trace.out 'http://localhost:6060/debug/pprof/trace?seconds=5'
go tool trace trace.out
```

```go
// Enable block and mutex profiling in main
package main

import (
    "runtime"
)

func init() {
    // Enable block profiling
    runtime.SetBlockProfileRate(1)

    // Enable mutex profiling
    runtime.SetMutexProfileFraction(1)
}
```

---

## Log Aggregation

### Structured Log Format

```json
{
  "level": "info",
  "timestamp": "2025-01-24T10:30:45.123Z",
  "caller": "service/redirect.go:42",
  "msg": "Redirect completed",
  "service": "linkrift",
  "version": "1.2.0",
  "request_id": "abc123-def456",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "short_code": "xyz789",
  "original_url": "https://example.com/long-url",
  "duration_ms": 2.5,
  "cache_hit": true
}
```

### Loki Integration

```yaml
# promtail/config.yml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: linkrift
    static_configs:
      - targets:
          - localhost
        labels:
          job: linkrift
          __path__: /var/log/linkrift/*.log
    pipeline_stages:
      - json:
          expressions:
            level: level
            timestamp: timestamp
            request_id: request_id
            trace_id: trace_id
      - labels:
          level:
      - timestamp:
          source: timestamp
          format: RFC3339Nano
```

---

## Best Practices

### Logging Best Practices

1. **Use structured logging** - Always use key-value pairs
2. **Include context** - Request ID, trace ID, user ID
3. **Log at appropriate levels** - Debug for details, Info for operations, Error for failures
4. **Avoid sensitive data** - Never log passwords, tokens, or PII
5. **Sample high-volume logs** - Consider sampling debug logs in production

### Metrics Best Practices

1. **Use meaningful names** - Follow Prometheus naming conventions
2. **Choose appropriate types** - Counter for totals, Gauge for values, Histogram for distributions
3. **Label wisely** - Avoid high-cardinality labels
4. **Document metrics** - Use Help text to describe each metric

### Alerting Best Practices

1. **Alert on symptoms, not causes** - Focus on user-facing impact
2. **Include runbooks** - Link to troubleshooting documentation
3. **Test alerts regularly** - Ensure alerts fire correctly
4. **Avoid alert fatigue** - Only alert on actionable conditions

---

## Related Documentation

- [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) - Debugging and issue resolution
- [MAINTENANCE.md](./MAINTENANCE.md) - Regular maintenance procedures
- [../architecture/INFRASTRUCTURE.md](../architecture/INFRASTRUCTURE.md) - Infrastructure overview
