package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/cache"
	"github.com/rhine-tech/scene/lens/infrastructure/cache/repository"
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
