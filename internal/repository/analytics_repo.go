package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/models"
	"go.uber.org/zap"
)

// AnalyticsRepository defines the analytics data access interface.
type AnalyticsRepository interface {
	GetLinkStats(ctx context.Context, linkID uuid.UUID, dr models.DateRange) (*models.LinkAnalytics, error)
	GetWorkspaceStats(ctx context.Context, workspaceID uuid.UUID, dr models.DateRange) (*models.WorkspaceAnalytics, error)
	GetTimeSeries(ctx context.Context, linkID uuid.UUID, interval models.TimeSeriesInterval, dr models.DateRange) ([]models.TimeSeriesPoint, error)
	GetTopReferrers(ctx context.Context, linkID uuid.UUID, dr models.DateRange, limit int) ([]models.ReferrerStats, error)
	GetTopCountries(ctx context.Context, linkID uuid.UUID, dr models.DateRange, limit int) ([]models.CountryStats, error)
	GetDeviceBreakdown(ctx context.Context, linkID uuid.UUID, dr models.DateRange) (*models.DeviceBreakdown, error)
	GetBrowserBreakdown(ctx context.Context, linkID uuid.UUID, dr models.DateRange, limit int) ([]models.BrowserStats, error)
}

type clickhouseAnalyticsRepo struct {
	conn   clickhouse.Conn
	logger *zap.Logger
}

// NewClickHouseAnalyticsRepository creates an analytics repo backed by ClickHouse.
func NewClickHouseAnalyticsRepository(conn clickhouse.Conn, logger *zap.Logger) AnalyticsRepository {
	return &clickhouseAnalyticsRepo{conn: conn, logger: logger}
}

func (r *clickhouseAnalyticsRepo) GetLinkStats(ctx context.Context, linkID uuid.UUID, dr models.DateRange) (*models.LinkAnalytics, error) {
	now := time.Now().UTC()
	stats := &models.LinkAnalytics{}

	err := r.conn.QueryRow(ctx, `
		SELECT
			countIf(clicked_at >= $1 AND clicked_at <= $2) AS total_clicks,
			uniqExactIf(ip_address, clicked_at >= $1 AND clicked_at <= $2) AS unique_clicks,
			countIf(clicked_at >= $3) AS clicks_24h,
			countIf(clicked_at >= $4) AS clicks_7d,
			countIf(clicked_at >= $5) AS clicks_30d
		FROM clicks
		WHERE link_id = $6 AND is_bot = 0
	`,
		dr.Start, dr.End,
		now.Add(-24*time.Hour),
		now.Add(-7*24*time.Hour),
		now.Add(-30*24*time.Hour),
		linkID,
	).Scan(
		&stats.TotalClicks,
		&stats.UniqueClicks,
		&stats.Clicks24h,
		&stats.Clicks7d,
		&stats.Clicks30d,
	)
	if err != nil {
		return nil, fmt.Errorf("clickhouse get link stats: %w", err)
	}

	return stats, nil
}

