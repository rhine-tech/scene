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

func (r RedisCache) Init(container *registry.Container) {
	cacheImpl := repository.NewRedisCache(nil)
	registry.Export[cache.ICache](container, cacheImpl)
	registry.Export[cache.ITaggedCache](container, cacheImpl)
}

type MemoryCache struct {
	scene.ModuleFactory
}

func (m MemoryCache) Init(container *registry.Container) {
	cacheImpl := repository.NewMemoryCache()
	registry.Export[cache.ICache](container, cacheImpl)
	registry.Export[cache.ITaggedCache](container, cacheImpl)
}

type LRUCache struct {
	scene.ModuleFactory
	Size int
}

func (l LRUCache) Init(container *registry.Container) {
	var cacheImpl cache.ITaggedCache
	if l.Size > 0 {
		cacheImpl = repository.NewLRUCacheWithSize(l.Size)
	} else {
		cacheImpl = repository.NewLRUCache()
	}
	registry.Export[cache.ICache](container, cacheImpl)
	registry.Export[cache.ITaggedCache](container, cacheImpl)
}
