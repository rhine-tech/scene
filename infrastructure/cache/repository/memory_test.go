package repository

import (
	"context"
	"strconv"
	"sync"
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

func TestMemoryCacheCopyOnReadSafeMode(t *testing.T) {
	c := NewMemoryCacheWithOptions(MemoryCacheOptions{CopyOnRead: true})
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

func TestMemoryCacheNoCopyFastMode(t *testing.T) {
	c := NewMemoryCacheWithOptions(MemoryCacheOptions{CopyOnRead: false})
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
	if string(got2) != "zbc" {
		t.Fatalf("fast mode should avoid copy on read, got=%s", string(got2))
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

func TestMemoryCacheInvalidateTagsEmptyAndDuplicate(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()
	if err := c.Set(ctx, "a", []byte("1"), time.Minute, "", "tag:a"); err != nil {
		t.Fatalf("set a failed: %v", err)
	}
	if err := c.InvalidateTags(ctx, "", "tag:a", "tag:a"); err != nil {
		t.Fatalf("invalidate tags failed: %v", err)
	}
	_, hit, err := c.Get(ctx, "a")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if hit {
		t.Fatal("expected key a to be invalidated")
	}
}

func TestMemoryCacheConcurrentAccess(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				key := "k:" + strconv.Itoa(id) + ":" + strconv.Itoa(j%16)
				_ = c.Set(ctx, key, []byte("v"), time.Second, "tag:"+strconv.Itoa(j%4))
				_, _, _ = c.Get(ctx, key)
				if j%5 == 0 {
					_ = c.Delete(ctx, key)
				}
				if j%11 == 0 {
					_ = c.InvalidateTags(ctx, "tag:"+strconv.Itoa(j%4))
				}
			}
		}(i)
	}
	wg.Wait()
}
