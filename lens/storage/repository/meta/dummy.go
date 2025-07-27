package meta

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/model"
)

type dummyImpl struct {
}

func (d dummyImpl) List(provider string, offset, limit int64) (model.PaginationResult[storage.FileMeta], error) {
	//TODO implement me
	panic("implement me")
}

func NewDummyImpl() storage.IFileMetaRepository {
	return &dummyImpl{}
}

func (d dummyImpl) ImplName() scene.ImplName {
	return storage.Lens.ImplName("IFileMetaRepository", "dummy")
}

func (d dummyImpl) Store(meta storage.FileMeta) error {
	return storage.ErrLoadingMeta
}

func (d dummyImpl) Load(fileId storage.FileID) (meta storage.FileMeta, err error) {
	return meta, storage.ErrLoadingMeta
}

func (d dummyImpl) Delete(fileId storage.FileID) error {
	return storage.ErrLoadingMeta
}
