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
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/qrcode"
	"github.com/link-rift/link-rift/internal/realtime"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/internal/service"
	"github.com/link-rift/link-rift/pkg/paseto"
	"github.com/link-rift/link-rift/pkg/storage"
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

	// 8. Connect ClickHouse (optional — analytics)
	var analyticsRepo repository.AnalyticsRepository
	if cfg.ClickHouse.URL != "" {
		chDB, err := database.NewClickHouse(cfg.ClickHouse, logger)
		if err != nil {
			logger.Warn("ClickHouse unavailable, using PostgreSQL for analytics", zap.Error(err))
			analyticsRepo = repository.NewPGAnalyticsRepository(pgDB.Pool(), logger)
		} else {
			defer chDB.Close()
			analyticsRepo = repository.NewClickHouseAnalyticsRepository(chDB.Conn(), logger)
		}
	} else {
		analyticsRepo = repository.NewPGAnalyticsRepository(pgDB.Pool(), logger)
	}

	// 9. Create repositories
	userRepo := repository.NewUserRepository(queries, logger)
	sessionRepo := repository.NewSessionRepository(queries, logger)
	resetRepo := repository.NewPasswordResetRepository(queries, logger)
	linkRepo := repository.NewLinkRepository(queries, logger)
	clickRepo := repository.NewClickRepository(queries, logger)
	workspaceRepo := repository.NewWorkspaceRepository(queries, logger)
	memberRepo := repository.NewWorkspaceMemberRepository(queries, logger)
	domainRepo := repository.NewDomainRepository(queries, logger)
	qrCodeRepo := repository.NewQRCodeRepository(queries, logger)
	bioPageRepo := repository.NewBioPageRepository(queries, logger)
	apiKeyRepo := repository.NewAPIKeyRepository(queries, logger)
	webhookRepo := repository.NewWebhookRepository(queries, logger)

	// 9b. Create storage client (local fallback for development)
	var objectStore storage.ObjectStorage
	if cfg.S3.Endpoint != "" && cfg.S3.AccessKey != "" {
		s3Store, err := storage.NewS3Storage(cfg.S3)
		if err != nil {
			logger.Warn("S3 storage unavailable, falling back to local storage", zap.Error(err))
			objectStore = storage.NewLocalStorage("./data/uploads/", cfg.App.BaseURL+"/uploads/")
		} else {
			objectStore = s3Store
		}
	} else {
		objectStore = storage.NewLocalStorage("./data/uploads/", cfg.App.BaseURL+"/uploads/")
	}

	// 9c. Create QR code generator
	qrGenerator := qrcode.NewGenerator(objectStore)
	qrBatchGenerator := qrcode.NewBatchGenerator(qrGenerator, 4)

	// 10. Create event publisher for webhooks
	eventPublisher := service.NewEventPublisher(redisDB.Client(), logger)

	// Create services
	authService := service.NewAuthService(
		userRepo, sessionRepo, resetRepo,
		tokenMaker, pgDB.Pool(), redisDB.Client(),
		cfg, logger,
	)
	linkService := service.NewLinkService(linkRepo, clickRepo, pgDB.Pool(), redisDB.Client(), cfg, eventPublisher, logger)
	workspaceService := service.NewWorkspaceService(workspaceRepo, memberRepo, userRepo, licManager, eventPublisher, pgDB.Pool(), logger)
	analyticsService := service.NewAnalyticsService(analyticsRepo, clickRepo, licManager, logger)
	sslProvider := service.NewMockSSLProvider()
	domainService := service.NewDomainService(domainRepo, licManager, sslProvider, cfg, eventPublisher, logger)
	qrService := service.NewQRCodeService(qrCodeRepo, linkRepo, qrGenerator, qrBatchGenerator, objectStore, licManager, cfg, logger)
	bioPageService := service.NewBioPageService(bioPageRepo, licManager, eventPublisher, logger)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, licManager, redisDB.Client(), logger)
	webhookService := service.NewWebhookService(webhookRepo, licManager, logger)

	// 11. Create handlers
	authHandler := handler.NewAuthHandler(authService, logger)
	licenseHandler := handler.NewLicenseHandler(licManager, logger)
	linkHandler := handler.NewLinkHandler(linkService, logger)
	workspaceHandler := handler.NewWorkspaceHandler(workspaceService, logger)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService, linkService, logger)
	domainHandler := handler.NewDomainHandler(domainService, logger)
	qrHandler := handler.NewQRHandler(qrService, logger)
	bioPageHandler := handler.NewBioPageHandler(bioPageService, logger)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyService, logger)
	webhookHandler := handler.NewWebhookHandler(webhookService, logger)

	// WebSocket real-time hub
	wsHub := realtime.NewHub(logger)
	go wsHub.Run()
	wsHandler := handler.NewWebSocketHandler(wsHub, tokenMaker, memberRepo, logger)

	// Start Redis subscriber for real-time click notifications
	realtimeCtx, realtimeCancel := context.WithCancel(context.Background())
	defer realtimeCancel()
	realtime.StartRedisSubscriber(realtimeCtx, redisDB.Client(), wsHub, logger)

	// 12. Create Gin router
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.App.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "X-RateLimit-Reset-After"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 13. Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "linkrift-api",
		})
	})

	// 14. API v1 routes
	v1 := router.Group("/api/v1")
	authMw := middleware.RequireAuth(tokenMaker, userRepo)
	authHandler.RegisterRoutes(v1, authMw)
	licenseHandler.RegisterRoutes(v1, authMw)

	// Workspace routes
	wsAccessMw := middleware.RequireWorkspaceAccess(workspaceRepo, memberRepo)
	workspaceHandler.RegisterRoutes(v1, authMw, wsAccessMw)

	// API key auth middleware (processes X-API-Key header before session auth)
	apiKeyAuthMw := middleware.APIKeyAuth(apiKeyService, userRepo, workspaceRepo, memberRepo)

	// Link routes now live under /api/v1/workspaces/:workspaceId/links
	wsScoped := v1.Group("/workspaces/:workspaceId", authMw, wsAccessMw)
	editorMw := middleware.RequireWorkspaceRole(models.RoleEditor)
	adminMw := middleware.RequireWorkspaceRole(models.RoleAdmin)
	linkHandler.RegisterRoutes(wsScoped, editorMw)
	domainHandler.RegisterRoutes(wsScoped, editorMw)
	qrHandler.RegisterRoutes(wsScoped, editorMw)
	bioPageHandler.RegisterRoutes(wsScoped, editorMw)
	analyticsHandler.RegisterRoutes(wsScoped)
	apiKeyHandler.RegisterRoutes(wsScoped, adminMw)
	webhookHandler.RegisterRoutes(wsScoped, adminMw)

	// API key authenticated routes (alternative auth for programmatic access)
	apiScoped := v1.Group("/workspaces/:workspaceId", apiKeyAuthMw, wsAccessMw)
	linkHandler.RegisterRoutes(apiScoped, editorMw)

	// Public bio page routes (no auth)
	bioPageHandler.RegisterPublicRoutes(router)

	// WebSocket endpoint (outside API group, no auth middleware — auth via query param)
	wsHandler.RegisterRoutes(router)

	// 15. Start server with graceful shutdown
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
