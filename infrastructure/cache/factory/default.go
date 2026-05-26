package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/cache"
	"github.com/rhine-tech/scene/infrastructure/cache/repository"
	"github.com/rhine-tech/scene/registry"
)

type RedisCache struct {
	scene.ModuleFactory
}

func (r RedisCache) Init() scene.LensInit {
	return func() {
		c := repository.NewRedisCache(nil)
		registry.Register[cache.ICache](c)
		registry.Register[cache.ITaggedCache](c)
	}
}

type MemoryCache struct {
	scene.ModuleFactory
}

func (m MemoryCache) Init() scene.LensInit {
	return func() {
		c := repository.NewMemoryCache()
		registry.Register[cache.ICache](c)
		registry.Register[cache.ITaggedCache](c)
	}
}

type LRUCache struct {
	scene.ModuleFactory
	Size int
}

func (l LRUCache) Init() scene.LensInit {
	return func() {
		var c cache.ITaggedCache
		if l.Size > 0 {
			c = repository.NewLRUCacheWithSize(l.Size)
		} else {
			c = repository.NewLRUCache()
		}
		registry.Register[cache.ICache](c)
		registry.Register[cache.ITaggedCache](c)
	}
}
