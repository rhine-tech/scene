package repository

import (
	"context"
	"testing"
	"time"
)

func TestLRUCacheEvictsOldest(t *testing.T) {
	c := NewLRUCacheWithSize(2)
	ctx := context.Background()

	if err := c.SetWithTags(ctx, "a", []byte("1"), time.Minute, "tag:a"); err != nil {
		t.Fatalf("set a failed: %v", err)
	}
	if err := c.SetWithTags(ctx, "b", []byte("2"), time.Minute, "tag:b"); err != nil {
		t.Fatalf("set b failed: %v", err)
	}
	if _, hit, err := c.Get(ctx, "a"); err != nil || !hit {
		t.Fatalf("expected a hit before eviction, hit=%v err=%v", hit, err)
	}
	if err := c.SetWithTags(ctx, "c", []byte("3"), time.Minute, "tag:c"); err != nil {
		t.Fatalf("set c failed: %v", err)
	}

	if _, hit, err := c.Get(ctx, "b"); err != nil {
		t.Fatalf("get b failed: %v", err)
	} else if hit {
		t.Fatal("expected b to be evicted")
	}
	if _, hit, err := c.Get(ctx, "a"); err != nil || !hit {
		t.Fatalf("expected a to remain after being recently used, hit=%v err=%v", hit, err)
	}
}

func TestLRUCacheInvalidateTagsAndCleanup(t *testing.T) {
	c := NewLRUCacheWithSize(4)
	ctx := context.Background()

	if err := c.SetWithTags(ctx, "a", []byte("1"), time.Minute, "tag:shared"); err != nil {
		t.Fatalf("set a failed: %v", err)
	}
	if err := c.SetWithTags(ctx, "b", []byte("2"), time.Minute, "tag:shared"); err != nil {
		t.Fatalf("set b failed: %v", err)
	}
	if err := c.InvalidateTags(ctx, "tag:shared"); err != nil {
		t.Fatalf("invalidate failed: %v", err)
	}

	for _, key := range []string{"a", "b"} {
		if _, hit, err := c.Get(ctx, key); err != nil {
			t.Fatalf("get %s failed: %v", key, err)
		} else if hit {
			t.Fatalf("expected %s to be invalidated", key)
		}
	}
	if keys := c.tagIndex["tag:shared"]; len(keys) > 0 {
		t.Fatal("expected tag index to be cleaned")
	}
}

func TestLRUCacheTTLExpiration(t *testing.T) {
	c := NewLRUCacheWithSize(2)
	ctx := context.Background()

	if err := c.SetWithTags(ctx, "ttl", []byte("v"), 10*time.Millisecond, "tag:ttl"); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	if _, hit, err := c.Get(ctx, "ttl"); err != nil {
		t.Fatalf("get failed: %v", err)
	} else if hit {
		t.Fatal("expected ttl key to expire")
	}
	if keys := c.tagIndex["tag:ttl"]; len(keys) > 0 {
		t.Fatal("expected expired key to be removed from tag index")
	}
}

func TestLRUCacheCopyOnReadSafeMode(t *testing.T) {
	c := NewLRUCacheWithOptions(LRUCacheOptions{Size: 2, CopyOnRead: true})
	ctx := context.Background()

	if err := c.Set(ctx, "k", []byte("abc"), time.Minute); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	got, hit, err := c.Get(ctx, "k")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !hit {
		t.Fatal("expected hit")
	}
	got[0] = 'z'
	got2, hit, err := c.Get(ctx, "k")
	if err != nil {
		t.Fatalf("get2 failed: %v", err)
	}
	if !hit {
		t.Fatal("expected hit")
	}
	if string(got2) != "abc" {
		t.Fatalf("safe mode should isolate mutation, got=%s", string(got2))
	}
}
