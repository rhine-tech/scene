package permission

import (
	"context"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/model"
)

const Lens scene.ModuleName = "permission"

type PermissionService interface {
	scene.Named
	HasPermission(ctx context.Context, owner string, perm *Permission) (bool, error)
	// HasPermissions checks each permission independently, preserving input order
	// and duplicates. Nil permissions are denied; errors invalidate all results.
	HasPermissions(ctx context.Context, owner string, permissions []*Permission) ([]bool, error)
	// HasPermissionsForOwners checks each permission against the union of owners' grants,
	// including grants inherited from ancestors. It follows HasPermissions's result
	// contract. Empty owners deny every permission; duplicate or unknown owners have no effect.
	HasPermissionsForOwners(ctx context.Context, owners []string, permissions []*Permission) ([]bool, error)
	HasPermissionStr(ctx context.Context, owner string, perm string) (bool, error)
	ListPermissions(ctx context.Context, owner string) ([]*Permission, error)
	// ListPermissionsForOwners returns the deduplicated union of explicit grants,
	// sorted by permission string. It neither expands descendants nor removes grants
	// covered by ancestors. Empty owners return an empty list; errors invalidate all results.
	ListPermissionsForOwners(ctx context.Context, owners []string) ([]*Permission, error)
	// ListExplicitGrantsByPrefix lists an owner's explicit grants at prefix and
	// below it. Grants inherited from ancestors of prefix are not included.
	ListExplicitGrantsByPrefix(ctx context.Context, owner string, prefix *Permission) ([]*Permission, error)
	AddPermission(ctx context.Context, owner string, perm string) error
	RemovePermission(ctx context.Context, owner string, perm string) error
	// ReplacePermissions atomically replaces an owner's explicit grants at prefix and below it.
	ReplacePermissions(ctx context.Context, owner string, prefix *Permission, permissions []*Permission) error
	// ListOwnersWithPermission lists owners whose explicit grants include perm or one of its ancestors.
	ListOwnersWithPermission(ctx context.Context, perm *Permission, offset, limit int64) (model.PaginationResult[OwnerPermissions], error)
	// ListOwnersWithGrantsByPrefix lists owners with explicit grants at prefix or below it.
	// Each result contains only the explicit grants matched by that prefix.
	ListOwnersWithGrantsByPrefix(ctx context.Context, prefix *Permission, offset, limit int64) (model.PaginationResult[OwnerPermissions], error)
	// RemovePermissionsByPrefix removes explicit grants at prefix and below it for every owner.
	RemovePermissionsByPrefix(ctx context.Context, prefix *Permission) error
}

type PermissionRepository interface {
	scene.Named
	// HasPermissions follows PermissionService.HasPermissionsForOwners's result contract.
	HasPermissions(ctx context.Context, owners []string, permissions ...*Permission) ([]bool, error)
	// GetPermissions follows PermissionService.ListPermissionsForOwners's result contract.
	GetPermissions(ctx context.Context, owners []string) ([]*Permission, error)
	ListExplicitGrantsByPrefix(ctx context.Context, owner string, prefix *Permission) ([]*Permission, error)
	AddPermission(ctx context.Context, owner string, perm *Permission) error
	RemovePermission(ctx context.Context, owner string, perm *Permission) error
	ReplacePermissions(ctx context.Context, owner string, prefix *Permission, permissions []*Permission) error
	ListOwnersWithPermission(ctx context.Context, perm *Permission, offset, limit int64) (model.PaginationResult[OwnerPermissions], error)
	ListOwnersWithGrantsByPrefix(ctx context.Context, prefix *Permission, offset, limit int64) (model.PaginationResult[OwnerPermissions], error)
	RemovePermissionsByPrefix(ctx context.Context, prefix *Permission) error
}
