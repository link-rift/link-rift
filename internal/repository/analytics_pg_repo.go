package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/link-rift/link-rift/internal/models"
	"go.uber.org/zap"
)

type pgAnalyticsRepo struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewPGAnalyticsRepository creates an analytics repo backed by PostgreSQL.
// Used as fallback when ClickHouse is not configured.
func NewPGAnalyticsRepository(pool *pgxpool.Pool, logger *zap.Logger) AnalyticsRepository {
	return &pgAnalyticsRepo{pool: pool, logger: logger}
}

func (r *pgAnalyticsRepo) GetLinkStats(ctx context.Context, linkID uuid.UUID, dr models.DateRange) (*models.LinkAnalytics, error) {
	now := time.Now().UTC()
	stats := &models.LinkAnalytics{}

	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE clicked_at >= $1 AND clicked_at <= $2) AS total_clicks,
			COUNT(DISTINCT ip_address) FILTER (WHERE clicked_at >= $1 AND clicked_at <= $2) AS unique_clicks,
			COUNT(*) FILTER (WHERE clicked_at >= $3) AS clicks_24h,
			COUNT(*) FILTER (WHERE clicked_at >= $4) AS clicks_7d,
			COUNT(*) FILTER (WHERE clicked_at >= $5) AS clicks_30d
		FROM clicks
		WHERE link_id = $6 AND is_bot = false
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
		return nil, fmt.Errorf("pg get link stats: %w", err)
	}

	return stats, nil
}

func (r *pgAnalyticsRepo) GetWorkspaceStats(ctx context.Context, workspaceID uuid.UUID, dr models.DateRange) (*models.WorkspaceAnalytics, error) {
	now := time.Now().UTC()
	stats := &models.WorkspaceAnalytics{}

	// Get link IDs for this workspace
	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT c.link_id),
			COUNT(*),
			COUNT(DISTINCT c.ip_address),
			COUNT(*) FILTER (WHERE c.clicked_at >= $1),
			COUNT(*) FILTER (WHERE c.clicked_at >= $2),
			COUNT(*) FILTER (WHERE c.clicked_at >= $3)
		FROM clicks c
		JOIN links l ON l.id = c.link_id
		WHERE l.workspace_id = $4
			AND c.clicked_at >= $5 AND c.clicked_at <= $6
			AND c.is_bot = false
			AND l.deleted_at IS NULL
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
		return nil, fmt.Errorf("pg get workspace stats: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT c.link_id, l.short_code, COUNT(*) AS clicks
		FROM clicks c
		JOIN links l ON l.id = c.link_id
		WHERE l.workspace_id = $1
			AND c.clicked_at >= $2 AND c.clicked_at <= $3
			AND c.is_bot = false
			AND l.deleted_at IS NULL
		GROUP BY c.link_id, l.short_code
		ORDER BY clicks DESC
		LIMIT 10
	`, workspaceID, dr.Start, dr.End)
	if err != nil {
		return nil, fmt.Errorf("pg get top links: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tl models.TopLink
		if err := rows.Scan(&tl.LinkID, &tl.ShortCode, &tl.TotalClicks); err != nil {
			return nil, fmt.Errorf("pg scan top link: %w", err)
		}
		stats.TopLinks = append(stats.TopLinks, tl)
	}

	return stats, nil
}

func (r *pgAnalyticsRepo) GetTimeSeries(ctx context.Context, linkID uuid.UUID, interval models.TimeSeriesInterval, dr models.DateRange) ([]models.TimeSeriesPoint, error) {
	trunc := pgTruncInterval(interval)

	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT
			date_trunc('%s', clicked_at) AS ts,
			COUNT(*) AS clicks,
			COUNT(DISTINCT ip_address) AS uniq
		FROM clicks
		WHERE link_id = $1 AND clicked_at >= $2 AND clicked_at <= $3 AND is_bot = false
		GROUP BY ts
		ORDER BY ts ASC
	`, trunc), linkID, dr.Start, dr.End)
	if err != nil {
		return nil, fmt.Errorf("pg get time series: %w", err)
	}
	defer rows.Close()

	var points []models.TimeSeriesPoint
	for rows.Next() {
		var p models.TimeSeriesPoint
		if err := rows.Scan(&p.Timestamp, &p.Clicks, &p.Unique); err != nil {
			return nil, fmt.Errorf("pg scan time series: %w", err)
		}
		points = append(points, p)
	}

	return points, nil
}

