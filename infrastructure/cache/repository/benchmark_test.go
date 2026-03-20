package repository

import (
	"context"
	"strconv"
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/rhine-tech/scene/infrastructure/cache"
)

func benchmarkStores() map[string]func() cache.ICache {
	return map[string]func() cache.ICache{
		"memory": func() cache.ICache {
			return NewMemoryCache()
		},
		"go-cache": func() cache.ICache {
			return &GoCache{c: gocache.New(time.Hour, 2*time.Hour)}
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
