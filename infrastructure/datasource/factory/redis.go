package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
	"github.com/rhine-tech/scene/registry"
)

type Redis struct {
	scene.ModuleFactory
	Config datasource.DatabaseConfig
}

func (r Redis) Init() scene.LensInit {
	return func() {
		registry.Register[datasource.RedisDataSource](
			datasources.NewRedisDataRepo(r.Config))
	}
}

func (r Redis) Default() Redis {
	return Redis{
		Config: datasource.DatabaseConfig{
			Host:     registry.Config.GetString("redis.host"),
			Port:     int(registry.Config.GetInt("redis.port")),
			Database: "0",
		},
	}
}
