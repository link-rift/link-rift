package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/link-rift/link-rift/internal/metrics"
)

// PrometheusMetrics returns a Gin middleware that records HTTP request metrics.
func PrometheusMetrics(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		statusCode := fmt.Sprintf("%d", c.Writer.Status())
		path := normalizeMetricPath(c.FullPath())

		// If FullPath is empty (no route matched), use a generic label.
		if path == "" {
			path = "unmatched"
		}

		metrics.HTTPRequestsTotal.WithLabelValues(serviceName, c.Request.Method, path, statusCode).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(serviceName, c.Request.Method, path).Observe(duration)
		metrics.HTTPResponseSize.WithLabelValues(serviceName, c.Request.Method, path).Observe(float64(c.Writer.Size()))

		if c.Request.ContentLength > 0 {
			metrics.HTTPRequestSize.WithLabelValues(serviceName, c.Request.Method, path).Observe(float64(c.Request.ContentLength))
		}
	}
}

// normalizeMetricPath normalizes route paths to prevent high-cardinality labels.
// Gin's FullPath() already returns the template (e.g. "/:shortCode") rather than
// the actual value, so this mainly handles edge cases.
func normalizeMetricPath(path string) string {
	if path == "" {
		return ""
	}
	return path
}
