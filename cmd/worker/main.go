package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/link-rift/link-rift/internal/config"
	"github.com/link-rift/link-rift/internal/database"
	"github.com/link-rift/link-rift/internal/redirect"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/internal/service"
	"github.com/link-rift/link-rift/internal/worker"
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

	// 5. Create dependencies
	queries := sqlc.New(pgDB.Pool())
	clickRepo := repository.NewClickRepository(queries, logger)
	linkRepo := repository.NewLinkRepository(queries, logger)
	webhookRepo := repository.NewWebhookRepository(queries, logger)
	botDetector := redirect.NewBotDetector()

	// 5b. Create event publisher for webhook events
	eventPublisher := service.NewEventPublisher(redisDB.Client(), logger)

	// 6. Create and start click processor
	processor := worker.NewClickProcessor(
		redisDB.Client(),
		clickRepo,
		linkRepo,
		botDetector,
		logger,
	)
	processor.SetEventPublisher(eventPublisher)

	// 6b. Create and start webhook delivery processor
	webhookProcessor := worker.NewWebhookDeliveryProcessor(
		redisDB.Client(),
		webhookRepo,
		logger,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go processor.Start(ctx)
	go webhookProcessor.Start(ctx)

	logger.Info("worker started, processing click events and webhook deliveries")

	// 7. Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down worker...")
	processor.Stop()
	webhookProcessor.Stop()
	cancel()

	logger.Info("worker stopped")
}
