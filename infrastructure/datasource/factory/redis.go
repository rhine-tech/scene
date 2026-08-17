package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
	"github.com/rhine-tech/scene/registry"
)

type Redis struct {
	scene.ModuleFactory
	Config datasource.RedisConfig
}

func (r Redis) Init(container *registry.Container) {
	registry.Export[datasource.RedisDataSource](container, datasources.NewRedisDataRepo(r.Config))
}

func (r Redis) Default() Redis {
	cfg := registry.Use[config.IConfig](nil)
	return Redis{
		Config: datasource.RedisConfig{
			Host:     cfg.GetString("redis.host"),
			Port:     int(cfg.GetInt("redis.port")),
			Database: 0,
		},
	}
}
