package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
	"github.com/rhine-tech/scene/registry"
)

type Postgres struct {
	scene.ModuleFactory
	Config datasource.PostgresConfig
}

func (p Postgres) Init() scene.LensInit {
	return func() {
		registry.Register[datasource.PostgresDataSource](
			datasources.NewPostgresDataSource(p.Config),
		)
	}
}

func (p Postgres) Default() Postgres {
	port := int(registry.Config.GetInt("postgres.port"))
	if port == 0 {
		port = 5432
	}
	database := registry.Config.GetString("postgres.database")
	if database == "" {
		database = "scene"
	}
	return Postgres{
		Config: datasource.PostgresConfig{
			Host:     registry.Config.GetString("postgres.host"),
			Port:     port,
			Username: registry.Config.GetString("postgres.username"),
			Password: registry.Config.GetString("postgres.password"),
			Database: database,
			Options:  registry.Config.GetString("postgres.options"),
		},
	}
}
