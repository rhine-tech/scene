package meta

import (
	"context"
	"errors"
	"time"

	"github.com/rhine-tech/scene"
	sceneorm "github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type fileMetaRow struct {
	StorageKey       storage.StorageKey `gorm:"column:storage_key;primaryKey"`
	Provider         string             `gorm:"column:provider"`
	Identifier       string             `gorm:"column:identifier"`
	OriginalFilename string             `gorm:"column:original_filename"`
	ContentType      string             `gorm:"column:content_type"`
	ContentLength    int64              `gorm:"column:content_length"`
	MD5Checksum      string             `gorm:"column:md5_checksum"`
	Finished         bool               `gorm:"column:finished"`
	CreatedAt        time.Time          `gorm:"column:created_at"`
	UpdatedAt        time.Time          `gorm:"column:updated_at"`
}

// TableName preserves the table selected by GORM for the former FileMeta
// persistence model.
func (fileMetaRow) TableName() string {
	return "file_metas"
}

func fileMetaRowFromDomain(meta storage.FileMeta) fileMetaRow {
	return fileMetaRow{
		StorageKey:       meta.StorageKey,
		Provider:         meta.Provider,
		Identifier:       meta.Identifier,
		OriginalFilename: meta.OriginalFilename,
		ContentType:      meta.ContentType,
		ContentLength:    meta.ContentLength,
		MD5Checksum:      meta.Md5Checksum,
		Finished:         meta.Finished,
		CreatedAt:        meta.CreatedAt,
		UpdatedAt:        meta.UpdatedAt,
	}
}

func (r fileMetaRow) toDomain() storage.FileMeta {
	return storage.FileMeta{
		StorageKey:       r.StorageKey,
		Provider:         r.Provider,
		Identifier:       r.Identifier,
		OriginalFilename: r.OriginalFilename,
		ContentType:      r.ContentType,
		ContentLength:    r.ContentLength,
		Md5Checksum:      r.MD5Checksum,
		Finished:         r.Finished,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

type GormFileMetaRepository struct {
	db *sceneorm.Gorm `aperture:""`
}

func NewGormFileMetaRepository(db *sceneorm.Gorm) storage.IFileMetaRepository {
	return &GormFileMetaRepository{db: db}
}

func (r *GormFileMetaRepository) Setup() error {
	return r.db.AutoMigrate(&fileMetaRow{})
}

func (r *GormFileMetaRepository) TearDown() error {
	return nil
}

func (r *GormFileMetaRepository) ImplName() scene.ImplName {
	return storage.Lens.ImplName("IFileMetaRepository", "gorm")
}

func (r *GormFileMetaRepository) Store(ctx context.Context, meta storage.FileMeta) error {
	row := fileMetaRowFromDomain(meta)
	return r.db.Session(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "storage_key"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"original_filename",
				"content_type",
				"content_length",
				"md5_checksum",
				"finished",
				"updated_at",
			}),
		}).
		Create(&row).Error
}

func (r *GormFileMetaRepository) Load(
	ctx context.Context,
	storageKey storage.StorageKey,
) (storage.FileMeta, error) {
	var row fileMetaRow
	err := r.db.Session(ctx).
		Where(&fileMetaRow{StorageKey: storageKey}).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return storage.FileMeta{}, storage.ErrMetaNotFound
	}
	if err != nil {
		return storage.FileMeta{}, err
	}
	return row.toDomain(), nil
}

func (r *GormFileMetaRepository) Delete(
	ctx context.Context,
	storageKey storage.StorageKey,
) error {
	return r.db.Session(ctx).
		Where(&fileMetaRow{StorageKey: storageKey}).
		Delete(&fileMetaRow{}).Error
}

func (r *GormFileMetaRepository) List(
	ctx context.Context,
	provider string,
	offset, limit int64,
) (model.PaginationResult[storage.FileMeta], error) {
	result := model.PaginationResult[storage.FileMeta]{
		Offset:  offset,
		Results: make([]storage.FileMeta, 0),
	}

	countQuery := r.db.Session(ctx).Model(&fileMetaRow{})
	findQuery := r.db.Session(ctx)
	if provider != "" {
		where := &fileMetaRow{Provider: provider}
		countQuery = countQuery.Where(where)
		findQuery = findQuery.Where(where)
	}
	if err := countQuery.Count(&result.Total).Error; err != nil {
		return result, err
	}

	var rows []fileMetaRow
	if err := findQuery.
		Offset(int(offset)).
		Limit(int(limit)).
		Find(&rows).Error; err != nil {
		return result, err
	}

	result.Results = make([]storage.FileMeta, len(rows))
	for i, row := range rows {
		result.Results[i] = row.toDomain()
	}
	result.Count = int64(len(result.Results))
	return result, nil
}
