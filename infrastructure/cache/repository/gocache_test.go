package repository

import (
	"context"
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

func TestGoCacheSetGetDelete(t *testing.T) {
	c := &GoCache{c: gocache.New(time.Second, 2*time.Second)}
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

func TestGoCacheTTLExpiration(t *testing.T) {
	c := &GoCache{c: gocache.New(5*time.Millisecond, 5*time.Millisecond)}
	ctx := context.Background()

	if err := c.Set(ctx, "ttl", []byte("v"), 5*time.Millisecond); err != nil {
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
