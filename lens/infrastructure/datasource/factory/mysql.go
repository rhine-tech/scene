package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource/repository"
	"github.com/rhine-tech/scene/model"
	"github.com/rhine-tech/scene/registry"
)

type Mysql struct {
	scene.ModuleFactory
	Config model.DatabaseConfig
}

func (m Mysql) Init() scene.LensInit {
	return func() {
		registry.Register[datasource.MysqlDataSource](
			repository.NewMysqlDatasource(m.Config))
	}
}

func (m Mysql) Default() Mysql {
	return Mysql{
		Config: model.DatabaseConfig{
			Host:     registry.Config.GetString("mysql.db.host"),
			Port:     int(registry.Config.GetInt("mysql.db.port")),
			Username: registry.Config.GetString("mysql.db.username"),
			Password: registry.Config.GetString("mysql.db.password"),
			Database: "scene",
		},
	}
}
