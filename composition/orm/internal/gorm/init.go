package gorm

import (
	"github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"gorm.io/gorm"
)

func GormWithMysql(ds datasource.MysqlDataSource) orm.Gorm {
	return NewImpl(func() gorm.Dialector {
		return orm.GormDialectorMySql(ds)
	}, ds)
}

func GormWithSqlite(ds datasource.SqliteDataSource) orm.Gorm {
	return NewImpl(func() gorm.Dialector {
		return orm.GormDialectorSqlite(ds)
	}, ds)
}
