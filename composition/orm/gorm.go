package orm

import (
	"github.com/rhine-tech/scene/model"
	"github.com/rhine-tech/scene/model/query"
	"gorm.io/gorm"
)

type Gorm interface {
	ORM
	query.QueryBuilder[*gorm.DB]
	DB() *gorm.DB
	RegisterModel(model ...any) error
}

type GormRepository[Model any] struct {
	db          Gorm
	fieldMapper query.FieldMapper
}

func (g *GormRepository[Model]) Setup() error {
	return g.db.RegisterModel(new(Model))
}

func (g *GormRepository[Model]) Count(options ...query.Option) (count int64, err error) {
	err = g.db.WithFieldMapper(g.fieldMapper).Build(options...).Count(&count).Error
	return count, err
}

func (g *GormRepository[Model]) List(offset, limit int64, options ...query.Option) (model.PaginationResult[Model], error) {
	var result = model.PaginationResult[Model]{
		Results: make([]Model, 0),
	}
	qry := g.db.WithFieldMapper(g.fieldMapper).Build(options...).Session(&gorm.Session{})
	err := qry.Offset(int(offset)).Limit(int(limit)).Find(&result.Results).Error
	if err != nil {
		return model.PaginationResult[Model]{}, err
	}
	qry.Count(&result.Total)
	result.Offset = offset
	result.Count = int64(len(result.Results))
	return result, nil
}
