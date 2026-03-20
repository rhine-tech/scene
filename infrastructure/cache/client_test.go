package cache

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rhine-tech/scene"
)

type memoryStore struct {
	mu      sync.RWMutex
	items   map[string]memoryItem
	lastTTL time.Duration
}

type memoryItem struct {
	val    []byte
	expire time.Time
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		items: make(map[string]memoryItem),
	}
}

func (m *memoryStore) ImplName() scene.ImplName {
	return Lens.ImplName("ICache", "memory")
}

func (m *memoryStore) Get(_ context.Context, key string) ([]byte, bool, error) {
	m.mu.RLock()
	item, ok := m.items[key]
	m.mu.RUnlock()
	if !ok {
		return nil, false, nil
	}
	if !item.expire.IsZero() && time.Now().After(item.expire) {
		m.mu.Lock()
		delete(m.items, key)
		m.mu.Unlock()
		return nil, false, nil
	}
	return item.val, true, nil
}

func (m *memoryStore) Set(_ context.Context, key string, value []byte, ttl time.Duration, _ ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item := memoryItem{val: value}
	if ttl > 0 {
		item.expire = time.Now().Add(ttl)
	}
	m.items[key] = item
	m.lastTTL = ttl
	return nil
}

func (m *memoryStore) Delete(_ context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, key := range keys {
		delete(m.items, key)
	}
	return nil
}

func (m *memoryStore) InvalidateTags(_ context.Context, _ ...string) error {
	return nil
}

func TestGetOrLoadHitAndMiss(t *testing.T) {
	store := newMemoryStore()
	client := NewClient(store, WithTTLJitter(0))
	ctx := context.Background()

	var loads int32
	loader := func(context.Context) (string, error) {
		atomic.AddInt32(&loads, 1)
		return "v1", nil
	}

	policy := GetOrLoadPolicy[string]{TTL: time.Second}
	v, err := GetOrLoad(ctx, client, "k", policy, loader)
	if err != nil {
		t.Fatalf("first load failed: %v", err)
	}
	if v != "v1" {
		t.Fatalf("unexpected first value: %s", v)
	}
	v, err = GetOrLoad(ctx, client, "k", policy, loader)
	if err != nil {
		t.Fatalf("second load failed: %v", err)
	}
	if v != "v1" {
		t.Fatalf("unexpected second value: %s", v)
	}
	if got := atomic.LoadInt32(&loads); got != 1 {
		t.Fatalf("loader called %d times, want 1", got)
	}
}

func TestGetOrLoadSingleflight(t *testing.T) {
	store := newMemoryStore()
	client := NewClient(store, WithTTLJitter(0))
	ctx := context.Background()

	var loads int32
	loader := func(context.Context) (string, error) {
		atomic.AddInt32(&loads, 1)
		time.Sleep(20 * time.Millisecond)
		return "ok", nil
	}

	policy := GetOrLoadPolicy[string]{TTL: time.Second}
	const n = 20
	errCh := make(chan error, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			v, err := GetOrLoad(ctx, client, "sf-key", policy, loader)
			if err != nil {
				errCh <- err
				return
			}
			if v != "ok" {
				errCh <- fmt.Errorf("unexpected value: %s", v)
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("unexpected concurrent error: %v", err)
		}
	}
	if got := atomic.LoadInt32(&loads); got != 1 {
		t.Fatalf("loader called %d times, want 1", got)
	}
}

func TestGetOrLoadTTLJitterRange(t *testing.T) {
	store := newMemoryStore()
	client := NewClient(store, WithTTLJitter(0.2))
	ctx := context.Background()

	baseTTL := 10 * time.Second
	_, err := GetOrLoad(ctx, client, "jitter-key", GetOrLoadPolicy[string]{TTL: baseTTL}, func(context.Context) (string, error) {
		return "v", nil
	})
	if err != nil {
		t.Fatalf("GetOrLoad failed: %v", err)
	}

	minTTL := time.Duration(float64(baseTTL) * 0.8)
	maxTTL := time.Duration(float64(baseTTL) * 1.2)
	if store.lastTTL < minTTL || store.lastTTL > maxTTL {
		t.Fatalf("ttl with jitter out of range: got=%s, want in [%s, %s]", store.lastTTL, minTTL, maxTTL)
	}
}
