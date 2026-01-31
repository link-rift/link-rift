package database

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/link-rift/link-rift/internal/config"
	"go.uber.org/zap"
)

type ClickHouseDB struct {
	conn   clickhouse.Conn
	logger *zap.Logger
}

func NewClickHouse(cfg config.ClickHouseConfig, logger *zap.Logger) (*ClickHouseDB, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.URL},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.User,
			Password: cfg.Password,
		},
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 10 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("opening ClickHouse connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging ClickHouse: %w", err)
	}

	logger.Info("connected to ClickHouse",
		zap.String("addr", cfg.URL),
		zap.String("database", cfg.Database),
	)

	return &ClickHouseDB{conn: conn, logger: logger}, nil
}

func (db *ClickHouseDB) Conn() clickhouse.Conn {
	return db.conn
}

func (db *ClickHouseDB) HealthCheck(ctx context.Context) error {
	return db.conn.Ping(ctx)
}

func (db *ClickHouseDB) Close() {
	if err := db.conn.Close(); err != nil {
		db.logger.Error("error closing ClickHouse connection", zap.Error(err))
	}
	db.logger.Info("ClickHouse connection closed")
}
