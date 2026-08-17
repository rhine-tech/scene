package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
	"github.com/rhine-tech/scene/registry"
)

type MongoDB struct {
	scene.ModuleFactory
	Config datasource.MongoConfig
}

func (m MongoDB) Init(container *registry.Container) {
	registry.Export[datasource.MongoDataSource](container, datasources.NewMongoDataSource(m.Config))
}

func (m MongoDB) Default() MongoDB {
	cfg := registry.Use[config.IConfig](nil)
	return MongoDB{
		Config: datasource.MongoConfig{
			Host:          cfg.GetString("mongodb.host"),
			Port:          int(cfg.GetInt("mongodb.port")),
			Username:      cfg.GetString("mongodb.username"),
			Password:      cfg.GetString("mongodb.password"),
			Database:      "scene",
			AuthSource:    cfg.GetString("mongodb.auth_source"),
			UseAPIVersion: true,
		},
	}
}

func (m MongoDB) UseApiVersion(value bool) MongoDB {
	m.Config.UseAPIVersion = value
	return m
}
