package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
	"github.com/rhine-tech/scene/registry"
)

type Mysql struct {
	scene.ModuleFactory
	Config datasource.MysqlConfig
}

func (m Mysql) Init(container *registry.Container) {
	registry.Export[datasource.MysqlDataSource](container, datasources.NewMysqlDatasource(m.Config))
}

func (m Mysql) Default() Mysql {
	cfg := registry.Use[config.IConfig](nil)
	return Mysql{
		Config: datasource.MysqlConfig{
			Host:     cfg.GetString("mysql.host"),
			Port:     int(cfg.GetInt("mysql.port")),
			Username: cfg.GetString("mysql.username"),
			Password: cfg.GetString("mysql.password"),
			Database: "scene",
		},
	}
}
