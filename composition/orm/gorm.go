package orm

import (
	"github.com/rhine-tech/scene/model/query"
	"gorm.io/gorm"
)

type Gorm interface {
	ORM
	query.QueryBuilder[*gorm.DB]
	DB() *gorm.DB
	RegisterModel(model ...any) error
}
