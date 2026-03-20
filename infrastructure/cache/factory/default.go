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
		registry.Register[cache.ICache](repository.NewRedisCache(nil))
	}
}

type MemoryCache struct {
	scene.ModuleFactory
}

func (m MemoryCache) Init() scene.LensInit {
	return func() {
		registry.Register[cache.ICache](repository.NewMemoryCache())
	}
}

type GoCache struct {
	scene.ModuleFactory
}

func (g GoCache) Init() scene.LensInit {
	return func() {
		registry.Register[cache.ICache](repository.NewGoCache())
	}
}
