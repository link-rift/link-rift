package redirect

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func makeCachedLink(shortCode string) *CachedLink {
	return &CachedLink{
		ID:             uuid.New(),
		ShortCode:      shortCode,
		DestinationURL: "https://example.com/" + shortCode,
		IsActive:       true,
	}
}

func TestL1Cache_SetGet(t *testing.T) {
	c := &Cache{l1TTL: 5 * time.Minute}

	link := makeCachedLink("abc123")
	c.SetL1("abc123", link)

	got, ok := c.GetL1("abc123")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.ShortCode != "abc123" {
		t.Errorf("expected short code abc123, got %s", got.ShortCode)
	}
}

func TestL1Cache_Miss(t *testing.T) {
	c := &Cache{l1TTL: 5 * time.Minute}

	_, ok := c.GetL1("missing")
	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestL1Cache_Expiration(t *testing.T) {
	c := &Cache{l1TTL: 1 * time.Millisecond}

	link := makeCachedLink("expire")
	c.SetL1("expire", link)

	time.Sleep(5 * time.Millisecond)

	_, ok := c.GetL1("expire")
	if ok {
		t.Fatal("expected cache miss after expiration")
	}
}

func TestL1Cache_Invalidate(t *testing.T) {
	c := &Cache{l1TTL: 5 * time.Minute}

	link := makeCachedLink("del")
	c.SetL1("del", link)

	c.l1.Delete("del")

	_, ok := c.GetL1("del")
	if ok {
		t.Fatal("expected cache miss after invalidation")
	}
}

func TestL1Cache_Overwrite(t *testing.T) {
	c := &Cache{l1TTL: 5 * time.Minute}

	link1 := makeCachedLink("overwrite")
	c.SetL1("overwrite", link1)

	link2 := &CachedLink{
		ID:             uuid.New(),
		ShortCode:      "overwrite",
		DestinationURL: "https://new-destination.com",
		IsActive:       true,
	}
	c.SetL1("overwrite", link2)

	got, ok := c.GetL1("overwrite")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.DestinationURL != "https://new-destination.com" {
		t.Errorf("expected new destination, got %s", got.DestinationURL)
	}
}

// --- Benchmarks ---

func BenchmarkCacheGetL1_Hit(b *testing.B) {
	c := &Cache{l1TTL: 5 * time.Minute}
	link := makeCachedLink("bench")
	c.SetL1("bench", link)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.GetL1("bench")
	}
}

func BenchmarkCacheGetL1_Miss(b *testing.B) {
	c := &Cache{l1TTL: 5 * time.Minute}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.GetL1("nonexistent")
	}
}

func BenchmarkCacheSetL1(b *testing.B) {
	c := &Cache{l1TTL: 5 * time.Minute}
	link := makeCachedLink("bench-set")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetL1("bench-set", link)
	}
}
