package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
	"github.com/rhine-tech/scene/registry"
)

type Sqlite struct {
	scene.ModuleFactory
	Config datasource.SqliteConfig
}

func (m Sqlite) Init(container *registry.Container) {
	registry.Export[datasource.SqliteDataSource](container, datasources.SqliteDatasource(m.Config))
}

func (m Sqlite) Default() Sqlite {
	cfg := registry.Use[config.IConfig](nil)
	return Sqlite{
		Config: datasource.SqliteConfig{
			Path:    cfg.GetString("sqlite.path"),
			Options: cfg.GetString("sqlite.options"),
		},
	}
}
