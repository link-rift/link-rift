package redirect

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const redisKeyPrefix = "link:resolve:"

// CachedLink holds the minimal fields needed for redirect resolution.
type CachedLink struct {
	ID             uuid.UUID `json:"id"`
	ShortCode      string    `json:"short_code"`
	DestinationURL string    `json:"destination_url"`
	IsActive       bool      `json:"is_active"`
	HasPassword    bool      `json:"has_password"`
	PasswordHash   string    `json:"password_hash,omitempty"`
	ExpiresAt      *int64    `json:"expires_at,omitempty"` // unix timestamp
	MaxClicks      *int32    `json:"max_clicks,omitempty"`
	TotalClicks    int64     `json:"total_clicks"`
}

type l1Entry struct {
	link      *CachedLink
	expiresAt time.Time
}

// Cache provides a multi-layer caching strategy for link resolution.
// L1: in-memory sync.Map with TTL entries.
// L2: Redis with configurable TTL.
type Cache struct {
	l1        sync.Map
	l1TTL     time.Duration
	redis     *redis.Client
	redisTTL  time.Duration
	logger    *zap.Logger
}

func NewCache(redisClient *redis.Client, l1TTL, redisTTL time.Duration, logger *zap.Logger) *Cache {
	return &Cache{
		l1TTL:    l1TTL,
		redis:    redisClient,
		redisTTL: redisTTL,
		logger:   logger,
	}
}

// GetL1 checks the local in-memory cache.
func (c *Cache) GetL1(shortCode string) (*CachedLink, bool) {
	val, ok := c.l1.Load(shortCode)
	if !ok {
		return nil, false
	}
	entry := val.(*l1Entry)
	if time.Now().After(entry.expiresAt) {
		c.l1.Delete(shortCode)
		return nil, false
	}
	return entry.link, true
}

// SetL1 stores a link in the local in-memory cache.
func (c *Cache) SetL1(shortCode string, link *CachedLink) {
	c.l1.Store(shortCode, &l1Entry{
		link:      link,
		expiresAt: time.Now().Add(c.l1TTL),
	})
}

// GetL2 checks the Redis cache.
func (c *Cache) GetL2(ctx context.Context, shortCode string) (*CachedLink, bool) {
	data, err := c.redis.Get(ctx, redisKeyPrefix+shortCode).Bytes()
	if err != nil {
		return nil, false
	}

	var link CachedLink
	if err := json.Unmarshal(data, &link); err != nil {
		c.logger.Warn("failed to unmarshal cached link", zap.Error(err), zap.String("short_code", shortCode))
		return nil, false
	}

	return &link, true
}

// SetL2 stores a link in the Redis cache.
func (c *Cache) SetL2(ctx context.Context, shortCode string, link *CachedLink) {
	data, err := json.Marshal(link)
	if err != nil {
		c.logger.Warn("failed to marshal link for cache", zap.Error(err))
		return
	}

	if err := c.redis.Set(ctx, redisKeyPrefix+shortCode, data, c.redisTTL).Err(); err != nil {
		c.logger.Warn("failed to set redis cache", zap.Error(err), zap.String("short_code", shortCode))
	}
}

// Get performs a multi-layer cache lookup: L1 → L2.
// Returns the cached link and which layer it was found in (0 = miss, 1 = L1, 2 = L2).
func (c *Cache) Get(ctx context.Context, shortCode string) (*CachedLink, int) {
	// L1
	if link, ok := c.GetL1(shortCode); ok {
		return link, 1
	}

	// L2
	if link, ok := c.GetL2(ctx, shortCode); ok {
		// Promote to L1
		c.SetL1(shortCode, link)
		return link, 2
	}

	return nil, 0
}

// Set stores a link in both cache layers.
func (c *Cache) Set(ctx context.Context, shortCode string, link *CachedLink) {
	c.SetL1(shortCode, link)
	c.SetL2(ctx, shortCode, link)
}

// Invalidate removes a link from both cache layers.
func (c *Cache) Invalidate(ctx context.Context, shortCode string) {
	c.l1.Delete(shortCode)
	if err := c.redis.Del(ctx, redisKeyPrefix+shortCode).Err(); err != nil {
		c.logger.Warn("failed to invalidate redis cache", zap.Error(err), zap.String("short_code", shortCode))
	}
}
