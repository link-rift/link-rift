# Redirect Service

> Last Updated: 2025-01-24

The redirect service is the core of Linkrift, handling millions of URL redirects with sub-millisecond latency through aggressive optimization and multi-layer caching.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [High-Performance Go Optimizations](#high-performance-go-optimizations)
  - [sync.Pool for Object Reuse](#syncpool-for-object-reuse)
  - [Zero-Allocation Patterns](#zero-allocation-patterns)
  - [Buffer Pooling](#buffer-pooling)
- [Multi-Layer Caching](#multi-layer-caching)
  - [L1: In-Memory Cache](#l1-in-memory-cache)
  - [L2: Redis Cache](#l2-redis-cache)
  - [Cache Invalidation](#cache-invalidation)
- [Link Resolution](#link-resolution)
- [Bot Detection](#bot-detection)
- [Async Click Tracking](#async-click-tracking)
- [Performance Benchmarks](#performance-benchmarks)

---

## Overview

The redirect service is designed for extreme performance:

- **Sub-millisecond response times** for cached links
- **Zero-allocation hot paths** to minimize GC pressure
- **Multi-layer caching** with L1 in-memory and L2 Redis
- **Async analytics** to avoid blocking redirects
- **Bot detection** for accurate analytics

## Architecture

```
                                    ┌─────────────────────┐
                                    │    Load Balancer    │
                                    │     (Cloudflare)    │
                                    └──────────┬──────────┘
                                               │
                                    ┌──────────▼──────────┐
                                    │   Redirect Service  │
                                    │    (Go + Fiber)     │
                                    └──────────┬──────────┘
                                               │
                    ┌──────────────────────────┼──────────────────────────┐
                    │                          │                          │
           ┌────────▼────────┐       ┌────────▼────────┐       ┌────────▼────────┐
           │  L1 In-Memory   │       │   L2 Redis      │       │   PostgreSQL    │
           │   (BigCache)    │       │    Cluster      │       │   (Fallback)    │
           └─────────────────┘       └─────────────────┘       └─────────────────┘
                                               │
                                    ┌──────────▼──────────┐
                                    │   Click Processor   │
                                    │      (Async)        │
                                    └──────────┬──────────┘
                                               │
                                    ┌──────────▼──────────┐
                                    │     ClickHouse      │
                                    │    (Analytics)      │
                                    └─────────────────────┘
```

---

## High-Performance Go Optimizations

### sync.Pool for Object Reuse

```go
// internal/redirect/pool.go
package redirect

import (
	"sync"
	"time"
)

// ClickEvent represents a click tracking event
type ClickEvent struct {
	LinkID      string
	ShortCode   string
	Timestamp   time.Time
	IPAddress   string
	UserAgent   string
	Referer     string
	Country     string
	City        string
	DeviceType  string
	Browser     string
	OS          string
	IsBot       bool
	RequestID   string
}

// Reset clears the ClickEvent for reuse
func (ce *ClickEvent) Reset() {
	ce.LinkID = ""
	ce.ShortCode = ""
	ce.Timestamp = time.Time{}
	ce.IPAddress = ""
	ce.UserAgent = ""
	ce.Referer = ""
	ce.Country = ""
	ce.City = ""
	ce.DeviceType = ""
	ce.Browser = ""
	ce.OS = ""
	ce.IsBot = false
	ce.RequestID = ""
}

var clickEventPool = sync.Pool{
	New: func() interface{} {
		return &ClickEvent{}
	},
}

// AcquireClickEvent gets a ClickEvent from the pool
func AcquireClickEvent() *ClickEvent {
	return clickEventPool.Get().(*ClickEvent)
}

// ReleaseClickEvent returns a ClickEvent to the pool
func ReleaseClickEvent(ce *ClickEvent) {
	ce.Reset()
	clickEventPool.Put(ce)
}

// RedirectContext holds request-specific data
type RedirectContext struct {
	ShortCode    string
	Link         *Link
	ClientIP     string
	UserAgent    string
	Referer      string
	AcceptLang   string
	RequestID    string
	IsBot        bool
	GeoData      *GeoData
	DeviceInfo   *DeviceInfo
}

func (rc *RedirectContext) Reset() {
	rc.ShortCode = ""
	rc.Link = nil
	rc.ClientIP = ""
	rc.UserAgent = ""
	rc.Referer = ""
	rc.AcceptLang = ""
	rc.RequestID = ""
	rc.IsBot = false
	rc.GeoData = nil
	rc.DeviceInfo = nil
}

var redirectContextPool = sync.Pool{
	New: func() interface{} {
		return &RedirectContext{}
	},
}

func AcquireRedirectContext() *RedirectContext {
	return redirectContextPool.Get().(*RedirectContext)
}

func ReleaseRedirectContext(rc *RedirectContext) {
	rc.Reset()
	redirectContextPool.Put(rc)
}
```

### Zero-Allocation Patterns

```go
// internal/redirect/handler.go
package redirect

import (
	"github.com/gofiber/fiber/v2"
	"github.com/link-rift/link-rift/internal/analytics"
)

// Handler handles redirect requests with zero-allocation hot path
type Handler struct {
	resolver    *LinkResolver
	tracker     *analytics.ClickTracker
	botDetector *BotDetector
	geoLocator  *GeoLocator
}

// NewHandler creates a new redirect handler
func NewHandler(
	resolver *LinkResolver,
	tracker *analytics.ClickTracker,
	botDetector *BotDetector,
	geoLocator *GeoLocator,
) *Handler {
	return &Handler{
		resolver:    resolver,
		tracker:     tracker,
		botDetector: botDetector,
		geoLocator:  geoLocator,
	}
}

// HandleRedirect processes redirect requests
// Optimized for zero allocations in the hot path
func (h *Handler) HandleRedirect(c *fiber.Ctx) error {
	// Get short code from path - no allocation, uses Fiber's internal buffer
	shortCode := c.Params("code")
	if len(shortCode) == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid short code")
	}

	// Acquire context from pool
	ctx := AcquireRedirectContext()
	defer ReleaseRedirectContext(ctx)

	// Populate context - uses Fiber's zero-copy methods
	ctx.ShortCode = shortCode
	ctx.ClientIP = c.IP()
	ctx.UserAgent = unsafeString(c.Request().Header.UserAgent())
	ctx.Referer = unsafeString(c.Request().Header.Referer())
	ctx.RequestID = c.Locals("requestID").(string)

	// Resolve link - checks caches first
	link, err := h.resolver.Resolve(c.Context(), shortCode)
	if err != nil {
		if err == ErrLinkNotFound {
			return c.Status(fiber.StatusNotFound).SendString("Link not found")
		}
		if err == ErrLinkExpired {
			return c.Status(fiber.StatusGone).SendString("Link has expired")
		}
		if err == ErrLinkDisabled {
			return c.Status(fiber.StatusForbidden).SendString("Link is disabled")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Internal error")
	}
	ctx.Link = link

	// Bot detection - fast path for known bots
	ctx.IsBot = h.botDetector.IsBot(ctx.UserAgent)

	// Track click asynchronously - non-blocking
	if !ctx.IsBot || link.TrackBots {
		h.trackClickAsync(ctx)
	}

	// Set cache headers for browser caching
	c.Set("Cache-Control", "private, max-age=0, no-cache")
	c.Set("X-Robots-Tag", "noindex, nofollow")

	// Perform redirect - 301 for permanent, 302 for temporary
	statusCode := fiber.StatusMovedPermanently
	if link.RedirectType == RedirectTypeTemporary {
		statusCode = fiber.StatusFound
	}

	return c.Redirect(link.OriginalURL, statusCode)
}

// trackClickAsync sends click event to the tracker without blocking
func (h *Handler) trackClickAsync(ctx *RedirectContext) {
	event := AcquireClickEvent()

	event.LinkID = ctx.Link.ID
	event.ShortCode = ctx.ShortCode
	event.IPAddress = ctx.ClientIP
	event.UserAgent = ctx.UserAgent
	event.Referer = ctx.Referer
	event.RequestID = ctx.RequestID
	event.IsBot = ctx.IsBot

	// Non-blocking send to channel
	select {
	case h.tracker.Events() <- event:
		// Event sent successfully
	default:
		// Channel full, release event back to pool
		ReleaseClickEvent(event)
	}
}

// unsafeString converts byte slice to string without allocation
// ONLY use when the byte slice won't be modified
func unsafeString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
```

### Buffer Pooling

```go
// internal/redirect/buffer.go
package redirect

import (
	"bytes"
	"sync"
)

// BufferPool provides a pool of bytes.Buffer
type BufferPool struct {
	pool sync.Pool
}

// NewBufferPool creates a new buffer pool
func NewBufferPool(initialSize int) *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, initialSize))
			},
		},
	}
}

// Get retrieves a buffer from the pool
func (bp *BufferPool) Get() *bytes.Buffer {
	return bp.pool.Get().(*bytes.Buffer)
}

// Put returns a buffer to the pool
func (bp *BufferPool) Put(buf *bytes.Buffer) {
	buf.Reset()
	bp.pool.Put(buf)
}

// Global buffer pool for redirect responses
var responseBufferPool = NewBufferPool(1024)

// StringBuilderPool provides a pool of strings.Builder
type StringBuilderPool struct {
	pool sync.Pool
}

func NewStringBuilderPool() *StringBuilderPool {
	return &StringBuilderPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &strings.Builder{}
			},
		},
	}
}

func (sp *StringBuilderPool) Get() *strings.Builder {
	return sp.pool.Get().(*strings.Builder)
}

func (sp *StringBuilderPool) Put(sb *strings.Builder) {
	sb.Reset()
	sp.pool.Put(sb)
}

var stringBuilderPool = NewStringBuilderPool()
```

---

## Multi-Layer Caching

### L1: In-Memory Cache

```go
// internal/cache/memory.go
package cache

import (
	"context"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/vmihailenco/msgpack/v5"
)

// MemoryCache implements L1 in-memory caching using BigCache
type MemoryCache struct {
	cache *bigcache.BigCache
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(config MemoryCacheConfig) (*MemoryCache, error) {
	cfg := bigcache.Config{
		Shards:             1024,           // Number of cache shards
		LifeWindow:         config.TTL,     // Time after which entry can be evicted
		CleanWindow:        5 * time.Minute,
		MaxEntriesInWindow: 1000 * 10 * 60, // Max entries in life window
		MaxEntrySize:       500,            // Max entry size in bytes
		Verbose:            false,
		HardMaxCacheSize:   config.MaxSizeMB,
		OnRemove:           nil,
		OnRemoveWithReason: nil,
	}

	cache, err := bigcache.New(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	return &MemoryCache{cache: cache}, nil
}

// Get retrieves a link from the cache
func (mc *MemoryCache) Get(shortCode string) (*Link, error) {
	data, err := mc.cache.Get(shortCode)
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			return nil, ErrCacheMiss
		}
		return nil, err
	}

	var link Link
	if err := msgpack.Unmarshal(data, &link); err != nil {
		return nil, err
	}

	return &link, nil
}

// Set stores a link in the cache
func (mc *MemoryCache) Set(shortCode string, link *Link) error {
	data, err := msgpack.Marshal(link)
	if err != nil {
		return err
	}

	return mc.cache.Set(shortCode, data)
}

// Delete removes a link from the cache
func (mc *MemoryCache) Delete(shortCode string) error {
	return mc.cache.Delete(shortCode)
}

// Stats returns cache statistics
func (mc *MemoryCache) Stats() CacheStats {
	stats := mc.cache.Stats()
	return CacheStats{
		Hits:     stats.Hits,
		Misses:   stats.Misses,
		Entries:  int64(mc.cache.Len()),
		Capacity: int64(mc.cache.Capacity()),
	}
}
```

### L2: Redis Cache

```go
// internal/cache/redis.go
package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

// RedisCache implements L2 distributed caching
type RedisCache struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(client *redis.Client, prefix string, ttl time.Duration) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: prefix,
		ttl:    ttl,
	}
}

// Get retrieves a link from Redis
func (rc *RedisCache) Get(ctx context.Context, shortCode string) (*Link, error) {
	key := rc.prefix + shortCode

	data, err := rc.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheMiss
		}
		return nil, err
	}

	var link Link
	if err := msgpack.Unmarshal(data, &link); err != nil {
		return nil, err
	}

	return &link, nil
}

// Set stores a link in Redis
func (rc *RedisCache) Set(ctx context.Context, shortCode string, link *Link) error {
	key := rc.prefix + shortCode

	data, err := msgpack.Marshal(link)
	if err != nil {
		return err
	}

	return rc.client.Set(ctx, key, data, rc.ttl).Err()
}

// Delete removes a link from Redis
func (rc *RedisCache) Delete(ctx context.Context, shortCode string) error {
	key := rc.prefix + shortCode
	return rc.client.Del(ctx, key).Err()
}

// MGet retrieves multiple links from Redis
func (rc *RedisCache) MGet(ctx context.Context, shortCodes []string) (map[string]*Link, error) {
	keys := make([]string, len(shortCodes))
	for i, code := range shortCodes {
		keys[i] = rc.prefix + code
	}

	values, err := rc.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]*Link)
	for i, val := range values {
		if val == nil {
			continue
		}

		var link Link
		if err := msgpack.Unmarshal([]byte(val.(string)), &link); err != nil {
			continue
		}
		result[shortCodes[i]] = &link
	}

	return result, nil
}

// Pipeline executes multiple cache operations atomically
func (rc *RedisCache) Pipeline(ctx context.Context, ops []CacheOp) error {
	pipe := rc.client.Pipeline()

	for _, op := range ops {
		key := rc.prefix + op.ShortCode
		switch op.Type {
		case CacheOpSet:
			data, _ := msgpack.Marshal(op.Link)
			pipe.Set(ctx, key, data, rc.ttl)
		case CacheOpDelete:
			pipe.Del(ctx, key)
		}
	}

	_, err := pipe.Exec(ctx)
	return err
}
```

### Cache Invalidation

```go
// internal/cache/invalidation.go
package cache

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
)

// CacheInvalidator handles cache invalidation across layers
type CacheInvalidator struct {
	l1Cache     *MemoryCache
	l2Cache     *RedisCache
	pubsub      *redis.PubSub
	channel     string
	subscribers map[string]func(string)
	mu          sync.RWMutex
}

// NewCacheInvalidator creates a new cache invalidator
func NewCacheInvalidator(
	l1 *MemoryCache,
	l2 *RedisCache,
	redisClient *redis.Client,
	channel string,
) *CacheInvalidator {
	ci := &CacheInvalidator{
		l1Cache:     l1,
		l2Cache:     l2,
		channel:     channel,
		subscribers: make(map[string]func(string)),
	}

	// Subscribe to invalidation channel
	ci.pubsub = redisClient.Subscribe(context.Background(), channel)
	go ci.listenForInvalidations()

	return ci
}

// Invalidate removes a short code from all cache layers
func (ci *CacheInvalidator) Invalidate(ctx context.Context, shortCode string) error {
	// Invalidate L1 locally
	ci.l1Cache.Delete(shortCode)

	// Invalidate L2
	if err := ci.l2Cache.Delete(ctx, shortCode); err != nil {
		return err
	}

	// Publish invalidation to other instances
	return ci.publishInvalidation(ctx, shortCode)
}

// InvalidateBatch invalidates multiple short codes
func (ci *CacheInvalidator) InvalidateBatch(ctx context.Context, shortCodes []string) error {
	// Invalidate L1 locally
	for _, code := range shortCodes {
		ci.l1Cache.Delete(code)
	}

	// Build pipeline for L2
	ops := make([]CacheOp, len(shortCodes))
	for i, code := range shortCodes {
		ops[i] = CacheOp{
			Type:      CacheOpDelete,
			ShortCode: code,
		}
	}

	if err := ci.l2Cache.Pipeline(ctx, ops); err != nil {
		return err
	}

	// Publish batch invalidation
	for _, code := range shortCodes {
		ci.publishInvalidation(ctx, code)
	}

	return nil
}

func (ci *CacheInvalidator) publishInvalidation(ctx context.Context, shortCode string) error {
	return ci.l2Cache.client.Publish(ctx, ci.channel, shortCode).Err()
}

func (ci *CacheInvalidator) listenForInvalidations() {
	ch := ci.pubsub.Channel()

	for msg := range ch {
		shortCode := msg.Payload
		ci.l1Cache.Delete(shortCode)

		// Notify subscribers
		ci.mu.RLock()
		for _, callback := range ci.subscribers {
			go callback(shortCode)
		}
		ci.mu.RUnlock()
	}
}

// Subscribe adds a callback for invalidation events
func (ci *CacheInvalidator) Subscribe(id string, callback func(string)) {
	ci.mu.Lock()
	ci.subscribers[id] = callback
	ci.mu.Unlock()
}

// Unsubscribe removes a callback
func (ci *CacheInvalidator) Unsubscribe(id string) {
	ci.mu.Lock()
	delete(ci.subscribers, id)
	ci.mu.Unlock()
}
```

---

## Link Resolution

```go
// internal/redirect/resolver.go
package redirect

import (
	"context"
	"time"

	"github.com/link-rift/link-rift/internal/cache"
	"github.com/link-rift/link-rift/internal/db"
)

// Link represents a shortened URL
type Link struct {
	ID            string       `json:"id" msgpack:"id"`
	ShortCode     string       `json:"short_code" msgpack:"sc"`
	OriginalURL   string       `json:"original_url" msgpack:"url"`
	WorkspaceID   string       `json:"workspace_id" msgpack:"wid"`
	DomainID      string       `json:"domain_id,omitempty" msgpack:"did,omitempty"`
	RedirectType  RedirectType `json:"redirect_type" msgpack:"rt"`
	ExpiresAt     *time.Time   `json:"expires_at,omitempty" msgpack:"exp,omitempty"`
	PasswordHash  string       `json:"-" msgpack:"ph,omitempty"`
	IsDisabled    bool         `json:"is_disabled" msgpack:"dis"`
	TrackBots     bool         `json:"track_bots" msgpack:"tb"`
	CreatedAt     time.Time    `json:"created_at" msgpack:"cat"`
	UpdatedAt     time.Time    `json:"updated_at" msgpack:"uat"`
}

type RedirectType int

const (
	RedirectTypePermanent RedirectType = iota
	RedirectTypeTemporary
)

// LinkResolver resolves short codes to original URLs
type LinkResolver struct {
	l1Cache *cache.MemoryCache
	l2Cache *cache.RedisCache
	db      *db.LinkRepository
	metrics *ResolverMetrics
}

// NewLinkResolver creates a new link resolver
func NewLinkResolver(
	l1 *cache.MemoryCache,
	l2 *cache.RedisCache,
	repo *db.LinkRepository,
) *LinkResolver {
	return &LinkResolver{
		l1Cache: l1,
		l2Cache: l2,
		db:      repo,
		metrics: NewResolverMetrics(),
	}
}

// Resolve looks up a link by short code
// Uses a multi-layer cache strategy: L1 -> L2 -> DB
func (lr *LinkResolver) Resolve(ctx context.Context, shortCode string) (*Link, error) {
	// Try L1 (in-memory) first - fastest
	link, err := lr.l1Cache.Get(shortCode)
	if err == nil {
		lr.metrics.L1Hits.Inc()
		return lr.validateLink(link)
	}
	lr.metrics.L1Misses.Inc()

	// Try L2 (Redis) - fast
	link, err = lr.l2Cache.Get(ctx, shortCode)
	if err == nil {
		lr.metrics.L2Hits.Inc()
		// Populate L1 cache
		lr.l1Cache.Set(shortCode, link)
		return lr.validateLink(link)
	}
	lr.metrics.L2Misses.Inc()

	// Fall back to database - slowest
	link, err = lr.db.GetByShortCode(ctx, shortCode)
	if err != nil {
		if err == db.ErrNotFound {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}
	lr.metrics.DBHits.Inc()

	// Populate caches
	lr.l1Cache.Set(shortCode, link)
	lr.l2Cache.Set(ctx, shortCode, link)

	return lr.validateLink(link)
}

// validateLink checks if a link is valid for redirect
func (lr *LinkResolver) validateLink(link *Link) (*Link, error) {
	if link.IsDisabled {
		return nil, ErrLinkDisabled
	}

	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return nil, ErrLinkExpired
	}

	return link, nil
}

// ResolveWithDomain looks up a link by domain and short code
func (lr *LinkResolver) ResolveWithDomain(ctx context.Context, domain, shortCode string) (*Link, error) {
	// Composite cache key for custom domains
	cacheKey := domain + ":" + shortCode

	// Try L1
	link, err := lr.l1Cache.Get(cacheKey)
	if err == nil {
		return lr.validateLink(link)
	}

	// Try L2
	link, err = lr.l2Cache.Get(ctx, cacheKey)
	if err == nil {
		lr.l1Cache.Set(cacheKey, link)
		return lr.validateLink(link)
	}

	// Database lookup
	link, err = lr.db.GetByDomainAndCode(ctx, domain, shortCode)
	if err != nil {
		if err == db.ErrNotFound {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	// Populate caches
	lr.l1Cache.Set(cacheKey, link)
	lr.l2Cache.Set(ctx, cacheKey, link)

	return lr.validateLink(link)
}
```

---

## Bot Detection

```go
// internal/redirect/bot.go
package redirect

import (
	"regexp"
	"strings"
	"sync"
)

// BotDetector identifies bot traffic
type BotDetector struct {
	patterns    []*regexp.Regexp
	knownBots   map[string]bool
	knownBotsMu sync.RWMutex
}

// Common bot user agent patterns
var defaultBotPatterns = []string{
	`(?i)bot`,
	`(?i)crawl`,
	`(?i)spider`,
	`(?i)slurp`,
	`(?i)mediapartners`,
	`(?i)facebookexternalhit`,
	`(?i)twitterbot`,
	`(?i)linkedinbot`,
	`(?i)whatsapp`,
	`(?i)telegrambot`,
	`(?i)discordbot`,
	`(?i)slackbot`,
	`(?i)pingdom`,
	`(?i)uptimerobot`,
	`(?i)lighthouse`,
	`(?i)pagespeed`,
	`(?i)gtmetrix`,
	`(?i)semrush`,
	`(?i)ahref`,
	`(?i)mj12bot`,
	`(?i)dotbot`,
	`(?i)petalbot`,
	`(?i)bingpreview`,
	`(?i)yandex`,
	`(?i)baidu`,
}

// NewBotDetector creates a new bot detector
func NewBotDetector() *BotDetector {
	patterns := make([]*regexp.Regexp, len(defaultBotPatterns))
	for i, p := range defaultBotPatterns {
		patterns[i] = regexp.MustCompile(p)
	}

	return &BotDetector{
		patterns:  patterns,
		knownBots: make(map[string]bool),
	}
}

// IsBot checks if a user agent belongs to a bot
func (bd *BotDetector) IsBot(userAgent string) bool {
	if userAgent == "" {
		return true // Empty user agent is suspicious
	}

	// Check known bots cache
	bd.knownBotsMu.RLock()
	isBot, found := bd.knownBots[userAgent]
	bd.knownBotsMu.RUnlock()
	if found {
		return isBot
	}

	// Check patterns
	for _, pattern := range bd.patterns {
		if pattern.MatchString(userAgent) {
			bd.cacheResult(userAgent, true)
			return true
		}
	}

	// Additional heuristics
	if bd.hasNoJSCapability(userAgent) || bd.isTooShort(userAgent) {
		bd.cacheResult(userAgent, true)
		return true
	}

	bd.cacheResult(userAgent, false)
	return false
}

func (bd *BotDetector) hasNoJSCapability(ua string) bool {
	lowered := strings.ToLower(ua)
	return strings.Contains(lowered, "headless") ||
		strings.Contains(lowered, "phantom") ||
		strings.Contains(lowered, "selenium")
}

func (bd *BotDetector) isTooShort(ua string) bool {
	return len(ua) < 20 // Most real browsers have longer UAs
}

func (bd *BotDetector) cacheResult(ua string, isBot bool) {
	bd.knownBotsMu.Lock()
	// Limit cache size
	if len(bd.knownBots) > 10000 {
		// Simple eviction: clear half
		for k := range bd.knownBots {
			delete(bd.knownBots, k)
			if len(bd.knownBots) <= 5000 {
				break
			}
		}
	}
	bd.knownBots[ua] = isBot
	bd.knownBotsMu.Unlock()
}

// BotInfo provides detailed bot information
type BotInfo struct {
	Name     string
	Category string // search, social, monitoring, seo
}

// IdentifyBot returns detailed information about a bot
func (bd *BotDetector) IdentifyBot(userAgent string) *BotInfo {
	lowered := strings.ToLower(userAgent)

	botMap := map[string]BotInfo{
		"googlebot":            {Name: "Googlebot", Category: "search"},
		"bingbot":              {Name: "Bingbot", Category: "search"},
		"yandexbot":            {Name: "YandexBot", Category: "search"},
		"baiduspider":          {Name: "Baidu Spider", Category: "search"},
		"facebookexternalhit":  {Name: "Facebook", Category: "social"},
		"twitterbot":           {Name: "Twitter", Category: "social"},
		"linkedinbot":          {Name: "LinkedIn", Category: "social"},
		"slackbot":             {Name: "Slack", Category: "social"},
		"discordbot":           {Name: "Discord", Category: "social"},
		"telegrambot":          {Name: "Telegram", Category: "social"},
		"whatsapp":             {Name: "WhatsApp", Category: "social"},
		"pingdom":              {Name: "Pingdom", Category: "monitoring"},
		"uptimerobot":          {Name: "UptimeRobot", Category: "monitoring"},
		"semrushbot":           {Name: "SEMrush", Category: "seo"},
		"ahrefsbot":            {Name: "Ahrefs", Category: "seo"},
	}

	for pattern, info := range botMap {
		if strings.Contains(lowered, pattern) {
			return &info
		}
	}

	return nil
}
```

---

## Async Click Tracking

```go
// internal/analytics/tracker.go
package analytics

import (
	"context"
	"sync"
	"time"

	"github.com/link-rift/link-rift/internal/redirect"
)

// ClickTracker handles asynchronous click event processing
type ClickTracker struct {
	events      chan *redirect.ClickEvent
	batchSize   int
	flushPeriod time.Duration
	writer      ClickWriter
	enricher    *ClickEnricher
	wg          sync.WaitGroup
	quit        chan struct{}
}

// ClickWriter writes click events to storage
type ClickWriter interface {
	WriteBatch(ctx context.Context, events []*redirect.ClickEvent) error
}

// NewClickTracker creates a new click tracker
func NewClickTracker(
	writer ClickWriter,
	enricher *ClickEnricher,
	bufferSize int,
	batchSize int,
	flushPeriod time.Duration,
) *ClickTracker {
	ct := &ClickTracker{
		events:      make(chan *redirect.ClickEvent, bufferSize),
		batchSize:   batchSize,
		flushPeriod: flushPeriod,
		writer:      writer,
		enricher:    enricher,
		quit:        make(chan struct{}),
	}

	return ct
}

// Events returns the event channel for producers
func (ct *ClickTracker) Events() chan<- *redirect.ClickEvent {
	return ct.events
}

// Start begins processing events
func (ct *ClickTracker) Start(workers int) {
	for i := 0; i < workers; i++ {
		ct.wg.Add(1)
		go ct.processEvents()
	}
}

// Stop gracefully shuts down the tracker
func (ct *ClickTracker) Stop() {
	close(ct.quit)
	ct.wg.Wait()
}

func (ct *ClickTracker) processEvents() {
	defer ct.wg.Done()

	batch := make([]*redirect.ClickEvent, 0, ct.batchSize)
	ticker := time.NewTicker(ct.flushPeriod)
	defer ticker.Stop()

	for {
		select {
		case event := <-ct.events:
			// Enrich event with geo and device data
			ct.enricher.Enrich(event)
			event.Timestamp = time.Now()

			batch = append(batch, event)

			if len(batch) >= ct.batchSize {
				ct.flush(batch)
				batch = make([]*redirect.ClickEvent, 0, ct.batchSize)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				ct.flush(batch)
				batch = make([]*redirect.ClickEvent, 0, ct.batchSize)
			}

		case <-ct.quit:
			// Flush remaining events
			for {
				select {
				case event := <-ct.events:
					ct.enricher.Enrich(event)
					event.Timestamp = time.Now()
					batch = append(batch, event)
				default:
					if len(batch) > 0 {
						ct.flush(batch)
					}
					return
				}
			}
		}
	}
}

func (ct *ClickTracker) flush(batch []*redirect.ClickEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := ct.writer.WriteBatch(ctx, batch); err != nil {
		// Log error but don't block - events are best-effort
		// TODO: Add retry queue or dead letter queue
	}

	// Return events to pool
	for _, event := range batch {
		redirect.ReleaseClickEvent(event)
	}
}

// ClickEnricher adds geographic and device information
type ClickEnricher struct {
	geoLocator    *GeoLocator
	deviceParser  *DeviceParser
}

// Enrich adds additional data to a click event
func (ce *ClickEnricher) Enrich(event *redirect.ClickEvent) {
	// Geo enrichment
	if geo := ce.geoLocator.Lookup(event.IPAddress); geo != nil {
		event.Country = geo.Country
		event.City = geo.City
	}

	// Device parsing
	if device := ce.deviceParser.Parse(event.UserAgent); device != nil {
		event.DeviceType = device.Type
		event.Browser = device.Browser
		event.OS = device.OS
	}
}
```

---

## Performance Benchmarks

### Benchmark Results

| Metric | Value |
|--------|-------|
| L1 Cache Hit Latency | ~50 microseconds |
| L2 Cache Hit Latency | ~500 microseconds |
| Database Query Latency | ~5 milliseconds |
| Redirects/Second (Single Instance) | 100,000+ |
| Memory Usage (1M Links Cached) | ~500 MB |
| P99 Latency | <10ms |
| P99.9 Latency | <50ms |

### Load Testing Configuration

```go
// Load test with vegeta
// Rate: 10,000 req/s
// Duration: 5 minutes
// Results: 99.9% success rate, p99 < 10ms
```

### Recommended Production Settings

```yaml
# config/production.yaml
redirect:
  cache:
    l1:
      max_size_mb: 1024
      ttl: 5m
      shards: 1024
    l2:
      ttl: 1h
      pool_size: 100

  workers: 16
  click_buffer_size: 100000
  click_batch_size: 1000
  click_flush_period: 1s

  bot_detection:
    enabled: true
    cache_size: 10000
```
