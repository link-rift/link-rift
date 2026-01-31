package models

import (
	"time"

	"github.com/google/uuid"
)

// DateRange represents a time window for analytics queries.
type DateRange struct {
	Start time.Time
	End   time.Time
}

// DateRangeFromPreset creates a DateRange from a named preset.
// Supported presets: "24h", "7d", "30d", "90d".
func DateRangeFromPreset(preset string) DateRange {
	now := time.Now().UTC()
	switch preset {
	case "24h":
		return DateRange{Start: now.Add(-24 * time.Hour), End: now}
	case "7d":
		return DateRange{Start: now.Add(-7 * 24 * time.Hour), End: now}
	case "30d":
		return DateRange{Start: now.Add(-30 * 24 * time.Hour), End: now}
	case "90d":
		return DateRange{Start: now.Add(-90 * 24 * time.Hour), End: now}
	default:
		return DateRange{Start: now.Add(-7 * 24 * time.Hour), End: now}
	}
}

// ClampToRetention clamps the date range start so it doesn't exceed the retention limit.
// retentionDays < 0 means unlimited.
func (dr DateRange) ClampToRetention(retentionDays int64) DateRange {
	if retentionDays < 0 {
		return dr
	}
	earliest := time.Now().UTC().Add(-time.Duration(retentionDays) * 24 * time.Hour)
	if dr.Start.Before(earliest) {
		dr.Start = earliest
	}
	return dr
}

// TimeSeriesInterval controls the aggregation bucket width.
type TimeSeriesInterval string

const (
	IntervalHour  TimeSeriesInterval = "hour"
	IntervalDay   TimeSeriesInterval = "day"
	IntervalWeek  TimeSeriesInterval = "week"
	IntervalMonth TimeSeriesInterval = "month"
)

// LinkAnalytics holds aggregated stats for a single link.
type LinkAnalytics struct {
	TotalClicks  int64 `json:"total_clicks"`
	UniqueClicks int64 `json:"unique_clicks"`
	Clicks24h    int64 `json:"clicks_24h"`
	Clicks7d     int64 `json:"clicks_7d"`
	Clicks30d    int64 `json:"clicks_30d"`
}

// WorkspaceAnalytics holds aggregated stats for a workspace.
type WorkspaceAnalytics struct {
	TotalLinks   int64      `json:"total_links"`
	TotalClicks  int64      `json:"total_clicks"`
	UniqueClicks int64      `json:"unique_clicks"`
	Clicks24h    int64      `json:"clicks_24h"`
	Clicks7d     int64      `json:"clicks_7d"`
	Clicks30d    int64      `json:"clicks_30d"`
	TopLinks     []TopLink  `json:"top_links"`
}

// TopLink is a link with its click count, used in workspace analytics.
type TopLink struct {
	LinkID     uuid.UUID `json:"link_id"`
	ShortCode  string    `json:"short_code"`
	TotalClicks int64    `json:"total_clicks"`
}

// TimeSeriesPoint is a single data point in a time-series chart.
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Clicks    int64     `json:"clicks"`
	Unique    int64     `json:"unique"`
}

// ReferrerStats holds click counts grouped by referrer domain.
type ReferrerStats struct {
	Referrer string  `json:"referrer"`
	Clicks   int64   `json:"clicks"`
	Percent  float64 `json:"percent"`
}

// CountryStats holds click counts grouped by country.
type CountryStats struct {
	CountryCode string  `json:"country_code"`
	Country     string  `json:"country"`
	Clicks      int64   `json:"clicks"`
	Percent     float64 `json:"percent"`
}

// DeviceBreakdown holds click percentages by device type.
type DeviceBreakdown struct {
	Desktop int64   `json:"desktop"`
	Mobile  int64   `json:"mobile"`
	Tablet  int64   `json:"tablet"`
	Other   int64   `json:"other"`
}

// BrowserStats holds click counts grouped by browser.
type BrowserStats struct {
	Browser string  `json:"browser"`
	Clicks  int64   `json:"clicks"`
	Percent float64 `json:"percent"`
}

// AnalyticsExportFormat specifies the export file format.
type AnalyticsExportFormat string

const (
	ExportCSV  AnalyticsExportFormat = "csv"
	ExportJSON AnalyticsExportFormat = "json"
)
