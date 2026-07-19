package meta

import (
	"context"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/model"
)

type dummyImpl struct {
}

func (d dummyImpl) List(
	_ context.Context,
	_ string,
	offset, _ int64,
) (model.PaginationResult[storage.FileMeta], error) {
	return model.PaginationResult[storage.FileMeta]{
		Offset:  offset,
		Results: []storage.FileMeta{},
	}, storage.ErrLoadingMeta
}

func NewDummyImpl() storage.IFileMetaRepository {
	return &dummyImpl{}
}

func (d dummyImpl) ImplName() scene.ImplName {
	return storage.Lens.ImplName("IFileMetaRepository", "dummy")
}

func (d dummyImpl) Store(context.Context, storage.FileMeta) error {
	return storage.ErrLoadingMeta
}

func (d dummyImpl) Load(context.Context, storage.StorageKey) (meta storage.FileMeta, err error) {
	return meta, storage.ErrLoadingMeta
}

func (d dummyImpl) Delete(context.Context, storage.StorageKey) error {
	return storage.ErrLoadingMeta
}
