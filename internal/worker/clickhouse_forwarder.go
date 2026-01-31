package worker

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/link-rift/link-rift/internal/models"
	"go.uber.org/zap"
)

// EnrichedClick holds parsed/enriched fields from click processing.
type EnrichedClick struct {
	CountryCode    string
	Region         string
	City           string
	Browser        string
	BrowserVersion string
	OS             string
	OSVersion      string
	DeviceType     string
	IsBot          bool
}

// ClickHouseForwarder writes enriched click events to ClickHouse for analytics.
// It is optional and best-effort: errors are logged but do not affect the PG pipeline.
type ClickHouseForwarder struct {
	conn   clickhouse.Conn
	logger *zap.Logger
}

// NewClickHouseForwarder creates a forwarder using the given ClickHouse connection.
// Returns nil, nil if conn is nil (opt-out).
func NewClickHouseForwarder(conn clickhouse.Conn, logger *zap.Logger) *ClickHouseForwarder {
	if conn == nil {
		return nil
	}
	return &ClickHouseForwarder{conn: conn, logger: logger}
}

// Forward inserts a single enriched click event into ClickHouse.
// This is best-effort: errors are logged but not returned.
func (f *ClickHouseForwarder) Forward(ctx context.Context, event *models.ClickEvent, enriched EnrichedClick) {
	var isBot uint8
	if enriched.IsBot {
		isBot = 1
	}

	err := f.conn.AsyncInsert(ctx,
		`INSERT INTO clicks (
			link_id, workspace_id, short_code, clicked_at, ip_address, user_agent, referer,
			country_code, region, city, browser, browser_version,
			os, os_version, device_type, is_bot
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		false,
		event.LinkID,
		event.WorkspaceID,
		event.ShortCode,
		event.Timestamp,
		event.IP,
		event.UserAgent,
		event.Referer,
		enriched.CountryCode,
		enriched.Region,
		enriched.City,
		enriched.Browser,
		enriched.BrowserVersion,
		enriched.OS,
		enriched.OSVersion,
		enriched.DeviceType,
		isBot,
	)
	if err != nil {
		f.logger.Warn("failed to forward click to ClickHouse",
			zap.Error(err),
			zap.String("link_id", event.LinkID.String()),
		)
	}
}

// ForwardBatch inserts multiple enriched click events into ClickHouse using a batch.
func (f *ClickHouseForwarder) ForwardBatch(ctx context.Context, events []*models.ClickEvent, enriched []EnrichedClick) {
	if len(events) == 0 {
		return
	}

	batch, err := f.conn.PrepareBatch(ctx,
		`INSERT INTO clicks (
			link_id, workspace_id, short_code, clicked_at, ip_address, user_agent, referer,
			country_code, region, city, browser, browser_version,
			os, os_version, device_type, is_bot
		)`,
	)
	if err != nil {
		f.logger.Warn("failed to prepare ClickHouse batch", zap.Error(err))
		return
	}

	for i, event := range events {
		e := enriched[i]
		var isBot uint8
		if e.IsBot {
			isBot = 1
		}

		if err := batch.Append(
			event.LinkID,
			event.WorkspaceID,
			event.ShortCode,
			event.Timestamp,
			event.IP,
			event.UserAgent,
			event.Referer,
			e.CountryCode,
			e.Region,
			e.City,
			e.Browser,
			e.BrowserVersion,
			e.OS,
			e.OSVersion,
			e.DeviceType,
			isBot,
		); err != nil {
			f.logger.Warn("failed to append to ClickHouse batch",
				zap.Error(err),
				zap.String("link_id", event.LinkID.String()),
			)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := batch.Send(); err != nil {
		f.logger.Warn("failed to send ClickHouse batch",
			zap.Error(err),
			zap.Int("count", len(events)),
		)
	}
}
