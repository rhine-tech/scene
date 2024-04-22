package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
	"github.com/rhine-tech/scene/model"
	"github.com/rhine-tech/scene/registry"
)

type Redis struct {
	scene.ModuleFactory
	Config model.DatabaseConfig
}

func (r Redis) Init() scene.LensInit {
	return func() {
		registry.Register[datasource.RedisDataSource](
			datasources.NewRedisDataRepo(r.Config))
	}
}

func (r Redis) Default() Redis {
	return Redis{
		Config: model.DatabaseConfig{
			Host:     registry.Config.GetString("scene.redis.host"),
			Port:     int(registry.Config.GetInt("scene.redis.port")),
			Database: "0",
		},
	}
}
