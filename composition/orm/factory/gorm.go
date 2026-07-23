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
		registry.Register[*orm.Gorm](orm.NewGormWithMySQL(registry.Use(datasource.MysqlDataSource(nil))))
	}
}

func (g GormMysql) Apps() []any {
	return nil
}

type GormSqlite struct{}

func (g GormSqlite) Init() scene.LensInit {
	return func() {
		registry.Register[*orm.Gorm](orm.NewGormWithSQLite(registry.Use(datasource.SqliteDataSource(nil))))
	}
}

func (g GormSqlite) Apps() []any {
	return nil
}

type GormPostgreSQL struct{}

func (g GormPostgreSQL) Init() scene.LensInit {
	return func() {
		registry.Register[*orm.Gorm](
			orm.NewGormWithPostgreSQL(registry.Use(datasource.PostgresDataSource(nil))),
		)
	}
}

func (g GormPostgreSQL) Apps() []any {
	return nil
}
