package permission

import "github.com/rhine-tech/scene"

const Lens scene.ModuleName = "permission"

type PermissionService interface {
	scene.Service
	HasPermission(owner string, perm *Permission) bool
	HasPermissionStr(owner string, perm string) bool
	//ListOwners() []string
	ListPermissions(owner string) PermissionSet
	AddPermission(owner string, perm string) error
	RemovePermission(owner string, perm string) error
}

type PermissionRepository interface {
	scene.Repository
	//GetOwners() []PermOwner
	GetPermissions(owner string) []*Permission
	AddPermission(owner string, perm string) (*Permission, error)
	RemovePermission(owner string, perm string) error
}
