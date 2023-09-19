package drivers

import (
	"context"
	"github.com/rhine-tech/scene/model"
)

// TODO: finsh definition
type RepositoryDriver[Model any] interface {
	// Create
	Create(model ...*Model) error
	// Read
	Count() (count int64, err error)
	FindPage(param model.PageParam) (result []model.PageResult[*Model], err error)
	FindPagination(param model.PaginationParam) (result model.PaginationResult[*Model], err error)
	// Update
	// Delete
	Delete(...*Model) (err error)
	// Context support
	WithContext(ctx context.Context) RepositoryDriver[Model]
}
