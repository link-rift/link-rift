package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "linkrift"
)

// HTTP metrics for API and redirect services.
var (
	// HTTPRequestsTotal counts total HTTP requests by method, path, and status code.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"service", "method", "path", "status_code"},
	)

	// HTTPRequestDuration tracks request duration in seconds.
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"service", "method", "path"},
	)

	// HTTPRequestSize tracks request body size.
	HTTPRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_size_bytes",
			Help:      "HTTP request body size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 6),
		},
		[]string{"service", "method", "path"},
	)

	// HTTPResponseSize tracks response body size.
	HTTPResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_response_size_bytes",
			Help:      "HTTP response body size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 6),
		},
		[]string{"service", "method", "path"},
	)
)

// Redirect-specific metrics.
var (
	// RedirectsTotal counts total redirects by status (success, not_found, expired, etc.).
	RedirectsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "redirect",
			Name:      "total",
			Help:      "Total number of redirect operations by status",
		},
		[]string{"status"},
	)

	// RedirectLatency tracks redirect resolution latency in seconds.
	RedirectLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "redirect",
			Name:      "latency_seconds",
			Help:      "Redirect resolution latency in seconds",
			Buckets:   []float64{0.0001, 0.00025, 0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
		[]string{"cache"},
	)

	// CacheHitsTotal counts cache hits and misses.
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "cache",
			Name:      "hits_total",
			Help:      "Total cache hits and misses",
		},
		[]string{"cache_type", "result"},
	)
)

// Link metrics.
var (
	// LinksCreatedTotal counts link creation events.
	LinksCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "links",
			Name:      "created_total",
			Help:      "Total number of links created",
		},
	)

	// ActiveLinksGauge tracks the number of active links (set periodically).
	ActiveLinksGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "links",
			Name:      "active",
			Help:      "Number of currently active links",
		},
	)
)

// Database metrics.
var (
	// DBQueryDuration tracks database query duration.
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "db",
			Name:      "query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"operation"},
	)

	// DBActiveConnections tracks the number of active database connections.
	DBActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "db",
			Name:      "active_connections",
			Help:      "Number of active database connections",
		},
	)
)

// Auth metrics.
var (
	// AuthAttemptsTotal counts authentication attempts by type and outcome.
	AuthAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "auth",
			Name:      "attempts_total",
			Help:      "Total authentication attempts",
		},
		[]string{"type", "result"},
	)
)

// WebSocket metrics.
var (
	// WSActiveConnections tracks the number of active WebSocket connections.
	WSActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "ws",
			Name:      "active_connections",
			Help:      "Number of active WebSocket connections",
		},
	)
)

// Worker metrics.
var (
	// WorkerJobsProcessed counts processed background jobs.
	WorkerJobsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "worker",
			Name:      "jobs_processed_total",
			Help:      "Total number of background jobs processed",
		},
		[]string{"task_type", "result"},
	)

	// WorkerJobDuration tracks background job processing duration.
	WorkerJobDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "worker",
			Name:      "job_duration_seconds",
			Help:      "Background job processing duration in seconds",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30, 60},
		},
		[]string{"task_type"},
	)
)

// BuildInfo exposes build metadata.
var BuildInfo = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "build_info",
		Help:      "Build information",
	},
	[]string{"version", "go_version"},
)
