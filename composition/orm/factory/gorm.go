package factory

import (
	gormSQLite "github.com/glebarez/sqlite"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/registry"
	gormMySQL "gorm.io/driver/mysql"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type gormBackend[T datasource.SqlDataSource] struct {
	dataSource T `aperture:""`
}

func registerGorm[T datasource.SqlDataSource](
	container *registry.Container,
	dialector func(T) gorm.Dialector,
) {
	backend := registry.Load(container, new(gormBackend[T]))
	registry.Export[*orm.Gorm](container, orm.NewGorm(
		func() gorm.Dialector { return dialector(backend.dataSource) },
		func() datasource.DataSource { return backend.dataSource },
	))
}

type GormMysql struct {
	scene.ModuleFactory
}

func (g GormMysql) Init(container *registry.Container) {
	registerGorm(container, func(dataSource datasource.MysqlDataSource) gorm.Dialector {
		return gormMySQL.New(gormMySQL.Config{Conn: dataSource.Connection()})
	})
}

type GormSqlite struct{ scene.ModuleFactory }

func (g GormSqlite) Init(container *registry.Container) {
	registerGorm(container, func(dataSource datasource.SqliteDataSource) gorm.Dialector {
		return &gormSQLite.Dialector{Conn: dataSource.Connection()}
	})
}

type GormPostgreSQL struct{ scene.ModuleFactory }

func (g GormPostgreSQL) Init(container *registry.Container) {
	registerGorm(container, func(dataSource datasource.PostgresDataSource) gorm.Dialector {
		return gormPostgres.New(gormPostgres.Config{Conn: dataSource.Connection()})
	})
}
