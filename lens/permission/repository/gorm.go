package repository

import (
	"context"

	"github.com/rhine-tech/scene"
	sceneorm "github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/lens/permission"
	"gorm.io/gorm/clause"
)

type permissionRow struct {
	ID    int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Owner string `gorm:"column:owner;type:varchar(255);not null;uniqueIndex:idx_permission_owner_perm,priority:1"`
	Perm  string `gorm:"column:perm;type:varchar(255);not null;uniqueIndex:idx_permission_owner_perm,priority:2"`
}

func (permissionRow) TableName() string {
	return permission.Lens.TableName("permissions")
}

var _ permission.PermissionRepository = (*gormImpl)(nil)

type gormImpl struct {
	db *sceneorm.Gorm `aperture:""`
}

func NewGormImpl(db *sceneorm.Gorm) permission.PermissionRepository {
	return &gormImpl{db: db}
}

func (r *gormImpl) Setup() error {
	return r.db.AutoMigrate(&permissionRow{})
}

func (r *gormImpl) TearDown() error {
	return nil
}

func (r *gormImpl) ImplName() scene.ImplName {
	return permission.Lens.ImplName("PermissionRepository", "gorm")
}

func (r *gormImpl) GetPermissions(
	ctx context.Context,
	owner string,
) ([]*permission.Permission, error) {
	var rows []permissionRow
	if err := r.db.Session(ctx).
		Where(&permissionRow{Owner: owner}).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	permissions := make([]*permission.Permission, len(rows))
	for i, row := range rows {
		parsed, err := permission.ParsePermission(row.Perm)
		if err != nil {
			return nil, err
		}
		permissions[i] = parsed
	}
	return permissions, nil
}

func (r *gormImpl) AddPermission(
	ctx context.Context,
	owner string,
	value string,
) (*permission.Permission, error) {
	parsed, err := permission.ParsePermission(value)
	if err != nil {
		return nil, err
	}

	row := permissionRow{Owner: owner, Perm: value}
	err = r.db.Session(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "owner"},
				{Name: "perm"},
			},
			DoNothing: true,
		}).
		Create(&row).Error
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

func (r *gormImpl) RemovePermission(
	ctx context.Context,
	owner string,
	value string,
) error {
	return r.db.Session(ctx).
		Where(&permissionRow{Owner: owner, Perm: value}).
		Delete(&permissionRow{}).Error
}
