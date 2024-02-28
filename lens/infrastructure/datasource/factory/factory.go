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
	Config        model.DatabaseConfig
	UseApiVersion bool
}

func (m MongoDB) Init() scene.LensInit {
	return func() {
		registry.Register[datasource.MongoDataSource](
			repository.NewMongoDataSource(m.Config, m.UseApiVersion))
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
		UseApiVersion: true,
	}
}

type Redis struct {
	scene.ModuleFactory
	Config model.DatabaseConfig
}

func (r Redis) Init() scene.LensInit {
	return func() {
		registry.Register[datasource.RedisDataSource](
			repository.NewRedisDataRepo(r.Config))
	}
}

func (r Redis) Default() Redis {
	return Redis{
		Config: model.DatabaseConfig{
			Host:     registry.Config.GetString("scene.redis.host"),
			Port:     int(registry.Config.GetInt("scene.redis.port")),
			Database: "0",
		},
	}
}

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