func (r *pgAnalyticsRepo) GetTopReferrers(ctx context.Context, linkID uuid.UUID, dr models.DateRange, limit int) ([]models.ReferrerStats, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			COALESCE(NULLIF(referer, ''), 'Direct') AS ref,
			COUNT(*) AS clicks
		FROM clicks
		WHERE link_id = $1 AND clicked_at >= $2 AND clicked_at <= $3 AND is_bot = false
		GROUP BY ref
		ORDER BY clicks DESC
		LIMIT $4
	`, linkID, dr.Start, dr.End, limit)
	if err != nil {
		return nil, fmt.Errorf("pg get referrers: %w", err)
	}
	defer rows.Close()

	var total int64
	var stats []models.ReferrerStats
	for rows.Next() {
		var s models.ReferrerStats
		if err := rows.Scan(&s.Referrer, &s.Clicks); err != nil {
			return nil, fmt.Errorf("pg scan referrer: %w", err)
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

func (r *pgAnalyticsRepo) GetTopCountries(ctx context.Context, linkID uuid.UUID, dr models.DateRange, limit int) ([]models.CountryStats, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			COALESCE(NULLIF(country_code, ''), 'Unknown') AS cc,
			COUNT(*) AS clicks
		FROM clicks
		WHERE link_id = $1 AND clicked_at >= $2 AND clicked_at <= $3 AND is_bot = false
		GROUP BY cc
		ORDER BY clicks DESC
		LIMIT $4
	`, linkID, dr.Start, dr.End, limit)
	if err != nil {
		return nil, fmt.Errorf("pg get countries: %w", err)
	}
	defer rows.Close()

	var total int64
	var stats []models.CountryStats
	for rows.Next() {
		var s models.CountryStats
		if err := rows.Scan(&s.CountryCode, &s.Clicks); err != nil {
			return nil, fmt.Errorf("pg scan country: %w", err)
		}
		s.Country = s.CountryCode
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

func (r *pgAnalyticsRepo) GetDeviceBreakdown(ctx context.Context, linkID uuid.UUID, dr models.DateRange) (*models.DeviceBreakdown, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			COALESCE(NULLIF(device_type, ''), 'desktop') AS dt,
			COUNT(*) AS clicks
		FROM clicks
		WHERE link_id = $1 AND clicked_at >= $2 AND clicked_at <= $3 AND is_bot = false
		GROUP BY dt
	`, linkID, dr.Start, dr.End)
	if err != nil {
		return nil, fmt.Errorf("pg get devices: %w", err)
	}
	defer rows.Close()

	breakdown := &models.DeviceBreakdown{}
	for rows.Next() {
		var dt string
		var clicks int64
		if err := rows.Scan(&dt, &clicks); err != nil {
			return nil, fmt.Errorf("pg scan device: %w", err)
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

func (r *pgAnalyticsRepo) GetBrowserBreakdown(ctx context.Context, linkID uuid.UUID, dr models.DateRange, limit int) ([]models.BrowserStats, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			COALESCE(NULLIF(browser, ''), 'Unknown') AS b,
			COUNT(*) AS clicks
		FROM clicks
		WHERE link_id = $1 AND clicked_at >= $2 AND clicked_at <= $3 AND is_bot = false
		GROUP BY b
		ORDER BY clicks DESC
		LIMIT $4
	`, linkID, dr.Start, dr.End, limit)
	if err != nil {
		return nil, fmt.Errorf("pg get browsers: %w", err)
	}
	defer rows.Close()

	var total int64
	var stats []models.BrowserStats
	for rows.Next() {
		var s models.BrowserStats
		if err := rows.Scan(&s.Browser, &s.Clicks); err != nil {
			return nil, fmt.Errorf("pg scan browser: %w", err)
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

func pgTruncInterval(interval models.TimeSeriesInterval) string {
	switch interval {
	case models.IntervalHour:
		return "hour"
	case models.IntervalDay:
		return "day"
	case models.IntervalWeek:
		return "week"
	case models.IntervalMonth:
		return "month"
	default:
		return "day"
	}
}
