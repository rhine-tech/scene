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
	return r.internal.List(offset, limit)
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

	found, err := r.Exists(meta.FileID)
	if err != nil {
		return err
	}
	if found {
		return r.internal.Update(updateFields, query.Field("file_id").Equal(meta.FileID))
	}
	return r.internal.Create(&meta)
}

func (r *GormFileMetaRepository) Load(fileId storage.FileID) (storage.FileMeta, error) {
	meta, found, err := r.internal.FindFirst(query.Field("file_id").Equal(fileId))
	if err != nil {
		return storage.FileMeta{}, err
	}
	if !found {
		return storage.FileMeta{}, storage.ErrMetaNotFound
	}
	return meta, nil
}

func (r *GormFileMetaRepository) Delete(fileId storage.FileID) error {
	return r.internal.Delete(query.Field("file_id").Equal(fileId))
}

func (r *GormFileMetaRepository) Exists(fileId storage.FileID) (bool, error) {
	count, err := r.internal.Count(query.Field("file_id").Equal(fileId))
	return count > 0, err
}
