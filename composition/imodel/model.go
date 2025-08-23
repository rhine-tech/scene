package imodel

import (
	"github.com/rhine-tech/scene/model"
	"github.com/rhine-tech/scene/model/query"
)

// IModel define a model with an id type
type IModel[T any] interface {
	GetID() T
}

// IModelService return a generic model
type IModelService[T any, Model IModel[T]] interface {
	Insert(models ...Model) (err error)

	DeleteById(id T) (err error)
	Delete(opts ...query.Option) (err error)

	GetById(id T) (model Model, err error)
	GetFirst(opts ...query.Option) (res Model, err error)
	List(limit, offset int64, opts ...query.Option) (result model.PaginationResult[Model], err error)
	Count(opts ...query.Option) (total int, err error)

	UpdateById(model Model) (err error)
	Update(updates map[query.Field]interface{}, opts ...query.Option) (err error)
}
