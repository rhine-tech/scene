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

//func UseGormRepository[Model any](
//	db Gorm,
//	fieldMapper query.FieldMapper) (GenericRepository[Model], error) {
//	val := &GormRepository[Model]{
//		db:          registry.Use(db),
//		fieldMapper: fieldMapper,
//	}
//	err := val.Setup()
//	return val, err
//}

func NewGormRepository[Model any](
	db Gorm, fieldMapper query.FieldMapper) *GormRepository[Model] {
	return &GormRepository[Model]{
		db:          db,
		fieldMapper: fieldMapper,
	}
}

func (g *GormRepository[Model]) Setup() error {
	return g.db.RegisterModel(new(Model))
}

type GormRepository[Model any] struct {
	db          Gorm `aperture:""`
	fieldMapper query.FieldMapper
}

func (g *GormRepository[Model]) Create(data *Model) error {
	return g.db.DB().Create(data).Error
}

func (g *GormRepository[Model]) Update(updates map[string]interface{}, options ...query.Option) error {
	db := g.db.WithFieldMapper(g.fieldMapper).Build(options...)
	if db.Error != nil {
		return db.Error
	}
	// Using UpdateColumns instead of Save to update the fields specified by the options
	return db.Model(new(Model)).Updates(updates).Error
}

func (g *GormRepository[Model]) Delete(options ...query.Option) error {
	db := g.db.WithFieldMapper(g.fieldMapper).Build(options...)
	if db.Error != nil {
		return db.Error
	}
	return db.Delete(new(Model)).Error
}

func (g *GormRepository[Model]) FindFirst(options ...query.Option) (data Model, found bool, err error) {
	err = g.db.WithFieldMapper(g.fieldMapper).Build(options...).First(&data).Error
	if err != nil {
		//if errors.Is(err,gorm.ErrRecordNotFound) {
		//	return data,false, err
		//}
		return data, false, err
	}
	return data, true, nil
}

func (g *GormRepository[Model]) Count(options ...query.Option) (count int64, err error) {
	err = g.db.WithFieldMapper(g.fieldMapper).Build(options...).Model(new(Model)).Count(&count).Error
	return count, err
}

func (g *GormRepository[Model]) List(offset, limit int64, options ...query.Option) (model.PaginationResult[Model], error) {
	var result = model.PaginationResult[Model]{
		Results: make([]Model, 0),
	}
	qry := g.db.WithFieldMapper(g.fieldMapper).Build(options...).Session(&gorm.Session{})
	err := qry.Model(new(Model)).Count(&result.Total).Error
	err = qry.Offset(int(offset)).Limit(int(limit)).Find(&result.Results).Error
	if err != nil {
		return model.PaginationResult[Model]{}, err
	}
	result.Offset = offset
	result.Count = int64(len(result.Results))
	return result, nil
}
