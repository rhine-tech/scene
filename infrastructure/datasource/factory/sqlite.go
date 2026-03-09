package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
	"github.com/rhine-tech/scene/registry"
)

type Sqlite struct {
	scene.ModuleFactory
	Config datasource.DatabaseConfig
}

func (m Sqlite) Init() scene.LensInit {
	return func() {
		registry.Register[datasource.SqliteDataSource](
			datasources.SqliteDatasource(m.Config))
	}
}

func (m Sqlite) Default() Sqlite {
	return Sqlite{
		Config: datasource.DatabaseConfig{
			Host:     registry.Config.GetString("sqlite.path"),
			Options:  registry.Config.GetString("sqlite.options"),
			Database: "scene",
		},
	}
}
