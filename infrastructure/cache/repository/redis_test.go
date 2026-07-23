package repository

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
)

func TestRedisCacheTaggedInvalidation(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewRedisDataRepo(datasource.RedisConfig{
		Host:     "127.0.0.1",
		Port:     6379,
		Database: 0,
	})
	if err := ds.Status(); err != nil {
		t.Skipf("redis not available: %v", err)
	}
	c := NewRedisCache(ds)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	key := "scene:test:cache:redis:" + suffix
	tag := "scene:test:cache:redis:tag:" + suffix
	defer func() { _ = c.Delete(ctx, key) }()

	if err := c.SetWithTags(ctx, key, []byte("v1"), time.Minute, tag); err != nil {
		t.Fatalf("set with tags failed: %v", err)
	}
	got, hit, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !hit || string(got) != "v1" {
		t.Fatalf("unexpected first get, hit=%v got=%q", hit, string(got))
	}

	if err := c.InvalidateTags(ctx, tag); err != nil {
		t.Fatalf("invalidate failed: %v", err)
	}
	_, hit, err = c.Get(ctx, key)
	if err != nil {
		t.Fatalf("get after invalidate failed: %v", err)
	}
	if hit {
		t.Fatal("expected miss after tag invalidation")
	}

	if err := c.SetWithTags(ctx, key, []byte("v2"), time.Minute, tag); err != nil {
		t.Fatalf("set after invalidate failed: %v", err)
	}
	got, hit, err = c.Get(ctx, key)
	if err != nil {
		t.Fatalf("get after reset failed: %v", err)
	}
	if !hit || string(got) != "v2" {
		t.Fatalf("unexpected second get, hit=%v got=%q", hit, string(got))
	}
}

func TestRedisCachePlainValueCompatibility(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewRedisDataRepo(datasource.RedisConfig{
		Host:     "127.0.0.1",
		Port:     6379,
		Database: 0,
	})
	if err := ds.Status(); err != nil {
		t.Skipf("redis not available: %v", err)
	}
	c := NewRedisCache(ds)
	key := "scene:test:cache:redis:plain:" + strconv.FormatInt(time.Now().UnixNano(), 10)
	defer func() { _ = c.Delete(ctx, key) }()

	if err := c.Set(ctx, key, []byte("plain"), time.Minute); err != nil {
		t.Fatalf("plain set failed: %v", err)
	}
	got, hit, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("plain get failed: %v", err)
	}
	if !hit || string(got) != "plain" {
		t.Fatalf("unexpected plain get, hit=%v got=%q", hit, string(got))
	}
}
