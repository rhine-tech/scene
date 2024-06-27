package repository

import (
	gocache "github.com/patrickmn/go-cache"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/cache"
	"time"
)

type GoCache struct {
	c *gocache.Cache
}

func NewGoCache() cache.ICache {
	return &GoCache{
		c: gocache.New(time.Hour, time.Hour*2),
	}
}

func (g *GoCache) RepoImplName() scene.ImplName {
	return cache.Lens.ImplName("ICache", "go-cache")
}

func (g *GoCache) Get(key cache.CacheKey) (string, bool) {
	foo, found := g.c.Get(string(key))
	if found {
		return foo.(string), found
	}
	return "", false
}

func (g *GoCache) Set(key cache.CacheKey, value string, expiration time.Duration) error {
	g.c.Set(string(key), value, expiration)
	return nil
}

func (g *GoCache) Delete(key cache.CacheKey) error {
	g.c.Delete(string(key))
	return nil
}
