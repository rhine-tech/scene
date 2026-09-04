package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rhine-tech/scene"
	sceneorm "github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type permissionRow struct {
	ID    int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Owner string `gorm:"column:owner;type:varchar(255);not null;uniqueIndex:idx_permission_owner_perm,priority:1;index:idx_permission_perm_owner,priority:2"`
	Perm  string `gorm:"column:perm;type:varchar(255);not null;uniqueIndex:idx_permission_owner_perm,priority:2;index:idx_permission_perm_owner,priority:1"`
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
		Where("owner = ?", owner).
		Order("perm ASC").
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

func (r *gormImpl) HasPermission(
	ctx context.Context,
	owner string,
	requested *permission.Permission,
) (bool, error) {
	ancestors := permissionAncestors(requested)
	if len(ancestors) == 0 {
		return false, nil
	}

	var row permissionRow
	err := r.db.Session(ctx).
		Select("id").
		Where("owner = ? AND perm IN ?", owner, ancestors).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *gormImpl) AddPermission(
	ctx context.Context,
	owner string,
	value *permission.Permission,
) error {
	row := permissionRow{Owner: owner, Perm: value.String()}
	return r.db.Session(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "owner"},
				{Name: "perm"},
			},
			DoNothing: true,
		}).
		Create(&row).Error
}

func (r *gormImpl) RemovePermission(
	ctx context.Context,
	owner string,
	value *permission.Permission,
) error {
	return r.db.Session(ctx).
		Where("owner = ? AND perm = ?", owner, value.String()).
		Delete(&permissionRow{}).Error
}

func (r *gormImpl) ReplacePermissions(
	ctx context.Context,
	owner string,
	prefix *permission.Permission,
	permissions []*permission.Permission,
) error {
	prefixValue, values, err := replacementValues(prefix, permissions)
	if err != nil {
		return err
	}

	return r.db.Transaction(ctx, func(tx *gorm.DB) error {
		if err := permissionSubtree(tx.Where("owner = ?", owner), prefixValue).
			Delete(&permissionRow{}).Error; err != nil {
			return err
		}
		if len(values) == 0 {
			return nil
		}

		rows := make([]permissionRow, len(values))
		for i, value := range values {
			rows[i] = permissionRow{Owner: owner, Perm: value}
		}
		return tx.Create(&rows).Error
	})
}

func (r *gormImpl) ListOwnersWithPermission(
	ctx context.Context,
	requested *permission.Permission,
	offset, limit int64,
) (model.PaginationResult[permission.OwnerPermissions], error) {
	ancestors := permissionAncestors(requested)
	if len(ancestors) == 0 {
		return model.PaginationResult[permission.OwnerPermissions]{
			Offset:  offset,
			Results: make([]permission.OwnerPermissions, 0),
		}, nil
	}
	return r.listOwners(ctx, offset, limit, func(query *gorm.DB) *gorm.DB {
		return query.Where("perm IN ?", ancestors)
	})
}

func (r *gormImpl) ListOwnersWithGrantsByPrefix(
	ctx context.Context,
	prefix *permission.Permission,
	offset, limit int64,
) (model.PaginationResult[permission.OwnerPermissions], error) {
	if prefix == nil {
		return model.PaginationResult[permission.OwnerPermissions]{
			Offset:  offset,
			Results: make([]permission.OwnerPermissions, 0),
		}, errors.New("permission prefix is required")
	}
	return r.listOwners(ctx, offset, limit, func(query *gorm.DB) *gorm.DB {
		return permissionSubtree(query, prefix.String())
	})
}

func (r *gormImpl) listOwners(
	ctx context.Context,
	offset, limit int64,
	filter func(*gorm.DB) *gorm.DB,
) (model.PaginationResult[permission.OwnerPermissions], error) {
	result := model.PaginationResult[permission.OwnerPermissions]{
		Offset:  offset,
		Results: make([]permission.OwnerPermissions, 0),
	}

	if err := filter(r.db.Session(ctx).Model(&permissionRow{})).
		Distinct("owner").
		Count(&result.Total).Error; err != nil {
		return result, err
	}

	var owners []string
	if err := filter(r.db.Session(ctx).Model(&permissionRow{})).
		Distinct("owner").
		Order("owner ASC").
		Offset(int(offset)).
		Limit(int(limit)).
		Pluck("owner", &owners).Error; err != nil {
		return result, err
	}
	if len(owners) == 0 {
		return result, nil
	}

	result.Results = make([]permission.OwnerPermissions, len(owners))
	ownerIndexes := make(map[string]int, len(owners))
	for i, owner := range owners {
		result.Results[i] = permission.OwnerPermissions{
			Owner:       owner,
			Permissions: make([]*permission.Permission, 0),
		}
		ownerIndexes[owner] = i
	}

	var rows []permissionRow
	if err := filter(r.db.Session(ctx).Where("owner IN ?", owners)).
		Order("owner ASC").
		Order("perm ASC").
		Find(&rows).Error; err != nil {
		return result, err
	}
	for _, row := range rows {
		parsed, err := permission.ParsePermission(row.Perm)
		if err != nil {
			return result, err
		}
		index := ownerIndexes[row.Owner]
		result.Results[index].Permissions = append(result.Results[index].Permissions, parsed)
	}
	result.Count = int64(len(result.Results))
	return result, nil
}

func (r *gormImpl) RemovePermissionsByPrefix(
	ctx context.Context,
	prefix *permission.Permission,
) error {
	if prefix == nil {
		return errors.New("permission prefix is required")
	}
	return permissionSubtree(r.db.Session(ctx), prefix.String()).
		Delete(&permissionRow{}).Error
}

func permissionAncestors(value *permission.Permission) []string {
	parts := value.Parts()
	if len(parts) == 0 {
		return nil
	}

	ancestors := make([]string, len(parts))
	ancestors[0] = parts[0]
	for i := 1; i < len(parts); i++ {
		ancestors[i] = ancestors[i-1] + ":" + parts[i]
	}
	return ancestors
}

func replacementValues(
	prefix *permission.Permission,
	permissions []*permission.Permission,
) (string, []string, error) {
	if prefix == nil {
		return "", nil, errors.New("permission prefix is required")
	}

	values := make([]string, 0, len(permissions))
	seen := make(map[string]struct{}, len(permissions))
	for _, candidate := range permissions {
		if candidate == nil || !prefix.HasPermission(candidate) {
			return "", nil, fmt.Errorf("permission %q is outside prefix %q", candidate, prefix)
		}
		value := candidate.String()
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return prefix.String(), values, nil
}

func permissionSubtree(query *gorm.DB, prefix string) *gorm.DB {
	escaped := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(prefix + ":")
	return query.Where("perm = ? OR perm LIKE ? ESCAPE '!'", prefix, escaped+"%")
}
