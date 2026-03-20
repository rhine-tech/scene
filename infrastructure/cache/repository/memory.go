package repository

import (
	"context"
	"sync"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/cache"
)

type memoryItem struct {
	value    []byte
	expireAt time.Time
	tags     map[string]struct{}
}

type MemoryCacheOptions struct {
	// CopyOnRead returns a cloned value on Get when enabled.
	// Disable it for better read throughput (returned bytes must be treated as read-only).
	CopyOnRead bool
}

type MemoryCache struct {
	mu       sync.RWMutex
	items    map[string]memoryItem
	tagIndex map[string]map[string]struct{}
	opts     MemoryCacheOptions
}

func NewMemoryCache() cache.ICache {
	return NewMemoryCacheWithOptions(MemoryCacheOptions{
		CopyOnRead: false,
	})
}

func NewMemoryCacheWithOptions(opts MemoryCacheOptions) *MemoryCache {
	return &MemoryCache{
		items:    make(map[string]memoryItem),
		tagIndex: make(map[string]map[string]struct{}),
		opts:     opts,
	}
}

func (m *MemoryCache) ImplName() scene.ImplName {
	return cache.Lens.ImplName("ICache", "memory")
}

func (m *MemoryCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	m.mu.RLock()
	item, ok := m.items[key]
	m.mu.RUnlock()
	if !ok {
		return nil, false, nil
	}
	if !item.expireAt.IsZero() && time.Now().After(item.expireAt) {
		m.mu.Lock()
		m.deleteLocked(key)
		m.mu.Unlock()
		return nil, false, nil
	}
	if !m.opts.CopyOnRead {
		return item.value, true, nil
	}
	val := make([]byte, len(item.value))
	copy(val, item.value)
	return val, true, nil
}

func (m *MemoryCache) Set(_ context.Context, key string, value []byte, expiration time.Duration, tags ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.items[key]; exists {
		m.deleteLocked(key)
	}

	item := memoryItem{
		value: make([]byte, len(value)),
	}
	copy(item.value, value)
	if expiration > 0 {
		item.expireAt = time.Now().Add(expiration)
	}
	if len(tags) > 0 {
		item.tags = make(map[string]struct{}, len(tags))
		for _, tag := range tags {
			if tag == "" {
				continue
			}
			item.tags[tag] = struct{}{}
			keys, ok := m.tagIndex[tag]
			if !ok {
				keys = make(map[string]struct{})
				m.tagIndex[tag] = keys
			}
			keys[key] = struct{}{}
		}
	}
	m.items[key] = item
	return nil
}

func (m *MemoryCache) Delete(_ context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, key := range keys {
		m.deleteLocked(key)
	}
	return nil
}

func (m *MemoryCache) InvalidateTags(_ context.Context, tags ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	uniq := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		uniq[tag] = struct{}{}
	}
	for tag := range uniq {
		keys := m.tagIndex[tag]
		for key := range keys {
			m.deleteLocked(key)
		}
	}
	return nil
}

func (m *MemoryCache) deleteLocked(key string) {
	item, ok := m.items[key]
	if !ok {
		return
	}
	delete(m.items, key)
	for tag := range item.tags {
		keys := m.tagIndex[tag]
		delete(keys, key)
		if len(keys) == 0 {
			delete(m.tagIndex, tag)
		}
	}
}
