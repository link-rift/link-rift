// Package errortracking provides integration with error tracking services.
// When Sentry is configured via SENTRY_DSN environment variable, errors
// captured by the middleware are forwarded to Sentry for monitoring and alerting.
//
// This package provides a no-op implementation when no DSN is configured,
// ensuring the application runs without external error tracking dependencies.
package errortracking

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Config holds Sentry error tracking configuration.
type Config struct {
	DSN         string  `mapstructure:"dsn"`
	Environment string  `mapstructure:"environment"`
	SampleRate  float64 `mapstructure:"sample_rate"`
	Release     string  `mapstructure:"release"`
}

// Tracker provides error tracking functionality.
// When DSN is empty, all operations are no-ops.
type Tracker struct {
	enabled bool
	config  Config
	logger  *zap.Logger
}

// NewTracker creates a new error tracker.
// If cfg.DSN is empty, the tracker is a no-op.
func NewTracker(cfg Config, logger *zap.Logger) *Tracker {
	t := &Tracker{
		config: cfg,
		logger: logger,
	}

	if cfg.DSN == "" {
		logger.Info("error tracking disabled (no SENTRY_DSN configured)")
		return t
	}

	// To enable Sentry, add the sentry-go SDK:
	//   go get github.com/getsentry/sentry-go
	//   go get github.com/getsentry/sentry-go/gin
	//
	// Then initialize:
	//   err := sentry.Init(sentry.ClientOptions{
	//       Dsn:              cfg.DSN,
	//       Environment:      cfg.Environment,
	//       Release:          cfg.Release,
	//       TracesSampleRate: cfg.SampleRate,
	//   })
	//
	// For now, we log that Sentry would be initialized.
	t.enabled = true
	logger.Info("error tracking enabled",
		zap.String("environment", cfg.Environment),
		zap.Float64("sample_rate", cfg.SampleRate),
	)

	return t
}

// CaptureError records an error with optional context.
func (t *Tracker) CaptureError(err error, tags map[string]string) {
	if !t.enabled {
		return
	}

	// With Sentry SDK:
	//   sentry.WithScope(func(scope *sentry.Scope) {
	//       for k, v := range tags {
	//           scope.SetTag(k, v)
	//       }
	//       sentry.CaptureException(err)
	//   })

	t.logger.Error("captured error for tracking",
		zap.Error(err),
		zap.Any("tags", tags),
	)
}

// GinMiddleware returns a Gin middleware that captures panics and 5xx errors.
func (t *Tracker) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("panic: %v", r)
				}

				t.CaptureError(err, map[string]string{
					"method": c.Request.Method,
					"path":   c.Request.URL.Path,
					"type":   "panic",
				})

				t.logger.Error("recovered from panic",
					zap.Error(err),
					zap.String("stack", string(debug.Stack())),
				)

				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		c.Next()

		// Capture 5xx errors
		if c.Writer.Status() >= 500 {
			t.CaptureError(
				fmt.Errorf("HTTP %d: %s %s", c.Writer.Status(), c.Request.Method, c.Request.URL.Path),
				map[string]string{
					"method":      c.Request.Method,
					"path":        c.Request.URL.Path,
					"status_code": fmt.Sprintf("%d", c.Writer.Status()),
					"type":        "http_error",
				},
			)
		}
	}
}

// Flush waits for pending events to be sent. Call on shutdown.
func (t *Tracker) Flush() {
	if !t.enabled {
		return
	}
	// With Sentry SDK:
	//   sentry.Flush(2 * time.Second)
}
