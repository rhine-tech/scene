package repository

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/simplelru"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/cache"
)

const DefaultLRUCacheSize = 4096

type LRUCacheOptions struct {
	Size       int
	CopyOnRead bool
}

type lruItem struct {
	value    []byte
	expireAt time.Time
	tags     map[string]struct{}
}

type LRUCache struct {
	mu       sync.Mutex
	items    *simplelru.LRU[string, lruItem]
	tagIndex map[string]map[string]struct{}
	opts     LRUCacheOptions
}

func NewLRUCache() cache.ITaggedCache {
	return NewLRUCacheWithOptions(LRUCacheOptions{
		Size: DefaultLRUCacheSize,
	})
}

func NewLRUCacheWithSize(size int) *LRUCache {
	return NewLRUCacheWithOptions(LRUCacheOptions{
		Size: size,
	})
}

func NewLRUCacheWithOptions(opts LRUCacheOptions) *LRUCache {
	if opts.Size <= 0 {
		opts.Size = DefaultLRUCacheSize
	}
	c := &LRUCache{
		tagIndex: make(map[string]map[string]struct{}),
		opts:     opts,
	}
	items, err := simplelru.NewLRU[string, lruItem](opts.Size, func(key string, item lruItem) {
		c.removeIndexLocked(key, item)
	})
	if err != nil {
		panic(err)
	}
	c.items = items
	return c
}

func (l *LRUCache) ImplName() scene.ImplName {
	return cache.Lens.ImplName("ICache", "lru")
}

func (l *LRUCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	item, ok := l.items.Get(key)
	if !ok {
		return nil, false, nil
	}
	if !item.expireAt.IsZero() && time.Now().After(item.expireAt) {
		l.items.Remove(key)
		return nil, false, nil
	}
	if !l.opts.CopyOnRead {
		return item.value, true, nil
	}
	val := make([]byte, len(item.value))
	copy(val, item.value)
	return val, true, nil
}

func (l *LRUCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return l.SetWithTags(ctx, key, value, expiration)
}

func (l *LRUCache) SetWithTags(_ context.Context, key string, value []byte, expiration time.Duration, tags ...string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if old, ok := l.items.Peek(key); ok {
		l.removeIndexLocked(key, old)
	}

	item := lruItem{
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
		}
		if len(item.tags) == 0 {
			item.tags = nil
		}
	}
	for tag := range item.tags {
		keys := l.tagIndex[tag]
		if keys == nil {
			keys = make(map[string]struct{})
			l.tagIndex[tag] = keys
		}
		keys[key] = struct{}{}
	}
	l.items.Add(key, item)
	return nil
}

func (l *LRUCache) Delete(_ context.Context, keys ...string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, key := range keys {
		l.items.Remove(key)
	}
	return nil
}

func (l *LRUCache) InvalidateTags(_ context.Context, tags ...string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	uniq := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		uniq[tag] = struct{}{}
	}
	for tag := range uniq {
		keys := l.tagIndex[tag]
		for key := range keys {
			l.items.Remove(key)
		}
	}
	return nil
}

func (l *LRUCache) removeIndexLocked(key string, item lruItem) {
	for tag := range item.tags {
		keys := l.tagIndex[tag]
		if keys == nil {
			continue
		}
		delete(keys, key)
		if len(keys) == 0 {
			delete(l.tagIndex, tag)
		}
	}
}
