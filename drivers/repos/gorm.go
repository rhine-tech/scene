package repos

import (
	"github.com/rhine-tech/scene/lens/infrastructure/datasource"
	"github.com/rhine-tech/scene/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type GormRepo[T any] struct {
	*gorm.DB
}

func UseGormMysql[T any](ds datasource.MysqlDataSource) *GormRepo[T] {
	gormDb, _ := gorm.Open(mysql.New(mysql.Config{
		Conn: ds.Connection(),
	}), &gorm.Config{})
	return &GormRepo[T]{DB: gormDb}
}

func (m *GormRepo[T]) FindPagination(scope func(db *gorm.DB) *gorm.DB, offset int, limit int) (result model.PaginationResult[T], err error) {
	var results []T
	cur := m.DB.Model(new(T)).Scopes(scope).Offset(offset).Limit(limit).Find(&results)
	if cur.Error != nil {
		return result, cur.Error
	}
	result.Results = results
	var cnt int64
	cur = m.DB.Model(new(T)).Count(&cnt)
	if cur.Error != nil {
		return result, cur.Error
	}
	result.Total = int(cnt)
	result.Offset = offset
	return result, nil
}

//func UseGormSqlite(ds datasource.SqliteDataSource) *GormRepo {
//	gormDb, _ := gorm.Open(sqlite.New(mysql.Config{
//		Conn: ds.Connection(),
//	}), &gorm.Config{})
//	return &GormRepo{DB: gormDb}
//}
