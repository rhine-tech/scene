package gorm

import (
	"github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"gorm.io/gorm"

	gormSqlite "github.com/glebarez/sqlite"
	gormMysql "gorm.io/driver/mysql"
)

func GormWithMysql(ds datasource.MysqlDataSource) orm.Gorm {
	return NewImpl(func() gorm.Dialector {
		return gormMysql.New(gormMysql.Config{
			Conn: ds.Connection(),
		})
	}, ds)
}

func GormWithSqlite(ds datasource.SqliteDataSource) orm.Gorm {
	return NewImpl(func() gorm.Dialector {
		return &gormSqlite.Dialector{
			Conn: ds.Connection(),
		}
	}, ds)
}
