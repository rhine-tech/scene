package orm

import (
	"context"
	"github.com/rhine-tech/scene/model"
	"github.com/rhine-tech/scene/model/query"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Gorm interface {
	ORM
	query.QueryBuilder[*gorm.DB]
	DB() *gorm.DB
	RegisterModel(model ...any) error
	WithDB(db *gorm.DB) Gorm
}

func (g *GormRepository[Model]) Setup() error {
	return g.db.RegisterModel(new(Model))
}

type GormRepository[Model any] struct {
	db          Gorm `aperture:""`
	fieldMapper query.FieldMapper
}

func NewGormRepository[Model any](
	db Gorm, fieldMapper query.FieldMapper) *GormRepository[Model] {
	return &GormRepository[Model]{
		db:          db,
		fieldMapper: fieldMapper,
	}
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
	if err != nil {
		return result, err
	}
	err = qry.Offset(int(offset)).Limit(int(limit)).Find(&result.Results).Error
	if err != nil {
		return result, err
	}
	result.Offset = offset
	result.Count = int64(len(result.Results))
	return result, nil
}

func (g *GormRepository[Model]) WithTx(fn func(repo GenericRepository[Model]) error) error {
	return g.db.DB().Transaction(func(tx *gorm.DB) error {
		// 克隆 gorm 包装器并替换为事务上下文
		txDB := g.db.WithDB(tx)

		// 重新构造仓库（复用 fieldMapper）
		txRepo := &GormRepository[Model]{
			db:          txDB,
			fieldMapper: g.fieldMapper,
		}
		return fn(txRepo)
	})
}

func (g *GormRepository[Model]) WithContext(ctx context.Context) GenericRepository[Model] {
	return NewGormRepository[Model](g.db.WithDB(g.db.DB().WithContext(ctx)), g.fieldMapper)
}

func (g *GormRepository[Model]) Upsert(data *Model, conflictKeys []query.Field, updateKeys []query.Field) error {
	if len(conflictKeys) == 0 {
		return g.Create(data)
	}

	// 使用 fieldMapper 映射字段
	mappedConflictKeys := g.fieldMapper.Map(conflictKeys)

	db := g.db.DB()

	if len(updateKeys) == 0 {
		// updateKeys 为空 -> 使用 UpdateAll
		return db.Clauses(clause.OnConflict{
			Columns:   g.toClauseColumns(mappedConflictKeys),
			UpdateAll: true,
		}).Create(data).Error
	}

	mappedUpdateKeys := g.fieldMapper.Map(updateKeys)

	return db.Clauses(clause.OnConflict{
		Columns:   g.toClauseColumns(mappedConflictKeys),
		DoUpdates: clause.AssignmentColumns(mappedUpdateKeys),
	}).Create(data).Error
}
