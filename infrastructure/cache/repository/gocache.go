package repository

import (
	"context"
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/cache"
)

type GoCache struct {
	c  *gocache.Cache
	mu sync.Mutex
	// tagIndex maps tag -> keys
	tagIndex map[string]map[string]struct{}
	// keyTags maps key -> tags
	keyTags map[string]map[string]struct{}
}

func NewGoCache() cache.ICache {
	g := &GoCache{
		c: gocache.New(time.Hour, time.Hour*2),
		tagIndex: make(map[string]map[string]struct{}),
		keyTags:  make(map[string]map[string]struct{}),
	}
	g.c.OnEvicted(func(key string, _ interface{}) {
		g.mu.Lock()
		defer g.mu.Unlock()
		g.removeIndexLocked(key)
	})
	return g
}

func (g *GoCache) ImplName() scene.ImplName {
	return cache.Lens.ImplName("ICache", "go-cache")
}

func (g *GoCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	foo, found := g.c.Get(key)
	if found {
		switch v := foo.(type) {
		case []byte:
			return v, true, nil
		case string:
			return []byte(v), true, nil
		default:
			g.mu.Lock()
			g.removeIndexLocked(key)
			g.mu.Unlock()
			return nil, false, nil
		}
	}
	g.mu.Lock()
	g.removeIndexLocked(key)
	g.mu.Unlock()
	return nil, false, nil
}

func (g *GoCache) Set(_ context.Context, key string, value []byte, expiration time.Duration, tags ...string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.removeIndexLocked(key)
	if len(tags) > 0 {
		keyTagSet := make(map[string]struct{}, len(tags))
		for _, tag := range tags {
			if tag == "" {
				continue
			}
			keyTagSet[tag] = struct{}{}
		}
		if len(keyTagSet) > 0 {
			g.keyTags[key] = keyTagSet
			for tag := range keyTagSet {
				keys := g.tagIndex[tag]
				if keys == nil {
					keys = make(map[string]struct{})
					g.tagIndex[tag] = keys
				}
				keys[key] = struct{}{}
			}
		}
	}
	g.c.Set(key, value, expiration)
	return nil
}

func (g *GoCache) Delete(_ context.Context, keys ...string) error {
	for _, key := range keys {
		g.c.Delete(key)
	}
	return nil
}

func (g *GoCache) InvalidateTags(_ context.Context, tags ...string) error {
	g.mu.Lock()
	uniq := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		uniq[tag] = struct{}{}
	}
	keysToDelete := make(map[string]struct{})
	for tag := range uniq {
		keys := g.tagIndex[tag]
		for key := range keys {
			keysToDelete[key] = struct{}{}
		}
	}
	g.mu.Unlock()

	for key := range keysToDelete {
		g.c.Delete(key)
	}
	return nil
}

func (g *GoCache) removeIndexLocked(key string) {
	tags := g.keyTags[key]
	if tags == nil {
		return
	}
	delete(g.keyTags, key)
	for tag := range tags {
		keys := g.tagIndex[tag]
		if keys == nil {
			continue
		}
		delete(keys, key)
		if len(keys) == 0 {
			delete(g.tagIndex, tag)
		}
	}
}
