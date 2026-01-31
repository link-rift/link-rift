package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/link-rift/link-rift/internal/config"
	"github.com/link-rift/link-rift/internal/database"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/redirect"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/crypto"
	"go.uber.org/zap"
)

var passwordPageTmpl = template.Must(template.New("password").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Password Required - Linkrift</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f9fafb; display: flex; align-items: center; justify-content: center; min-height: 100vh; }
    .card { background: white; border-radius: 12px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); padding: 2rem; max-width: 400px; width: 90%; }
    h1 { font-size: 1.25rem; margin-bottom: 0.5rem; color: #111827; }
    p { font-size: 0.875rem; color: #6b7280; margin-bottom: 1.5rem; }
    .error { color: #dc2626; font-size: 0.875rem; margin-bottom: 1rem; }
    input { width: 100%; padding: 0.625rem 0.75rem; border: 1px solid #d1d5db; border-radius: 6px; font-size: 0.875rem; margin-bottom: 1rem; outline: none; }
    input:focus { border-color: #2563eb; box-shadow: 0 0 0 2px rgba(37,99,235,0.15); }
    button { width: 100%; padding: 0.625rem; background: #2563eb; color: white; border: none; border-radius: 6px; font-size: 0.875rem; font-weight: 500; cursor: pointer; }
    button:hover { background: #1d4ed8; }
  </style>
</head>
<body>
  <div class="card">
    <h1>Password Required</h1>
    <p>This link is password protected. Enter the password to continue.</p>
    {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
    <form method="POST" action="/{{.ShortCode}}/verify">
      <input type="password" name="password" placeholder="Enter password" required autofocus>
      <button type="submit">Continue</button>
    </form>
  </div>
</body>
</html>`))

var errorPageTmpl = template.Must(template.New("error").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Title}} - Linkrift</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f9fafb; display: flex; align-items: center; justify-content: center; min-height: 100vh; }
    .card { background: white; border-radius: 12px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); padding: 2rem; max-width: 400px; width: 90%; text-align: center; }
    h1 { font-size: 1.5rem; margin-bottom: 0.5rem; color: #111827; }
    p { font-size: 0.875rem; color: #6b7280; }
  </style>
</head>
<body>
  <div class="card">
    <h1>{{.Title}}</h1>
    <p>{{.Message}}</p>
  </div>
</body>
</html>`))

func main() {
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Init logger
	var logger *zap.Logger
	if cfg.App.Env == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 3. Connect PostgreSQL (read pool)
	pgDB, err := database.NewPostgres(cfg.Database, logger)
	if err != nil {
		logger.Fatal("failed to connect to PostgreSQL", zap.Error(err))
	}
	defer pgDB.Close()

	// 4. Connect Redis
	redisDB, err := database.NewRedis(cfg.Redis, logger)
	if err != nil {
		logger.Fatal("failed to connect to Redis", zap.Error(err))
	}
	defer redisDB.Close()

	// 5. Create dependencies
	queries := sqlc.New(pgDB.Pool())
	linkRepo := repository.NewLinkRepository(queries, logger)

	cache := redirect.NewCache(
		redisDB.Client(),
		cfg.Redirect.LocalCacheTTL,
		cfg.Redirect.RedisCacheTTL,
		logger,
	)
	resolver := redirect.NewResolver(cache, linkRepo, logger)
	tracker := redirect.NewClickTracker(
		redisDB.Client(),
		cfg.Redirect.TrackerBuffer,
		cfg.Redirect.TrackerFlush,
		logger,
	)
	botDetector := redirect.NewBotDetector()
	ruleEngine := redirect.NewRuleEngine(queries, logger)

	// 6. Create Gin router in release mode
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 7. Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "linkrift-redirect",
		})
	})

	// 8. Password verification endpoint
	router.POST("/:shortCode/verify", func(c *gin.Context) {
		shortCode := c.Param("shortCode")
		password := c.PostForm("password")

		result, err := resolver.Resolve(c.Request.Context(), shortCode)
		if err != nil {
			renderError(c, http.StatusNotFound, "Link Not Found", "The link you're looking for doesn't exist.")
			return
		}

		if !result.HasPassword {
			c.Redirect(http.StatusFound, result.DestinationURL)
			return
		}

		match, err := crypto.VerifyPassword(password, result.PasswordHash)
		if err != nil || !match {
			passwordPageTmpl.Execute(c.Writer, map[string]interface{}{
				"ShortCode": shortCode,
				"Error":     "Incorrect password. Please try again.",
			})
			return
		}

		// Track click
		if !botDetector.IsBot(c.Request.UserAgent()) {
			tracker.Track(&models.ClickEvent{
				LinkID:    result.LinkID,
				ShortCode: result.ShortCode,
				IP:        c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
				Referer:   c.Request.Referer(),
				Timestamp: time.Now(),
			})
		}

		c.Redirect(http.StatusFound, result.DestinationURL)
	})

	// 9. Preview handler (shortCode+)
	router.GET("/:shortCode/preview", func(c *gin.Context) {
		shortCode := c.Param("shortCode")

		result, err := resolver.Resolve(c.Request.Context(), shortCode)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"short_code":      result.ShortCode,
			"destination_url": result.DestinationURL,
			"is_active":       result.IsActive,
			"has_password":    result.HasPassword,
			"is_expired":      result.IsExpired,
		})
	})

	// 10. Main redirect handler
	router.GET("/:shortCode", func(c *gin.Context) {
		shortCode := c.Param("shortCode")

		// Skip common favicon/robot requests
		if shortCode == "favicon.ico" || shortCode == "robots.txt" {
			c.Status(http.StatusNotFound)
			return
		}

		result, err := resolver.Resolve(c.Request.Context(), shortCode)
		if err != nil {
			renderError(c, http.StatusNotFound, "Link Not Found", "The link you're looking for doesn't exist or has been removed.")
			return
		}

		// Check if active
		if !result.IsActive {
			renderError(c, http.StatusGone, "Link Disabled", "This link has been disabled by its owner.")
			return
		}

		// Check if expired
		if result.IsExpired {
			renderError(c, http.StatusGone, "Link Expired", "This link has expired and is no longer available.")
			return
		}

		// Check click limit
		if result.IsOverLimit {
			renderError(c, http.StatusGone, "Link Limit Reached", "This link has reached its maximum number of clicks.")
			return
		}

		// Password protected — show form
		if result.HasPassword {
			// Check for auth cookie
			cookie, err := c.Cookie("lr_auth_" + shortCode)
			if err != nil || cookie != "1" {
				c.Header("Content-Type", "text/html; charset=utf-8")
				c.Status(http.StatusOK)
				passwordPageTmpl.Execute(c.Writer, map[string]interface{}{
					"ShortCode": shortCode,
				})
				return
			}
		}

		// Evaluate conditional redirect rules
		destinationURL := result.DestinationURL
		if ruleURL, matched := ruleEngine.Evaluate(c.Request.Context(), result.LinkID, c.Request); matched {
			destinationURL = ruleURL
		}

		// Track click (non-blocking, skip bots)
		if !botDetector.IsBot(c.Request.UserAgent()) {
			tracker.Track(&models.ClickEvent{
				LinkID:    result.LinkID,
				ShortCode: result.ShortCode,
				IP:        c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
				Referer:   c.Request.Referer(),
				Timestamp: time.Now(),
			})
		}

		// Append UTM params if the destination doesn't already have them
		c.Redirect(http.StatusFound, destinationURL)
	})

	// 11. Start server with graceful shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Redirect.Port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		logger.Info("starting redirect server",
			zap.Int("port", cfg.Redirect.Port),
			zap.String("env", cfg.App.Env),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("redirect server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down redirect server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Flush tracker before shutdown
	tracker.Shutdown(ctx)

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("redirect server forced to shutdown", zap.Error(err))
	}

	logger.Info("redirect server stopped")
}

func renderError(c *gin.Context, status int, title, message string) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(status)
	errorPageTmpl.Execute(c.Writer, map[string]string{
		"Title":   title,
		"Message": message,
	})
}

