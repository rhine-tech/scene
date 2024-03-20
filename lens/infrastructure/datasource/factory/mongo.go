package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource/repository"
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
			repository.NewMongoDataSource(m.Config, m.UseApiVer))
	}
}

func (m MongoDB) Default() MongoDB {
	return MongoDB{
		Config: model.DatabaseConfig{
			Host:     registry.Config.GetString("scene.db.host"),
			Port:     int(registry.Config.GetInt("scene.db.port")),
			Username: registry.Config.GetString("scene.db.username"),
			Password: registry.Config.GetString("scene.db.password"),
			Database: "scene",
		},
		UseApiVer: true,
	}
}

func (m MongoDB) UseApiVersion(value bool) MongoDB {
	m.UseApiVer = value
	return m
}
