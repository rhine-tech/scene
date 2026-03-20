package repository

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCacheSetGetDelete(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	if err := c.Set(ctx, "k", []byte("v"), time.Second); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	got, hit, err := c.Get(ctx, "k")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !hit {
		t.Fatal("expected cache hit")
	}
	if string(got) != "v" {
		t.Fatalf("unexpected value: %s", string(got))
	}

	if err := c.Delete(ctx, "k"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	_, hit, err = c.Get(ctx, "k")
	if err != nil {
		t.Fatalf("get after delete failed: %v", err)
	}
	if hit {
		t.Fatal("expected cache miss after delete")
	}
}

func TestMemoryCacheTTLExpiration(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	if err := c.Set(ctx, "ttl", []byte("v"), 10*time.Millisecond); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	time.Sleep(20 * time.Millisecond)

	_, hit, err := c.Get(ctx, "ttl")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if hit {
		t.Fatal("expected cache miss after ttl expiration")
	}
}

func TestMemoryCacheInvalidateTags(t *testing.T) {
	c := NewMemoryCache()
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
