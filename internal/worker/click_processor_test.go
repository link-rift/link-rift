package worker

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/redirect"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"go.uber.org/zap"
)

// --- Mock Repositories ---

type mockClickRepo struct {
	insertFn func(ctx context.Context, params sqlc.InsertClickParams) error
	getByFn  func(ctx context.Context, params sqlc.GetClicksByLinkIDParams) ([]*models.Click, error)
}

func (m *mockClickRepo) Insert(ctx context.Context, params sqlc.InsertClickParams) error {
	if m.insertFn != nil {
		return m.insertFn(ctx, params)
	}
	return nil
}

func (m *mockClickRepo) GetByLinkID(ctx context.Context, params sqlc.GetClicksByLinkIDParams) ([]*models.Click, error) {
	if m.getByFn != nil {
		return m.getByFn(ctx, params)
	}
	return nil, nil
}

type mockLinkRepo struct {
	incrementFn func(ctx context.Context, id uuid.UUID) error
}

func (m *mockLinkRepo) Create(_ context.Context, _ sqlc.CreateLinkParams) (*models.Link, error) {
	return nil, nil
}
func (m *mockLinkRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.Link, error) {
	return nil, nil
}
func (m *mockLinkRepo) GetByShortCode(_ context.Context, _ string) (*models.Link, error) {
	return nil, nil
}
func (m *mockLinkRepo) GetByURL(_ context.Context, _ sqlc.GetLinkByURLParams) (*models.Link, error) {
	return nil, nil
}
func (m *mockLinkRepo) List(_ context.Context, _ sqlc.ListLinksForWorkspaceParams) ([]*models.Link, int64, error) {
	return nil, 0, nil
}
func (m *mockLinkRepo) Update(_ context.Context, _ sqlc.UpdateLinkParams) (*models.Link, error) {
	return nil, nil
}
func (m *mockLinkRepo) SoftDelete(_ context.Context, _ uuid.UUID) error   { return nil }
func (m *mockLinkRepo) ShortCodeExists(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (m *mockLinkRepo) IncrementClicks(ctx context.Context, id uuid.UUID) error {
	if m.incrementFn != nil {
		return m.incrementFn(ctx, id)
	}
	return nil
}
func (m *mockLinkRepo) IncrementUniqueClicks(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockLinkRepo) GetQuickStats(_ context.Context, _ uuid.UUID) (*models.LinkQuickStats, error) {
	return nil, nil
}
func (m *mockLinkRepo) GetCountForWorkspace(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

// --- UA Parsing Tests ---

func TestParseBrowser(t *testing.T) {
	tests := []struct {
		ua          string
		wantName    string
		wantVersion string
	}{
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			"Chrome", "91.0.4472.124",
		},
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
			"Firefox", "89.0",
		},
		{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
			"Safari", "14.1.1",
		},
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59",
			"Edge", "91.0.864.59",
		},
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0 Safari/537.36 OPR/77.0",
			"Opera", "77.0",
		},
		{"", "", ""},
		{"some random string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			name, version := parseBrowser(tt.ua)
			if name != tt.wantName {
				t.Errorf("parseBrowser(%q) name = %q, want %q", tt.ua, name, tt.wantName)
			}
			if version != tt.wantVersion {
				t.Errorf("parseBrowser(%q) version = %q, want %q", tt.ua, version, tt.wantVersion)
			}
		})
	}
}

func TestParseOS(t *testing.T) {
	tests := []struct {
		ua          string
		wantName    string
		wantVersion string
	}{
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			"Windows", "10.0",
		},
		{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
			"macOS", "10.15.7",
		},
		{
			"Mozilla/5.0 (X11; Linux x86_64)",
			"Linux", "",
		},
		{
			"Mozilla/5.0 (Linux; Android 11; SM-G998B)",
			"Android", "11",
		},
		{
			"Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X)",
			"iOS", "14.6",
		},
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			name, version := parseOS(tt.ua)
			if name != tt.wantName {
				t.Errorf("parseOS(%q) name = %q, want %q", tt.ua, name, tt.wantName)
			}
			if version != tt.wantVersion {
				t.Errorf("parseOS(%q) version = %q, want %q", tt.ua, version, tt.wantVersion)
			}
		})
	}
}

