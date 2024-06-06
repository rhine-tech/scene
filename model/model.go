package model

import (
	"github.com/rhine-tech/scene/model/query"
)

// IModel
type IModel[T any] interface {
	GetID() T
}

type IModelService[T any] interface {
	Insert(models ...IModel[T]) (err error)

	GetById(id T) (model IModel[T], err error)
	Get(opts ...query.Option) (res IModel[T], err error)
	List(limit, offset int64, opts ...query.Option) (result PaginationResult[IModel[T]], err error)

	//UpdateById(id primitive.ObjectID, update bson.M, args ...interface{}) (err error)
	//Update(query bson.M, update bson.M, fields []string, args ...interface{}) (err error)
	//UpdateDoc(query bson.M, doc Model, fields []string, args ...interface{}) (err error)

	DeleteById(id T) (err error)
	Delete(opts ...query.Option) (err error)

	Count(opts ...query.Option) (total int, err error)
}
