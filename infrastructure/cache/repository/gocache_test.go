package repository

import (
	"context"
	"testing"
	"time"
)

func TestGoCacheInvalidateTags(t *testing.T) {
	c := NewGoCache()
	ctx := context.Background()

	if err := c.Set(ctx, "a", []byte("1"), time.Minute, "user:1", "list:user"); err != nil {
		t.Fatalf("set a failed: %v", err)
	}
	if err := c.Set(ctx, "b", []byte("2"), time.Minute, "user:2"); err != nil {
		t.Fatalf("set b failed: %v", err)
	}
	if err := c.InvalidateTags(ctx, "user:1"); err != nil {
		t.Fatalf("invalidate tags failed: %v", err)
	}

	_, hitA, err := c.Get(ctx, "a")
	if err != nil {
		t.Fatalf("get a failed: %v", err)
	}
	_, hitB, err := c.Get(ctx, "b")
	if err != nil {
		t.Fatalf("get b failed: %v", err)
	}
	if hitA {
		t.Fatal("expected key a to be invalidated by tag")
	}
	if !hitB {
		t.Fatal("expected key b to remain")
	}
}

func TestGoCacheSetOverwriteRetags(t *testing.T) {
	c := NewGoCache()
	ctx := context.Background()

	if err := c.Set(ctx, "k", []byte("1"), time.Minute, "tag:old"); err != nil {
		t.Fatalf("set old failed: %v", err)
	}
	if err := c.Set(ctx, "k", []byte("2"), time.Minute, "tag:new"); err != nil {
		t.Fatalf("set new failed: %v", err)
	}
	if err := c.InvalidateTags(ctx, "tag:old"); err != nil {
		t.Fatalf("invalidate old tag failed: %v", err)
	}
	_, hit, err := c.Get(ctx, "k")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !hit {
		t.Fatal("expected key to remain after invalidating old tag")
	}

	if err := c.InvalidateTags(ctx, "tag:new"); err != nil {
		t.Fatalf("invalidate new tag failed: %v", err)
	}
	_, hit, err = c.Get(ctx, "k")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if hit {
		t.Fatal("expected key to be invalidated by new tag")
	}
}

func TestGoCacheEvictedKeyIndexCleanup(t *testing.T) {
	store := NewGoCache()
	g, ok := store.(*GoCache)
	if !ok {
		t.Fatal("expected *GoCache")
	}
	ctx := context.Background()

	if err := g.Set(ctx, "ttl", []byte("v"), 10*time.Millisecond, "tag:ttl"); err != nil {
		t.Fatalf("set ttl failed: %v", err)
	}
	time.Sleep(20 * time.Millisecond)

	_, hit, err := g.Get(ctx, "ttl")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if hit {
		t.Fatal("expected key to expire")
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.keyTags["ttl"]; ok {
		t.Fatal("expected keyTags cleaned after eviction")
	}
	if keys, ok := g.tagIndex["tag:ttl"]; ok && len(keys) > 0 {
		t.Fatal("expected tagIndex cleaned after eviction")
	}
}
