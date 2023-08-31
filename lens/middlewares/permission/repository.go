package permission

import "github.com/aynakeya/scene"

type PermissionRepository interface {
	scene.Repository
	//GetOwners() []PermOwner
	GetPermissions(owner string) []*Permission
	AddPermission(owner string, perm string) (*Permission, error)
	RemovePermission(owner string, perm string) error
}
