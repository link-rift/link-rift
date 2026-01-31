package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

// AnalyticsService provides analytics data with feature gating and retention clamping.
type AnalyticsService interface {
	GetLinkStats(ctx context.Context, linkID uuid.UUID, dr models.DateRange) (*models.LinkAnalytics, error)
	GetWorkspaceStats(ctx context.Context, workspaceID uuid.UUID, dr models.DateRange) (*models.WorkspaceAnalytics, error)
	GetTimeSeries(ctx context.Context, linkID uuid.UUID, interval models.TimeSeriesInterval, dr models.DateRange) ([]models.TimeSeriesPoint, error)
	GetTopReferrers(ctx context.Context, linkID uuid.UUID, dr models.DateRange, limit int) ([]models.ReferrerStats, error)
	GetTopCountries(ctx context.Context, linkID uuid.UUID, dr models.DateRange, limit int) ([]models.CountryStats, error)
	GetDeviceBreakdown(ctx context.Context, linkID uuid.UUID, dr models.DateRange) (*models.DeviceBreakdown, error)
	GetBrowserBreakdown(ctx context.Context, linkID uuid.UUID, dr models.DateRange, limit int) ([]models.BrowserStats, error)
	ExportLinkData(ctx context.Context, linkID uuid.UUID, dr models.DateRange, format models.AnalyticsExportFormat) ([]byte, string, error)
}

type analyticsService struct {
	repo       repository.AnalyticsRepository
	clickRepo  repository.ClickRepository
	licManager *license.Manager
	logger     *zap.Logger
}

func NewAnalyticsService(
	repo repository.AnalyticsRepository,
	clickRepo repository.ClickRepository,
	licManager *license.Manager,
	logger *zap.Logger,
) AnalyticsService {
	return &analyticsService{
		repo:       repo,
		clickRepo:  clickRepo,
		licManager: licManager,
		logger:     logger,
	}
}

func (s *analyticsService) clampDateRange(dr models.DateRange) models.DateRange {
	retentionDays := s.licManager.GetLimits().AnalyticsRetentionDays
	return dr.ClampToRetention(retentionDays)
}

func (s *analyticsService) GetLinkStats(ctx context.Context, linkID uuid.UUID, dr models.DateRange) (*models.LinkAnalytics, error) {
	dr = s.clampDateRange(dr)
	return s.repo.GetLinkStats(ctx, linkID, dr)
}

func (s *analyticsService) GetWorkspaceStats(ctx context.Context, workspaceID uuid.UUID, dr models.DateRange) (*models.WorkspaceAnalytics, error) {
	dr = s.clampDateRange(dr)
	return s.repo.GetWorkspaceStats(ctx, workspaceID, dr)
}

func (s *analyticsService) GetTimeSeries(ctx context.Context, linkID uuid.UUID, interval models.TimeSeriesInterval, dr models.DateRange) ([]models.TimeSeriesPoint, error) {
	dr = s.clampDateRange(dr)
	return s.repo.GetTimeSeries(ctx, linkID, interval, dr)
}

func (s *analyticsService) GetTopReferrers(ctx context.Context, linkID uuid.UUID, dr models.DateRange, limit int) ([]models.ReferrerStats, error) {
	if !s.licManager.HasFeature(license.FeatureAdvancedAnalytics) {
		return nil, httputil.PaymentRequiredWithDetails(string(license.FeatureAdvancedAnalytics), "pro")
	}
	dr = s.clampDateRange(dr)
	return s.repo.GetTopReferrers(ctx, linkID, dr, limit)
}

func (s *analyticsService) GetTopCountries(ctx context.Context, linkID uuid.UUID, dr models.DateRange, limit int) ([]models.CountryStats, error) {
	if !s.licManager.HasFeature(license.FeatureAdvancedAnalytics) {
		return nil, httputil.PaymentRequiredWithDetails(string(license.FeatureAdvancedAnalytics), "pro")
	}
	dr = s.clampDateRange(dr)
	return s.repo.GetTopCountries(ctx, linkID, dr, limit)
}

func (s *analyticsService) GetDeviceBreakdown(ctx context.Context, linkID uuid.UUID, dr models.DateRange) (*models.DeviceBreakdown, error) {
	if !s.licManager.HasFeature(license.FeatureAdvancedAnalytics) {
		return nil, httputil.PaymentRequiredWithDetails(string(license.FeatureAdvancedAnalytics), "pro")
	}
	dr = s.clampDateRange(dr)
	return s.repo.GetDeviceBreakdown(ctx, linkID, dr)
}

func (s *analyticsService) GetBrowserBreakdown(ctx context.Context, linkID uuid.UUID, dr models.DateRange, limit int) ([]models.BrowserStats, error) {
	if !s.licManager.HasFeature(license.FeatureAdvancedAnalytics) {
		return nil, httputil.PaymentRequiredWithDetails(string(license.FeatureAdvancedAnalytics), "pro")
	}
	dr = s.clampDateRange(dr)
	return s.repo.GetBrowserBreakdown(ctx, linkID, dr, limit)
}

func (s *analyticsService) ExportLinkData(ctx context.Context, linkID uuid.UUID, dr models.DateRange, format models.AnalyticsExportFormat) ([]byte, string, error) {
	if !s.licManager.HasFeature(license.FeatureExportData) {
		return nil, "", httputil.PaymentRequiredWithDetails(string(license.FeatureExportData), "pro")
	}

	dr = s.clampDateRange(dr)

	// Get stats + time series for export
	stats, err := s.repo.GetLinkStats(ctx, linkID, dr)
	if err != nil {
		return nil, "", fmt.Errorf("export get stats: %w", err)
	}

	timeSeries, err := s.repo.GetTimeSeries(ctx, linkID, models.IntervalDay, dr)
	if err != nil {
		return nil, "", fmt.Errorf("export get time series: %w", err)
	}

	switch format {
	case models.ExportJSON:
		exportData := map[string]any{
			"link_id":     linkID.String(),
			"date_range":  map[string]string{"start": dr.Start.Format("2006-01-02"), "end": dr.End.Format("2006-01-02")},
			"stats":       stats,
			"time_series": timeSeries,
		}
		data, err := json.MarshalIndent(exportData, "", "  ")
		if err != nil {
			return nil, "", fmt.Errorf("export marshal json: %w", err)
		}
		return data, "application/json", nil

	case models.ExportCSV:
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		_ = w.Write([]string{"date", "clicks", "unique_clicks"})
		for _, p := range timeSeries {
			_ = w.Write([]string{
				p.Timestamp.Format("2006-01-02"),
				fmt.Sprintf("%d", p.Clicks),
				fmt.Sprintf("%d", p.Unique),
			})
		}
		w.Flush()
		return buf.Bytes(), "text/csv", nil

	default:
		return nil, "", httputil.Validation("format", "unsupported export format, use csv or json")
	}
}
