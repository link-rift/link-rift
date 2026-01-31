package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/link-rift/link-rift/internal/config"
	"go.uber.org/zap"
)

type PostgresDB struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewPostgres(cfg config.DatabaseConfig, logger *zap.Logger) (*PostgresDB, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing database URL: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.MaxIdleConns)
	poolCfg.MaxConnLifetime = cfg.ConnMaxLifetime
	poolCfg.MaxConnIdleTime = 5 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	logger.Info("connected to PostgreSQL",
		zap.String("host", poolCfg.ConnConfig.Host),
		zap.Int32("max_conns", poolCfg.MaxConns),
	)

	return &PostgresDB{pool: pool, logger: logger}, nil
}

func (db *PostgresDB) Pool() *pgxpool.Pool {
	return db.pool
}

func (db *PostgresDB) HealthCheck(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

func (db *PostgresDB) Close() {
	db.pool.Close()
	db.logger.Info("PostgreSQL connection closed")
}
