package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/registry"
)

type GormMysql struct {
}

func (g GormMysql) Init() scene.LensInit {
	return func() {
		registry.Register[orm.Gorm](orm.GormWithMysql(registry.Use(datasource.MysqlDataSource(nil))))
	}
}

func (g GormMysql) Apps() []any {
	return nil
}

type GormSqlite struct{}

func (g GormSqlite) Init() scene.LensInit {
	return func() {
		registry.Register[orm.Gorm](orm.GormWithSqlite(registry.Use(datasource.SqliteDataSource(nil))))
	}
}

func (g GormSqlite) Apps() []any {
	return nil
}