func TestParseDeviceType(t *testing.T) {
	tests := []struct {
		ua   string
		want string
	}{
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64)", "desktop"},
		{"Mozilla/5.0 (Linux; Android 11) Mobile Safari", "mobile"},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 14_6)", "mobile"},
		{"Mozilla/5.0 (iPad; CPU OS 14_6 like Mac OS X)", "tablet"},
		{"Mozilla/5.0 (Linux; Android 11; SM-T870) Tablet", "tablet"},
		{"", "desktop"},
	}

	for _, tt := range tests {
		t.Run(tt.want+"_"+tt.ua[:min(20, len(tt.ua))], func(t *testing.T) {
			got := parseDeviceType(tt.ua)
			if got != tt.want {
				t.Errorf("parseDeviceType(%q) = %q, want %q", tt.ua, got, tt.want)
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// --- processEvents Tests ---

func TestProcessEvents_HumanClick(t *testing.T) {
	var insertedParams sqlc.InsertClickParams
	var incrementedID uuid.UUID

	clickRepo := &mockClickRepo{
		insertFn: func(_ context.Context, params sqlc.InsertClickParams) error {
			insertedParams = params
			return nil
		},
	}

	linkRepo := &mockLinkRepo{
		incrementFn: func(_ context.Context, id uuid.UUID) error {
			incrementedID = id
			return nil
		},
	}

	logger, _ := zap.NewDevelopment()
	cp := &ClickProcessor{
		clickRepo:   clickRepo,
		linkRepo:    linkRepo,
		botDetector: redirect.NewBotDetector(),
		logger:      logger,
	}

	linkID := uuid.New()
	events := []*models.ClickEvent{
		{
			LinkID:    linkID,
			ShortCode: "human1",
			IP:        "1.2.3.4",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			Referer:   "https://google.com",
			Timestamp: time.Now(),
		},
	}

	cp.processEvents(context.Background(), events)

	// Verify click was inserted with correct data
	if insertedParams.LinkID != linkID {
		t.Errorf("expected link_id %s, got %s", linkID, insertedParams.LinkID)
	}
	if insertedParams.IsBot {
		t.Error("expected IsBot to be false for human browser")
	}
	if insertedParams.Browser.String != "Chrome" {
		t.Errorf("expected browser Chrome, got %s", insertedParams.Browser.String)
	}
	if insertedParams.Os.String != "Windows" {
		t.Errorf("expected OS Windows, got %s", insertedParams.Os.String)
	}
	if insertedParams.DeviceType.String != "desktop" {
		t.Errorf("expected device_type desktop, got %s", insertedParams.DeviceType.String)
	}

	// Verify click counter was incremented for human
	if incrementedID != linkID {
		t.Errorf("expected increment for %s, got %s", linkID, incrementedID)
	}
}

func TestProcessEvents_BotClick(t *testing.T) {
	var inserted bool
	var incremented bool

	clickRepo := &mockClickRepo{
		insertFn: func(_ context.Context, params sqlc.InsertClickParams) error {
			inserted = true
			if !params.IsBot {
				t.Error("expected IsBot to be true for bot UA")
			}
			return nil
		},
	}

	linkRepo := &mockLinkRepo{
		incrementFn: func(_ context.Context, _ uuid.UUID) error {
			incremented = true
			return nil
		},
	}

	logger, _ := zap.NewDevelopment()
	cp := &ClickProcessor{
		clickRepo:   clickRepo,
		linkRepo:    linkRepo,
		botDetector: redirect.NewBotDetector(),
		logger:      logger,
	}

	events := []*models.ClickEvent{
		{
			LinkID:    uuid.New(),
			ShortCode: "bot1",
			IP:        "1.2.3.4",
			UserAgent: "Googlebot/2.1 (+http://www.google.com/bot.html)",
			Timestamp: time.Now(),
		},
	}

	cp.processEvents(context.Background(), events)

	if !inserted {
		t.Error("bot click should still be inserted in DB")
	}
	if incremented {
		t.Error("bot click should NOT increment link counter")
	}
}

func TestProcessEvents_WithGeoLookup(t *testing.T) {
	var params sqlc.InsertClickParams

	clickRepo := &mockClickRepo{
		insertFn: func(_ context.Context, p sqlc.InsertClickParams) error {
			params = p
			return nil
		},
	}

	logger, _ := zap.NewDevelopment()
	cp := &ClickProcessor{
		clickRepo:   clickRepo,
		linkRepo:    &mockLinkRepo{},
		botDetector: redirect.NewBotDetector(),
		logger:      logger,
		// GeoLookup is nil — geo fields should be empty
	}

	events := []*models.ClickEvent{
		{
			LinkID:    uuid.New(),
			ShortCode: "geo1",
			IP:        "1.2.3.4",
			UserAgent: "Mozilla/5.0 Chrome/91.0",
			Timestamp: time.Now(),
		},
	}

	cp.processEvents(context.Background(), events)

	// With nil GeoLookup, country/region/city should be empty
	if params.CountryCode.Valid {
		t.Errorf("expected empty country_code when no GeoLookup, got %s", params.CountryCode.String)
	}
	if params.Region.Valid {
		t.Errorf("expected empty region when no GeoLookup, got %s", params.Region.String)
	}
	if params.City.Valid {
		t.Errorf("expected empty city when no GeoLookup, got %s", params.City.String)
	}
}

func TestProcessEvents_WithClickHouseForwarder(t *testing.T) {
	// When chForwarder is nil, processEvents should not panic
	logger, _ := zap.NewDevelopment()
	cp := &ClickProcessor{
		clickRepo:   &mockClickRepo{},
		linkRepo:    &mockLinkRepo{},
		botDetector: redirect.NewBotDetector(),
		logger:      logger,
		chForwarder: nil, // explicitly nil
	}

	events := []*models.ClickEvent{
		{
			LinkID:    uuid.New(),
			ShortCode: "ch1",
			IP:        "1.2.3.4",
			UserAgent: "Mozilla/5.0 Chrome/91.0",
			Timestamp: time.Now(),
		},
	}

	// Should not panic
	cp.processEvents(context.Background(), events)
}

func TestProcessEvents_InsertError(t *testing.T) {
	var incrementCalled bool

	clickRepo := &mockClickRepo{
		insertFn: func(_ context.Context, _ sqlc.InsertClickParams) error {
			return &testError{"insert failed"}
		},
	}

	linkRepo := &mockLinkRepo{
		incrementFn: func(_ context.Context, _ uuid.UUID) error {
			incrementCalled = true
			return nil
		},
	}

	logger, _ := zap.NewDevelopment()
	cp := &ClickProcessor{
		clickRepo:   clickRepo,
		linkRepo:    linkRepo,
		botDetector: redirect.NewBotDetector(),
		logger:      logger,
	}

	events := []*models.ClickEvent{
		{
			LinkID:    uuid.New(),
			ShortCode: "err1",
			IP:        "1.2.3.4",
			UserAgent: "Mozilla/5.0 Chrome/91.0",
			Timestamp: time.Now(),
		},
	}

	cp.processEvents(context.Background(), events)

	if incrementCalled {
		t.Error("should not increment clicks when insert fails")
	}
}

func TestProcessEvents_BatchMultipleEvents(t *testing.T) {
	insertCount := 0
	incrementCount := 0

	clickRepo := &mockClickRepo{
		insertFn: func(_ context.Context, _ sqlc.InsertClickParams) error {
			insertCount++
			return nil
		},
	}

	linkRepo := &mockLinkRepo{
		incrementFn: func(_ context.Context, _ uuid.UUID) error {
			incrementCount++
			return nil
		},
	}

	logger, _ := zap.NewDevelopment()
	cp := &ClickProcessor{
		clickRepo:   clickRepo,
		linkRepo:    linkRepo,
		botDetector: redirect.NewBotDetector(),
		logger:      logger,
	}

	events := []*models.ClickEvent{
		{
			LinkID: uuid.New(), ShortCode: "b1", IP: "1.1.1.1",
			UserAgent: "Mozilla/5.0 Chrome/91.0", Timestamp: time.Now(),
		},
		{
			LinkID: uuid.New(), ShortCode: "b2", IP: "2.2.2.2",
			UserAgent: "Googlebot/2.1", Timestamp: time.Now(),
		},
		{
			LinkID: uuid.New(), ShortCode: "b3", IP: "3.3.3.3",
			UserAgent: "Mozilla/5.0 Firefox/89.0", Timestamp: time.Now(),
		},
	}

	cp.processEvents(context.Background(), events)

	if insertCount != 3 {
		t.Errorf("expected 3 inserts, got %d", insertCount)
	}
	// Only 2 human events should increment (1 bot should not)
	if incrementCount != 2 {
		t.Errorf("expected 2 increments (skipping bot), got %d", incrementCount)
	}
}

// --- Helper ---

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }
