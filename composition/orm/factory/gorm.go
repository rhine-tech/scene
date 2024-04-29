package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/registry"

	gormImpl "github.com/rhine-tech/scene/composition/orm/internal/gorm"
)

type GormMysql struct {
}

func (g GormMysql) Init() scene.LensInit {
	return func() {
		registry.Register[orm.Gorm](gormImpl.GormWithMysql(registry.Use(datasource.MysqlDataSource(nil))))
	}
}

func (g GormMysql) Apps() []any {
	return nil
}

type GormSqlite struct{}

func (g GormSqlite) Init() scene.LensInit {
	return func() {
		registry.Register[orm.Gorm](gormImpl.GormWithSqlite(registry.Use(datasource.SqliteDataSource(nil))))
	}
}

func (g GormSqlite) Apps() []any {
	return nil
}
