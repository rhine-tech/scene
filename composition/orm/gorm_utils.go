package orm

import (
	gormSQLite "github.com/glebarez/sqlite"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	gormMySQL "gorm.io/driver/mysql"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewGormWithMySQL creates a GORM component backed by a MySQL data source.
func NewGormWithMySQL(ds datasource.MysqlDataSource) *Gorm {
	return NewGorm(func() gorm.Dialector {
		return gormMySQL.New(gormMySQL.Config{
			Conn: ds.Connection(),
		})
	}, func() datasource.DataSource {
		return ds
	})
}

// NewGormWithSQLite creates a GORM component backed by a SQLite data source.
func NewGormWithSQLite(ds datasource.SqliteDataSource) *Gorm {
	return NewGorm(func() gorm.Dialector {
		return &gormSQLite.Dialector{
			Conn: ds.Connection(),
		}
	}, func() datasource.DataSource {
		return ds
	})
}

// NewGormWithPostgreSQL creates a GORM component backed by a PostgreSQL data source.
func NewGormWithPostgreSQL(ds datasource.PostgresDataSource) *Gorm {
	return NewGorm(func() gorm.Dialector {
		return gormPostgres.New(gormPostgres.Config{
			Conn: ds.Connection(),
		})
	}, func() datasource.DataSource {
		return ds
	})
}
