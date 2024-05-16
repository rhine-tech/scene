package repository

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/lens/permission"
)

type tablePermission struct {
	ID    int64  `gorm:"column:id;primary_key;auto_increment"`
	Owner string `gorm:"column:owner;type:varchar(255);not null"`
	Perm  string `gorm:"column:perm;type:varchar(255);not null"`
}

func (tablePermission) TableName() string {
	return permission.Lens.TableName("permissions")
}

var _ = permission.PermissionRepository(&gormImpl{})

type gormImpl struct {
	gorm orm.Gorm `aperture:""`
}

func NewGormImpl(gorm orm.Gorm) permission.PermissionRepository {
	return &gormImpl{gorm: gorm}
}

func (m *gormImpl) Setup() error {
	err := m.gorm.RegisterModel(new(tablePermission))
	if err != nil {
		return err
	}
	return nil
}

func (g *gormImpl) RepoImplName() scene.ImplName {
	return permission.Lens.ImplName("PermissionRepository", "gorm")
}

func (g *gormImpl) GetPermissions(owner string) []*permission.Permission {
	perms := make([]*permission.Permission, 0)
	permResult := make([]tablePermission, 0)
	g.gorm.DB().Where("owner = ?", owner).Find(&permResult)
	for _, perm := range permResult {
		perms = append(perms, permission.MustParsePermission(perm.Perm))
	}
	return perms
}

func (g *gormImpl) AddPermission(owner string, perm string) (*permission.Permission, error) {
	// check if permission exists
	var count int64
	g.gorm.DB().Model(&tablePermission{}).Where("owner = ? AND perm = ?", owner, perm).Count(&count)
	if count > 0 {
		return permission.MustParsePermission(perm), nil
	}
	err := g.gorm.DB().Create(&tablePermission{Owner: owner, Perm: perm}).Error
	if err != nil {
		return nil, err
	}
	return permission.MustParsePermission(perm), nil
}

func (g gormImpl) RemovePermission(owner string, perm string) error {
	return g.gorm.DB().Delete(&tablePermission{}, "owner = ? AND perm = ?", owner, perm).Error
}
