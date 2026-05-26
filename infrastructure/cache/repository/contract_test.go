package repository

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/rhine-tech/scene/infrastructure/cache"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
)

type cacheContractCase struct {
	name string
	new  func(t *testing.T) cache.ITaggedCache
}

func taggedCacheContractCases() []cacheContractCase {
	return []cacheContractCase{
		{
			name: "memory",
			new: func(t *testing.T) cache.ITaggedCache {
				t.Helper()
				return NewMemoryCache()
			},
		},
		{
			name: "lru",
			new: func(t *testing.T) cache.ITaggedCache {
				t.Helper()
				return NewLRUCacheWithSize(32)
			},
		},
		{
			name: "redis",
			new: func(t *testing.T) cache.ITaggedCache {
				t.Helper()
				ds := datasources.NewRedisDataRepo(datasource.DatabaseConfig{
					Host:     "127.0.0.1",
					Port:     6379,
					Database: "0",
				})
				if err := ds.Status(); err != nil {
					t.Skipf("redis not available: %v", err)
				}
				return NewRedisCache(ds)
			},
		},
	}
}

func TestTaggedCacheContract(t *testing.T) {
	for _, tc := range taggedCacheContractCases() {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.new(t)
			prefix := fmt.Sprintf("scene:test:cache:contract:%s:%d", tc.name, time.Now().UnixNano())
			ctx := context.Background()

			t.Run("set get delete", func(t *testing.T) {
				key := prefix + ":set-get-delete"
				if err := c.Set(ctx, key, []byte("v"), time.Minute); err != nil {
					t.Fatalf("set failed: %v", err)
				}
				assertCacheHit(t, c, ctx, key, "v")
				if err := c.Delete(ctx, key); err != nil {
					t.Fatalf("delete failed: %v", err)
				}
				assertCacheMiss(t, c, ctx, key)
			})

			t.Run("multi delete ignores missing keys", func(t *testing.T) {
				keyA := prefix + ":delete:a"
				keyB := prefix + ":delete:b"
				if err := c.Set(ctx, keyA, []byte("a"), time.Minute); err != nil {
					t.Fatalf("set a failed: %v", err)
				}
				if err := c.Set(ctx, keyB, []byte("b"), time.Minute); err != nil {
					t.Fatalf("set b failed: %v", err)
				}
				if err := c.Delete(ctx, keyA, prefix+":delete:missing", keyB); err != nil {
					t.Fatalf("delete failed: %v", err)
				}
				assertCacheMiss(t, c, ctx, keyA)
				assertCacheMiss(t, c, ctx, keyB)
			})

			t.Run("ttl expiration", func(t *testing.T) {
				key := prefix + ":ttl"
				if err := c.Set(ctx, key, []byte("ttl"), 10*time.Millisecond); err != nil {
					t.Fatalf("set ttl failed: %v", err)
				}
				assertCacheHit(t, c, ctx, key, "ttl")
				time.Sleep(25 * time.Millisecond)
				assertCacheMiss(t, c, ctx, key)
			})

			t.Run("no expiration", func(t *testing.T) {
				key := prefix + ":no-expiration"
				if err := c.Set(ctx, key, []byte("stable"), cache.NoExpiration); err != nil {
					t.Fatalf("set no expiration failed: %v", err)
				}
				time.Sleep(5 * time.Millisecond)
				assertCacheHit(t, c, ctx, key, "stable")
			})

			t.Run("invalidate one tag keeps unrelated keys", func(t *testing.T) {
				keyA := prefix + ":tag:a"
				keyB := prefix + ":tag:b"
				tagA := prefix + ":tag:a"
				tagB := prefix + ":tag:b"
				if err := c.SetWithTags(ctx, keyA, []byte("a"), time.Minute, tagA, prefix+":tag:shared"); err != nil {
					t.Fatalf("set a failed: %v", err)
				}
				if err := c.SetWithTags(ctx, keyB, []byte("b"), time.Minute, tagB); err != nil {
					t.Fatalf("set b failed: %v", err)
				}
				if err := c.InvalidateTags(ctx, tagA); err != nil {
					t.Fatalf("invalidate failed: %v", err)
				}
				assertCacheMiss(t, c, ctx, keyA)
				assertCacheHit(t, c, ctx, keyB, "b")
			})

			t.Run("invalidate shared tag removes all tagged keys", func(t *testing.T) {
				tag := prefix + ":tag:shared-all"
				for i := 0; i < 3; i++ {
					key := prefix + ":shared:" + strconv.Itoa(i)
					if err := c.SetWithTags(ctx, key, []byte("v"+strconv.Itoa(i)), time.Minute, tag); err != nil {
						t.Fatalf("set shared key %d failed: %v", i, err)
					}
				}
				if err := c.InvalidateTags(ctx, tag); err != nil {
					t.Fatalf("invalidate shared failed: %v", err)
				}
				for i := 0; i < 3; i++ {
					assertCacheMiss(t, c, ctx, prefix+":shared:"+strconv.Itoa(i))
				}
			})

			t.Run("empty and duplicate tags are ignored", func(t *testing.T) {
				key := prefix + ":empty-duplicate-tags"
				tag := prefix + ":tag:dedup"
				if err := c.SetWithTags(ctx, key, []byte("dedup"), time.Minute, "", tag, tag); err != nil {
					t.Fatalf("set with duplicate tags failed: %v", err)
				}
				if err := c.InvalidateTags(ctx, "", tag, tag); err != nil {
					t.Fatalf("invalidate duplicate tags failed: %v", err)
				}
				assertCacheMiss(t, c, ctx, key)
			})

			t.Run("overwrite replaces old tags", func(t *testing.T) {
				key := prefix + ":overwrite-retag"
				oldTag := prefix + ":tag:old"
				newTag := prefix + ":tag:new"
				if err := c.SetWithTags(ctx, key, []byte("old"), time.Minute, oldTag); err != nil {
					t.Fatalf("set old tag failed: %v", err)
				}
				if err := c.SetWithTags(ctx, key, []byte("new"), time.Minute, newTag); err != nil {
					t.Fatalf("set new tag failed: %v", err)
				}
				if err := c.InvalidateTags(ctx, oldTag); err != nil {
					t.Fatalf("invalidate old tag failed: %v", err)
				}
				assertCacheHit(t, c, ctx, key, "new")
				if err := c.InvalidateTags(ctx, newTag); err != nil {
					t.Fatalf("invalidate new tag failed: %v", err)
				}
				assertCacheMiss(t, c, ctx, key)
			})

			t.Run("delete removes tag membership", func(t *testing.T) {
				key := prefix + ":delete-removes-tag"
				tag := prefix + ":tag:delete-removes"
				if err := c.SetWithTags(ctx, key, []byte("v1"), time.Minute, tag); err != nil {
					t.Fatalf("set tagged key failed: %v", err)
				}
				if err := c.Delete(ctx, key); err != nil {
					t.Fatalf("delete tagged key failed: %v", err)
				}
				if err := c.Set(ctx, key, []byte("v2"), time.Minute); err != nil {
					t.Fatalf("set plain key failed: %v", err)
				}
				if err := c.InvalidateTags(ctx, tag); err != nil {
					t.Fatalf("invalidate old tag failed: %v", err)
				}
				assertCacheHit(t, c, ctx, key, "v2")
			})
		})
	}
}

func assertCacheHit(t *testing.T, c cache.ICache, ctx context.Context, key string, want string) {
	t.Helper()
	got, hit, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("get %q failed: %v", key, err)
	}
	if !hit {
		t.Fatalf("expected %q to hit", key)
	}
	if string(got) != want {
		t.Fatalf("unexpected value for %q: got=%q want=%q", key, string(got), want)
	}
}

func assertCacheMiss(t *testing.T, c cache.ICache, ctx context.Context, key string) {
	t.Helper()
	_, hit, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("get %q failed: %v", key, err)
	}
	if hit {
		t.Fatalf("expected %q to miss", key)
	}
}
