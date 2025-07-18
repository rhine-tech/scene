package orm

import (
	"context"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/model"
	"github.com/rhine-tech/scene/model/query"
)

const Lens scene.CompositionName = "orm"

type ORM interface {
	OrmName() scene.ImplName
}

type GenericRepository[Model any] interface {
	Create(data *Model) error
	Update(updates map[string]interface{}, options ...query.Option) error
	Upsert(data *Model, conflictKeys []query.Field, updateKeys []query.Field) error
	Delete(options ...query.Option) error
	FindFirst(options ...query.Option) (data Model, found bool, err error)
	Count(options ...query.Option) (count int64, err error)
	List(offset, limit int64, options ...query.Option) (model.PaginationResult[Model], error)
	// WithTx Transaction 事务支持
	WithTx(fn func(repo GenericRepository[Model]) error) error
	WithContext(ctx context.Context) GenericRepository[Model]
}
