package meta

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/model"
	"github.com/rhine-tech/scene/model/query"
)

type GormFileMetaRepository struct {
	internal *orm.GormRepository[storage.FileMeta]
	db       orm.Gorm `aperture:""`
}

func (r *GormFileMetaRepository) List(provider string, offset, limit int64) (model.PaginationResult[storage.FileMeta], error) {
	if provider == "" {
		return r.internal.List(offset, limit)
	}
	return r.internal.List(offset, limit, query.Field("provider").Equal(provider))
}

func (r *GormFileMetaRepository) Setup() error {
	r.internal = orm.NewGormRepository[storage.FileMeta](r.db, make(query.FieldMapper))
	return r.internal.Setup()
}

func (r *GormFileMetaRepository) ImplName() scene.ImplName {
	return storage.Lens.ImplNameNoVer("IFileMetaRepository")
}

func NewGormFileMetaRepository() storage.IFileMetaRepository {
	return &GormFileMetaRepository{}
}

func (r *GormFileMetaRepository) Store(meta storage.FileMeta) error {
	// Use upsert logic: if exists, update; otherwise insert
	// Try update first, fallback to create
	updateFields := map[string]interface{}{
		"original_filename": meta.OriginalFilename,
		"content_type":      meta.ContentType,
		"content_length":    meta.ContentLength,
		"md5_checksum":      meta.Md5Checksum,
		"finished":          meta.Finished,
		"updated_at":        meta.UpdatedAt,
	}

	found, err := r.Exists(meta.StorageKey)
	if err != nil {
		return err
	}
	if found {
		return r.internal.Update(updateFields, query.Field("storage_key").Equal(meta.StorageKey))
	}
	return r.internal.Create(&meta)
}

func (r *GormFileMetaRepository) Load(storageKey storage.StorageKey) (storage.FileMeta, error) {
	meta, found, err := r.internal.FindFirst(query.Field("storage_key").Equal(storageKey))
	if err != nil {
		return storage.FileMeta{}, err
	}
	if !found {
		return storage.FileMeta{}, storage.ErrMetaNotFound
	}
	return meta, nil
}

func (r *GormFileMetaRepository) Delete(storageKey storage.StorageKey) error {
	return r.internal.Delete(query.Field("storage_key").Equal(storageKey))
}

func (r *GormFileMetaRepository) Exists(storageKey storage.StorageKey) (bool, error) {
	count, err := r.internal.Count(query.Field("storage_key").Equal(storageKey))
	return count > 0, err
}
