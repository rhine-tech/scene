package repository

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/rhine-tech/scene/infrastructure/cache"
)

func benchmarkStores() map[string]func() cache.ITaggedCache {
	return map[string]func() cache.ITaggedCache{
		"memory": func() cache.ITaggedCache {
			return NewMemoryCache()
		},
		"lru": func() cache.ITaggedCache {
			return NewLRUCache()
		},
	}
}

func BenchmarkCacheGetHit(b *testing.B) {
	ctx := context.Background()
	value := []byte("value")

	for name, newStore := range benchmarkStores() {
		b.Run(name, func(b *testing.B) {
			store := newStore()
			if err := store.Set(ctx, "k", value, time.Hour); err != nil {
				b.Fatalf("prefill failed: %v", err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, hit, err := store.Get(ctx, "k")
				if err != nil {
					b.Fatalf("get failed: %v", err)
				}
				if !hit {
					b.Fatal("unexpected miss")
				}
			}
		})
	}
}

func BenchmarkCacheSetOverwrite(b *testing.B) {
	ctx := context.Background()
	value := []byte("value")

	for name, newStore := range benchmarkStores() {
		b.Run(name, func(b *testing.B) {
			store := newStore()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := store.Set(ctx, "k", value, time.Hour); err != nil {
					b.Fatalf("set failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkCacheSetNewKey(b *testing.B) {
	ctx := context.Background()
	value := []byte("value")

	for name, newStore := range benchmarkStores() {
		b.Run(name, func(b *testing.B) {
			store := newStore()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := "k:" + strconv.Itoa(i)
				if err := store.Set(ctx, key, value, time.Hour); err != nil {
					b.Fatalf("set failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkCacheSetWithTags(b *testing.B) {
	ctx := context.Background()
	value := []byte("value")
	tags := []string{"tag:user:1", "tag:list", "tag:facet"}

	for name, newStore := range benchmarkStores() {
		b.Run(name, func(b *testing.B) {
			store := newStore()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := "k:" + strconv.Itoa(i)
				if err := store.SetWithTags(ctx, key, value, time.Hour, tags...); err != nil {
					b.Fatalf("set failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkCacheDelete(b *testing.B) {
	ctx := context.Background()
	value := []byte("value")

	for name, newStore := range benchmarkStores() {
		b.Run(name, func(b *testing.B) {
			store := newStore()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := "k:" + strconv.Itoa(i)
				if err := store.Set(ctx, key, value, time.Hour); err != nil {
					b.Fatalf("set failed: %v", err)
				}
				if err := store.Delete(ctx, key); err != nil {
					b.Fatalf("delete failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkCacheInvalidateTags(b *testing.B) {
	ctx := context.Background()
	value := []byte("value")
	const keysPerTag = 16

	for name, newStore := range benchmarkStores() {
		b.Run(name, func(b *testing.B) {
			store := newStore()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				tag := "tag:invalidate:" + strconv.Itoa(i)
				for j := 0; j < keysPerTag; j++ {
					key := "k:" + strconv.Itoa(i) + ":" + strconv.Itoa(j)
					if err := store.SetWithTags(ctx, key, value, time.Hour, tag); err != nil {
						b.Fatalf("prefill failed: %v", err)
					}
				}
				b.StartTimer()
				if err := store.InvalidateTags(ctx, tag); err != nil {
					b.Fatalf("invalidate failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkCacheGetParallelHit(b *testing.B) {
	ctx := context.Background()
	value := []byte("value")

	for name, newStore := range benchmarkStores() {
		b.Run(name, func(b *testing.B) {
			store := newStore()
			if err := store.Set(ctx, "k", value, time.Hour); err != nil {
				b.Fatalf("prefill failed: %v", err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, hit, err := store.Get(ctx, "k")
					if err != nil {
						b.Fatalf("get failed: %v", err)
					}
					if !hit {
						b.Fatal("unexpected miss")
					}
				}
			})
		})
	}
}
