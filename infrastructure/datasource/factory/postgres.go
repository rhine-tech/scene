package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
	"github.com/rhine-tech/scene/registry"
)

type Postgres struct {
	scene.ModuleFactory
	Config datasource.PostgresConfig
}

func (p Postgres) Init(container *registry.Container) {
	registry.Export[datasource.PostgresDataSource](container, datasources.NewPostgresDataSource(p.Config))
}

func (p Postgres) Default() Postgres {
	cfg := registry.Use[config.IConfig](nil)
	port := int(cfg.GetInt("postgres.port"))
	if port == 0 {
		port = 5432
	}
	database := cfg.GetString("postgres.database")
	if database == "" {
		database = "scene"
	}
	return Postgres{
		Config: datasource.PostgresConfig{
			Host:     cfg.GetString("postgres.host"),
			Port:     port,
			Username: cfg.GetString("postgres.username"),
			Password: cfg.GetString("postgres.password"),
			Database: database,
			Options:  cfg.GetString("postgres.options"),
		},
	}
}
