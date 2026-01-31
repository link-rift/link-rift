package redirect

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository"
	"go.uber.org/zap"
)

// ResolveResult contains all information needed to handle a redirect.
type ResolveResult struct {
	LinkID         uuid.UUID
	WorkspaceID    uuid.UUID
	ShortCode      string
	DestinationURL string
	IsActive       bool
	HasPassword    bool
	PasswordHash   string
	IsExpired      bool
	IsOverLimit    bool
}

// Resolver resolves short codes to their destination URLs using multi-layer caching.
type Resolver struct {
	cache    *Cache
	linkRepo repository.LinkRepository
	logger   *zap.Logger
}

func NewResolver(cache *Cache, linkRepo repository.LinkRepository, logger *zap.Logger) *Resolver {
	return &Resolver{
		cache:    cache,
		linkRepo: linkRepo,
		logger:   logger,
	}
}

// Resolve looks up a short code through the cache layers and returns the resolve result.
func (r *Resolver) Resolve(ctx context.Context, shortCode string) (*ResolveResult, error) {
	// Try cache first (L1 → L2)
	cached, layer := r.cache.Get(ctx, shortCode)
	if cached != nil {
		r.logger.Debug("cache hit",
			zap.String("short_code", shortCode),
			zap.Int("layer", layer),
		)
		return r.cachedToResult(cached), nil
	}

	// Cache miss — go to database
	link, err := r.linkRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	// Build cached entry
	cl := &CachedLink{
		ID:             link.ID,
		WorkspaceID:    link.WorkspaceID,
		ShortCode:      link.ShortCode,
		DestinationURL: link.URL,
		IsActive:       link.IsActive,
		HasPassword:    link.HasPassword,
		TotalClicks:    link.TotalClicks,
	}
	if link.PasswordHash != nil {
		cl.PasswordHash = *link.PasswordHash
	}
	if link.ExpiresAt != nil {
		ts := link.ExpiresAt.Unix()
		cl.ExpiresAt = &ts
	}
	if link.MaxClicks != nil {
		cl.MaxClicks = link.MaxClicks
	}

	// Populate caches
	r.cache.Set(ctx, shortCode, cl)

	return r.cachedToResult(cl), nil
}

func (r *Resolver) cachedToResult(cl *CachedLink) *ResolveResult {
	result := &ResolveResult{
		LinkID:         cl.ID,
		WorkspaceID:    cl.WorkspaceID,
		ShortCode:      cl.ShortCode,
		DestinationURL: cl.DestinationURL,
		IsActive:       cl.IsActive,
		HasPassword:    cl.HasPassword,
		PasswordHash:   cl.PasswordHash,
	}

	// Check expiration
	if cl.ExpiresAt != nil {
		result.IsExpired = time.Now().Unix() > *cl.ExpiresAt
	}

	// Check click limit
	if cl.MaxClicks != nil {
		result.IsOverLimit = cl.TotalClicks >= int64(*cl.MaxClicks)
	}

	return result
}

// InvalidateCache removes the short code from all cache layers.
func (r *Resolver) InvalidateCache(ctx context.Context, shortCode string) {
	r.cache.Invalidate(ctx, shortCode)
}
