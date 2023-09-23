package builder

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/cache/repository"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource"
	"github.com/rhine-tech/scene/registry"
)

type RedisCache struct {
	scene.Builder
}

func (r RedisCache) Init() scene.LensInit {
	return func() {
		registry.Register(repository.NewRedisCache(
			registry.Use(datasource.RedisDataSource(nil))))
	}
}
