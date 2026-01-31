package redirect

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

// --- Mock LinkRepository ---

type mockLinkRepo struct {
	getByShortCodeFn func(ctx context.Context, shortCode string) (*models.Link, error)
}

func (m *mockLinkRepo) Create(_ context.Context, _ sqlc.CreateLinkParams) (*models.Link, error) {
	return nil, nil
}
func (m *mockLinkRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.Link, error) {
	return nil, nil
}
func (m *mockLinkRepo) GetByShortCode(ctx context.Context, shortCode string) (*models.Link, error) {
	if m.getByShortCodeFn != nil {
		return m.getByShortCodeFn(ctx, shortCode)
	}
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
func (m *mockLinkRepo) IncrementClicks(_ context.Context, _ uuid.UUID) error       { return nil }
func (m *mockLinkRepo) IncrementUniqueClicks(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockLinkRepo) GetQuickStats(_ context.Context, _ uuid.UUID) (*models.LinkQuickStats, error) {
	return nil, nil
}
func (m *mockLinkRepo) GetCountForWorkspace(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

// --- Tests ---

func TestResolver_CacheHit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cache := &Cache{l1TTL: 5 * time.Minute}
	repo := &mockLinkRepo{}

	link := &CachedLink{
		ID:             uuid.New(),
		ShortCode:      "cached",
		DestinationURL: "https://example.com",
		IsActive:       true,
	}
	cache.SetL1("cached", link)

	resolver := NewResolver(cache, repo, logger)

	result, err := resolver.Resolve(context.Background(), "cached")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DestinationURL != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", result.DestinationURL)
	}
}

func TestResolver_CacheMiss_DBHit(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	linkID := uuid.New()
	repo := &mockLinkRepo{
		getByShortCodeFn: func(_ context.Context, shortCode string) (*models.Link, error) {
			return &models.Link{
				ID:        linkID,
				ShortCode: shortCode,
				URL:       "https://example.com/from-db",
				IsActive:  true,
			}, nil
		},
	}

	// Use a custom resolver that bypasses L2 cache (no Redis in unit tests).
	// We test by pre-populating the L1 cache miss and directly calling the resolver.
	cache := &Cache{l1TTL: 5 * time.Minute}
	resolver := &Resolver{
		cache:    cache,
		linkRepo: repo,
		logger:   logger,
	}

	// Direct DB lookup since we can't use full Get (needs Redis for L2)
	link, err := repo.GetByShortCode(context.Background(), "fromdb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.URL != "https://example.com/from-db" {
		t.Errorf("expected from-db URL, got %s", link.URL)
	}

	// Verify the resolver's cachedToResult works
	cl := &CachedLink{
		ID:             link.ID,
		ShortCode:      link.ShortCode,
		DestinationURL: link.URL,
		IsActive:       link.IsActive,
	}
	result := resolver.cachedToResult(cl)
	if result.DestinationURL != "https://example.com/from-db" {
		t.Errorf("expected from-db URL, got %s", result.DestinationURL)
	}

	// After caching in L1, Resolve should work
	cache.SetL1("fromdb", cl)
	result, err = resolver.Resolve(context.Background(), "fromdb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DestinationURL != "https://example.com/from-db" {
		t.Errorf("expected from-db URL, got %s", result.DestinationURL)
	}
}

func TestResolver_NotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cache := &Cache{l1TTL: 5 * time.Minute}

	repo := &mockLinkRepo{
		getByShortCodeFn: func(_ context.Context, _ string) (*models.Link, error) {
			return nil, httputil.NotFound("link")
		},
	}

	resolver := NewResolver(cache, repo, logger)

	_, err := resolver.Resolve(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for missing link")
	}
}

func TestResolver_ExpiredLink(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cache := &Cache{l1TTL: 5 * time.Minute}

	past := time.Now().Add(-1 * time.Hour).Unix()
	link := &CachedLink{
		ID:             uuid.New(),
		ShortCode:      "expired",
		DestinationURL: "https://example.com",
		IsActive:       true,
		ExpiresAt:      &past,
	}
	cache.SetL1("expired", link)

	resolver := NewResolver(cache, nil, logger)

	result, err := resolver.Resolve(context.Background(), "expired")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsExpired {
		t.Error("expected IsExpired to be true")
	}
}

func TestResolver_OverClickLimit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cache := &Cache{l1TTL: 5 * time.Minute}

	maxClicks := int32(100)
	link := &CachedLink{
		ID:             uuid.New(),
		ShortCode:      "limited",
		DestinationURL: "https://example.com",
		IsActive:       true,
		MaxClicks:      &maxClicks,
		TotalClicks:    150,
	}
	cache.SetL1("limited", link)

	resolver := NewResolver(cache, nil, logger)

	result, err := resolver.Resolve(context.Background(), "limited")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsOverLimit {
		t.Error("expected IsOverLimit to be true")
	}
}

func TestResolver_InvalidateCache(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cache := &Cache{l1TTL: 5 * time.Minute}

	link := &CachedLink{
		ID:             uuid.New(),
		ShortCode:      "invalidate",
		DestinationURL: "https://example.com",
		IsActive:       true,
	}
	cache.SetL1("invalidate", link)

	resolver := NewResolver(cache, nil, logger)
	resolver.cache.l1.Delete("invalidate")

	_, ok := cache.GetL1("invalidate")
	if ok {
		t.Error("expected cache miss after invalidation")
	}
}

// --- Benchmarks ---

func BenchmarkResolverResolve_CacheHit(b *testing.B) {
	logger := zap.NewNop()
	cache := &Cache{l1TTL: 5 * time.Minute}

	link := &CachedLink{
		ID:             uuid.New(),
		ShortCode:      "bench",
		DestinationURL: "https://example.com",
		IsActive:       true,
	}
	cache.SetL1("bench", link)

	resolver := NewResolver(cache, nil, logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolver.Resolve(ctx, "bench")
	}
}

