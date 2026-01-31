package database

import (
	"context"
	"fmt"
	"time"

	"github.com/link-rift/link-rift/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisDB struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRedis(cfg config.RedisConfig, logger *zap.Logger) (*RedisDB, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis URL: %w", err)
	}

	if cfg.Password != "" {
		opts.Password = cfg.Password
	}
	opts.DB = cfg.DB

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	logger.Info("connected to Redis",
		zap.String("addr", opts.Addr),
		zap.Int("db", opts.DB),
	)

	return &RedisDB{client: client, logger: logger}, nil
}

func (db *RedisDB) Client() *redis.Client {
	return db.client
}

func (db *RedisDB) HealthCheck(ctx context.Context) error {
	return db.client.Ping(ctx).Err()
}

func (db *RedisDB) Close() {
	if err := db.client.Close(); err != nil {
		db.logger.Error("error closing Redis connection", zap.Error(err))
	}
	db.logger.Info("Redis connection closed")
}
