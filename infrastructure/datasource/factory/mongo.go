package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
	"github.com/rhine-tech/scene/model"
	"github.com/rhine-tech/scene/registry"
)

type MongoDB struct {
	scene.ModuleFactory
	Config    model.DatabaseConfig
	UseApiVer bool
}

func (m MongoDB) Init() scene.LensInit {
	return func() {
		registry.Register[datasource.MongoDataSource](
			datasources.NewMongoDataSource(m.Config, m.UseApiVer))
	}
}

func (m MongoDB) Default() MongoDB {
	return MongoDB{
		Config: model.DatabaseConfig{
			Host:     registry.Config.GetString("mongodb.host"),
			Port:     int(registry.Config.GetInt("mongodb.port")),
			Username: registry.Config.GetString("mongodb.username"),
			Password: registry.Config.GetString("mongodb.password"),
			Database: "scene",
		},
		UseApiVer: true,
	}
}

func (m MongoDB) UseApiVersion(value bool) MongoDB {
	m.UseApiVer = value
	return m
}
