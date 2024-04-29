package orm

import (
	"github.com/rhine-tech/scene/model/filter"
	"gorm.io/gorm"
)

type Gorm interface {
	ORM
	filter.FilterBuilder[*gorm.DB]
	DB() *gorm.DB
	RegisterModel(model ...any) error
}
