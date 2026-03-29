package repository

import (
	"context"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/cache"
)

type GoCache struct {
	c *gocache.Cache
}

func NewGoCache() cache.ICache {
	return &GoCache{
		c: gocache.New(time.Hour, time.Hour*2),
	}
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
			return nil, false, nil
		}
	}
	return nil, false, nil
}

func (g *GoCache) Set(_ context.Context, key string, value []byte, expiration time.Duration, _ ...string) error {
	g.c.Set(key, value, expiration)
	return nil
}

func (g *GoCache) Delete(_ context.Context, keys ...string) error {
	for _, key := range keys {
		g.c.Delete(key)
	}
	return nil
}

func (g *GoCache) InvalidateTags(_ context.Context, _ ...string) error {
	return nil
}
