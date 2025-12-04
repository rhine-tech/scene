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

func (g *GoCache) ImplName() scene.ImplName {
	return cache.Lens.ImplName("ICache", "go-cache")
}

func (g *GoCache) Get(key string) (string, bool) {
	foo, found := g.c.Get(key)
	if found {
		return foo.(string), found
	}
	return "", false
}

func (g *GoCache) Set(key string, value string, expiration time.Duration) error {
	g.c.Set(key, value, expiration)
	return nil
}

func (g *GoCache) Delete(key string) error {
	g.c.Delete(key)
	return nil
}
