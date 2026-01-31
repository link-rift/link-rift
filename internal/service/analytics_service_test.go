package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

// mockAnalyticsRepo is a test double for AnalyticsRepository.
type mockAnalyticsRepo struct {
	linkStats       *models.LinkAnalytics
	workspaceStats  *models.WorkspaceAnalytics
	timeSeries      []models.TimeSeriesPoint
	referrers       []models.ReferrerStats
	countries       []models.CountryStats
	deviceBreakdown *models.DeviceBreakdown
	browsers        []models.BrowserStats
	err             error
}

func (m *mockAnalyticsRepo) GetLinkStats(_ context.Context, _ uuid.UUID, _ models.DateRange) (*models.LinkAnalytics, error) {
	return m.linkStats, m.err
}
func (m *mockAnalyticsRepo) GetWorkspaceStats(_ context.Context, _ uuid.UUID, _ models.DateRange) (*models.WorkspaceAnalytics, error) {
	return m.workspaceStats, m.err
}
func (m *mockAnalyticsRepo) GetTimeSeries(_ context.Context, _ uuid.UUID, _ models.TimeSeriesInterval, _ models.DateRange) ([]models.TimeSeriesPoint, error) {
	return m.timeSeries, m.err
}
func (m *mockAnalyticsRepo) GetTopReferrers(_ context.Context, _ uuid.UUID, _ models.DateRange, _ int) ([]models.ReferrerStats, error) {
	return m.referrers, m.err
}
func (m *mockAnalyticsRepo) GetTopCountries(_ context.Context, _ uuid.UUID, _ models.DateRange, _ int) ([]models.CountryStats, error) {
	return m.countries, m.err
}
func (m *mockAnalyticsRepo) GetDeviceBreakdown(_ context.Context, _ uuid.UUID, _ models.DateRange) (*models.DeviceBreakdown, error) {
	return m.deviceBreakdown, m.err
}
func (m *mockAnalyticsRepo) GetBrowserBreakdown(_ context.Context, _ uuid.UUID, _ models.DateRange, _ int) ([]models.BrowserStats, error) {
	return m.browsers, m.err
}

func newTestLicenseManager(tier license.Tier) *license.Manager {
	v, _ := license.NewVerifier()
	m := license.NewManager(v, zap.NewNop())
	// For testing, the community edition defaults are fine for free tier.
	// For higher tiers we'd need a real license key — just test free tier behaviour.
	_ = tier // we rely on the default (free) for gating tests
	return m
}

func TestGetLinkStats(t *testing.T) {
	repo := &mockAnalyticsRepo{
		linkStats: &models.LinkAnalytics{
			TotalClicks:  100,
			UniqueClicks: 80,
			Clicks24h:    10,
			Clicks7d:     50,
			Clicks30d:    100,
		},
	}

	svc := NewAnalyticsService(repo, nil, newTestLicenseManager(license.TierFree), zap.NewNop())

	dr := models.DateRangeFromPreset("7d")
	stats, err := svc.GetLinkStats(context.Background(), uuid.New(), dr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalClicks != 100 {
		t.Errorf("expected 100 total clicks, got %d", stats.TotalClicks)
	}
}

func TestGetTimeSeries(t *testing.T) {
	now := time.Now().UTC()
	repo := &mockAnalyticsRepo{
		timeSeries: []models.TimeSeriesPoint{
			{Timestamp: now.Add(-48 * time.Hour), Clicks: 5, Unique: 3},
			{Timestamp: now.Add(-24 * time.Hour), Clicks: 10, Unique: 7},
		},
	}

	svc := NewAnalyticsService(repo, nil, newTestLicenseManager(license.TierFree), zap.NewNop())

	dr := models.DateRangeFromPreset("7d")
	points, err := svc.GetTimeSeries(context.Background(), uuid.New(), models.IntervalDay, dr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(points) != 2 {
		t.Errorf("expected 2 points, got %d", len(points))
	}
}

func TestAdvancedAnalyticsGated(t *testing.T) {
	repo := &mockAnalyticsRepo{
		referrers: []models.ReferrerStats{{Referrer: "google.com", Clicks: 10, Percent: 100}},
	}

	// Free tier should not have advanced analytics
	svc := NewAnalyticsService(repo, nil, newTestLicenseManager(license.TierFree), zap.NewNop())
	dr := models.DateRangeFromPreset("7d")

	_, err := svc.GetTopReferrers(context.Background(), uuid.New(), dr, 10)
	if err == nil {
		t.Fatal("expected payment required error for free tier")
	}

	appErr, ok := err.(*httputil.AppError)
	if !ok || appErr.Code != "PAYMENT_REQUIRED" {
		t.Errorf("expected PAYMENT_REQUIRED error, got: %v", err)
	}
}

func TestExportDataGated(t *testing.T) {
	repo := &mockAnalyticsRepo{}

	svc := NewAnalyticsService(repo, nil, newTestLicenseManager(license.TierFree), zap.NewNop())
	dr := models.DateRangeFromPreset("7d")

	_, _, err := svc.ExportLinkData(context.Background(), uuid.New(), dr, models.ExportJSON)
	if err == nil {
		t.Fatal("expected payment required error for free tier export")
	}

	appErr, ok := err.(*httputil.AppError)
	if !ok || appErr.Code != "PAYMENT_REQUIRED" {
		t.Errorf("expected PAYMENT_REQUIRED error, got: %v", err)
	}
}

func TestDateRangeClampToRetention(t *testing.T) {
	now := time.Now().UTC()
	dr := models.DateRange{
		Start: now.Add(-365 * 24 * time.Hour), // 1 year ago
		End:   now,
	}

	clamped := dr.ClampToRetention(30)
	earliest := now.Add(-30 * 24 * time.Hour)
	if clamped.Start.Before(earliest.Add(-time.Second)) {
		t.Errorf("expected start >= %v, got %v", earliest, clamped.Start)
	}

	// Unlimited (-1)
	unclamped := dr.ClampToRetention(-1)
	if !unclamped.Start.Equal(dr.Start) {
		t.Error("unlimited retention should not clamp")
	}
}
