package permission

import "github.com/aynakeya/scene"

type PermissionService interface {
	scene.Service
	HasPermission(owner string, perm string) bool
	//ListOwners() []string
	ListPermissions(owner string) PermissionSet
	AddPermission(owner string, perm string) error
	RemovePermission(owner string, perm string) error
}
