package orm

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/model"
	"github.com/rhine-tech/scene/model/query"
)

const Lens scene.CompositionName = "orm"

type ORM interface {
	OrmName() scene.ImplName
}

type GenericRepository[Model any] interface {
	//Create(model ...*Model) error
	Count(options ...query.Option) (count int64, err error)
	List(offset, limit int64, options ...query.Option) (model.PaginationResult[Model], error)
	//Delete(...*Model) (err error)
	//WithContext(ctx context.Context) RepositoryDriver[Model]
	//FindOne(options ...query.Option)
}
