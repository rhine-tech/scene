package permission

import "github.com/rhine-tech/scene"

type PermissionRepository interface {
	scene.Repository
	//GetOwners() []PermOwner
	GetPermissions(owner string) []*Permission
	AddPermission(owner string, perm string) (*Permission, error)
	RemovePermission(owner string, perm string) error
}
