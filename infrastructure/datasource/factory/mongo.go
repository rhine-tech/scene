package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
	"github.com/rhine-tech/scene/registry"
)

type MongoDB struct {
	scene.ModuleFactory
	Config datasource.MongoConfig
}

func (m MongoDB) Init() scene.LensInit {
	return func() {
		registry.Register[datasource.MongoDataSource](
			datasources.NewMongoDataSource(m.Config))
	}
}

func (m MongoDB) Default() MongoDB {
	return MongoDB{
		Config: datasource.MongoConfig{
			Host:          registry.Config.GetString("mongodb.host"),
			Port:          int(registry.Config.GetInt("mongodb.port")),
			Username:      registry.Config.GetString("mongodb.username"),
			Password:      registry.Config.GetString("mongodb.password"),
			Database:      "scene",
			AuthSource:    registry.Config.GetString("mongodb.auth_source"),
			UseAPIVersion: true,
		},
	}
}

func (m MongoDB) UseApiVersion(value bool) MongoDB {
	m.Config.UseAPIVersion = value
	return m
}
