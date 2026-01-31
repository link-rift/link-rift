package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/link-rift/link-rift/internal/config"
	"github.com/link-rift/link-rift/internal/database"
	"github.com/link-rift/link-rift/internal/handler"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/middleware"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/internal/service"
	"github.com/link-rift/link-rift/pkg/paseto"
	"go.uber.org/zap"
)

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

	// 3. Connect PostgreSQL
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

	// 5. Create sqlc queries
	queries := sqlc.New(pgDB.Pool())

	// 6. Create PASETO token maker
	tokenMaker, err := paseto.NewPasetoMaker(cfg.Auth.TokenSecret)
	if err != nil {
		logger.Fatal("failed to create token maker", zap.Error(err))
	}

	// 7. Initialize license system
	licVerifier, err := license.NewVerifier()
	if err != nil {
		logger.Fatal("failed to create license verifier", zap.Error(err))
	}

	licManager := license.NewManager(licVerifier, logger)
	if cfg.License.Key != "" {
		if err := licManager.LoadLicense(cfg.License.Key); err != nil {
			logger.Warn("failed to load license key, running as community edition", zap.Error(err))
		} else {
			logger.Info("license loaded",
				zap.String("tier", string(licManager.GetTier())),
			)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			licManager.StartPeriodicCheck(ctx, cfg.License.CheckInterval)
		}
	} else {
		logger.Info("no license key configured, running as community edition")
	}

	// 8. Create repositories
	userRepo := repository.NewUserRepository(queries, logger)
	sessionRepo := repository.NewSessionRepository(queries, logger)
	resetRepo := repository.NewPasswordResetRepository(queries, logger)

	// 9. Create services
	authService := service.NewAuthService(
		userRepo, sessionRepo, resetRepo,
		tokenMaker, pgDB.Pool(), redisDB.Client(),
		cfg, logger,
	)

	// 10. Create handlers
	authHandler := handler.NewAuthHandler(authService, logger)
	licenseHandler := handler.NewLicenseHandler(licManager, logger)

	// 11. Create Gin router
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.App.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 12. Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "linkrift-api",
		})
	})

	// 13. API v1 routes
	v1 := router.Group("/api/v1")
	authMw := middleware.RequireAuth(tokenMaker, userRepo)
	authHandler.RegisterRoutes(v1, authMw)
	licenseHandler.RegisterRoutes(v1, authMw)

	// 14. Start server with graceful shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting API server",
			zap.Int("port", cfg.App.Port),
			zap.String("env", cfg.App.Env),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server stopped")
}