func (r *clickhouseAnalyticsRepo) GetWorkspaceStats(ctx context.Context, workspaceID uuid.UUID, dr models.DateRange) (*models.WorkspaceAnalytics, error) {
	now := time.Now().UTC()
	stats := &models.WorkspaceAnalytics{}

	err := r.conn.QueryRow(ctx, `
		SELECT
			uniqExact(link_id) AS total_links,
			count() AS total_clicks,
			uniqExact(ip_address) AS unique_clicks,
			countIf(clicked_at >= $1) AS clicks_24h,
			countIf(clicked_at >= $2) AS clicks_7d,
			countIf(clicked_at >= $3) AS clicks_30d
		FROM clicks
		WHERE workspace_id = $4 AND clicked_at >= $5 AND clicked_at <= $6 AND is_bot = 0
	`,
		now.Add(-24*time.Hour),
		now.Add(-7*24*time.Hour),
		now.Add(-30*24*time.Hour),
		workspaceID,
		dr.Start, dr.End,
	).Scan(
		&stats.TotalLinks,
		&stats.TotalClicks,
		&stats.UniqueClicks,
		&stats.Clicks24h,
		&stats.Clicks7d,
		&stats.Clicks30d,
	)
	if err != nil {
		return nil, fmt.Errorf("clickhouse get workspace stats: %w", err)
	}

	// Top links
	rows, err := r.conn.Query(ctx, `
		SELECT link_id, any(short_code) AS short_code, count() AS clicks
		FROM clicks
		WHERE workspace_id = $1 AND clicked_at >= $2 AND clicked_at <= $3 AND is_bot = 0
		GROUP BY link_id
		ORDER BY clicks DESC
		LIMIT 10
	`, workspaceID, dr.Start, dr.End)
	if err != nil {
		return nil, fmt.Errorf("clickhouse get top links: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tl models.TopLink
		if err := rows.Scan(&tl.LinkID, &tl.ShortCode, &tl.TotalClicks); err != nil {
			return nil, fmt.Errorf("clickhouse scan top link: %w", err)
		}
		stats.TopLinks = append(stats.TopLinks, tl)
	}

	return stats, nil
}

func (r *clickhouseAnalyticsRepo) GetTimeSeries(ctx context.Context, linkID uuid.UUID, interval models.TimeSeriesInterval, dr models.DateRange) ([]models.TimeSeriesPoint, error) {
	fn := chTruncFunc(interval)

	rows, err := r.conn.Query(ctx, fmt.Sprintf(`
		SELECT
			%s(clicked_at) AS ts,
			count() AS clicks,
			uniqExact(ip_address) AS uniq
		FROM clicks
		WHERE link_id = $1 AND clicked_at >= $2 AND clicked_at <= $3 AND is_bot = 0
		GROUP BY ts
		ORDER BY ts ASC
	`, fn), linkID, dr.Start, dr.End)
	if err != nil {
		return nil, fmt.Errorf("clickhouse get time series: %w", err)
	}
	defer rows.Close()

	var points []models.TimeSeriesPoint
	for rows.Next() {
		var p models.TimeSeriesPoint
		if err := rows.Scan(&p.Timestamp, &p.Clicks, &p.Unique); err != nil {
			return nil, fmt.Errorf("clickhouse scan time series: %w", err)
		}
		points = append(points, p)
	}

	return points, nil
}

func (r *clickhouseAnalyticsRepo) GetTopReferrers(ctx context.Context, linkID uuid.UUID, dr models.DateRange, limit int) ([]models.ReferrerStats, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT
			if(referer = '', 'Direct', domain(referer)) AS ref,
			count() AS clicks
		FROM clicks
		WHERE link_id = $1 AND clicked_at >= $2 AND clicked_at <= $3 AND is_bot = 0
		GROUP BY ref
		ORDER BY clicks DESC
		LIMIT $4
	`, linkID, dr.Start, dr.End, limit)
	if err != nil {
		return nil, fmt.Errorf("clickhouse get referrers: %w", err)
	}
	defer rows.Close()

	var total int64
	var stats []models.ReferrerStats
	for rows.Next() {
		var s models.ReferrerStats
		if err := rows.Scan(&s.Referrer, &s.Clicks); err != nil {
			return nil, fmt.Errorf("clickhouse scan referrer: %w", err)
		}
		total += s.Clicks
		stats = append(stats, s)
	}

	for i := range stats {
		if total > 0 {
			stats[i].Percent = float64(stats[i].Clicks) / float64(total) * 100
		}
	}

	return stats, nil
}

func (r *clickhouseAnalyticsRepo) GetTopCountries(ctx context.Context, linkID uuid.UUID, dr models.DateRange, limit int) ([]models.CountryStats, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT
			if(country_code = '', 'Unknown', country_code) AS cc,
			count() AS clicks
		FROM clicks
		WHERE link_id = $1 AND clicked_at >= $2 AND clicked_at <= $3 AND is_bot = 0
		GROUP BY cc
		ORDER BY clicks DESC
		LIMIT $4
	`, linkID, dr.Start, dr.End, limit)
	if err != nil {
		return nil, fmt.Errorf("clickhouse get countries: %w", err)
	}
	defer rows.Close()

	var total int64
	var stats []models.CountryStats
	for rows.Next() {
		var s models.CountryStats
		if err := rows.Scan(&s.CountryCode, &s.Clicks); err != nil {
			return nil, fmt.Errorf("clickhouse scan country: %w", err)
		}
		s.Country = s.CountryCode // frontend maps code → name
		total += s.Clicks
		stats = append(stats, s)
	}

	for i := range stats {
		if total > 0 {
			stats[i].Percent = float64(stats[i].Clicks) / float64(total) * 100
		}
	}

	return stats, nil
}

func (r *clickhouseAnalyticsRepo) GetDeviceBreakdown(ctx context.Context, linkID uuid.UUID, dr models.DateRange) (*models.DeviceBreakdown, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT
			if(device_type = '', 'desktop', device_type) AS dt,
			count() AS clicks
		FROM clicks
		WHERE link_id = $1 AND clicked_at >= $2 AND clicked_at <= $3 AND is_bot = 0
		GROUP BY dt
	`, linkID, dr.Start, dr.End)
	if err != nil {
		return nil, fmt.Errorf("clickhouse get devices: %w", err)
	}
	defer rows.Close()

	breakdown := &models.DeviceBreakdown{}
	for rows.Next() {
		var dt string
		var clicks int64
		if err := rows.Scan(&dt, &clicks); err != nil {
			return nil, fmt.Errorf("clickhouse scan device: %w", err)
		}
		switch dt {
		case "desktop":
			breakdown.Desktop = clicks
		case "mobile":
			breakdown.Mobile = clicks
		case "tablet":
			breakdown.Tablet = clicks
		default:
			breakdown.Other += clicks
		}
	}

	return breakdown, nil
}

func (r *clickhouseAnalyticsRepo) GetBrowserBreakdown(ctx context.Context, linkID uuid.UUID, dr models.DateRange, limit int) ([]models.BrowserStats, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT
			if(browser = '', 'Unknown', browser) AS b,
			count() AS clicks
		FROM clicks
		WHERE link_id = $1 AND clicked_at >= $2 AND clicked_at <= $3 AND is_bot = 0
		GROUP BY b
		ORDER BY clicks DESC
		LIMIT $4
	`, linkID, dr.Start, dr.End, limit)
	if err != nil {
		return nil, fmt.Errorf("clickhouse get browsers: %w", err)
	}
	defer rows.Close()

	var total int64
	var stats []models.BrowserStats
	for rows.Next() {
		var s models.BrowserStats
		if err := rows.Scan(&s.Browser, &s.Clicks); err != nil {
			return nil, fmt.Errorf("clickhouse scan browser: %w", err)
		}
		total += s.Clicks
		stats = append(stats, s)
	}

	for i := range stats {
		if total > 0 {
			stats[i].Percent = float64(stats[i].Clicks) / float64(total) * 100
		}
	}

	return stats, nil
}

func chTruncFunc(interval models.TimeSeriesInterval) string {
	switch interval {
	case models.IntervalHour:
		return "toStartOfHour"
	case models.IntervalDay:
		return "toStartOfDay"
	case models.IntervalWeek:
		return "toStartOfWeek"
	case models.IntervalMonth:
		return "toStartOfMonth"
	default:
		return "toStartOfDay"
	}
}
