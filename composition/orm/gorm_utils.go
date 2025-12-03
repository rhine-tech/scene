package orm

import (
	gormSqlite "github.com/glebarez/sqlite"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (g *GormRepository[Model]) toClauseColumns(keys []string) []clause.Column {
	cols := make([]clause.Column, len(keys))
	for i, key := range keys {
		cols[i] = clause.Column{Name: key}
	}
	return cols
}

func GormDialectorMySql(ds datasource.MysqlDataSource) gorm.Dialector {
	return gormMysql.New(gormMysql.Config{
		Conn: ds.Connection(),
	})
}

func GormDialectorSqlite(ds datasource.SqliteDataSource) gorm.Dialector {
	return &gormSqlite.Dialector{
		Conn: ds.Connection(),
	}
}
