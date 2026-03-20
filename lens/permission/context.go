package permission

import (
	"context"

	"github.com/rhine-tech/scene"
)

type PermContext struct {
	Owner string
	srv   PermissionService
}

type permContext struct{}

var permissionContextKey = permContext{}

func NewPermContext(owner string, srv PermissionService) PermContext {
	return PermContext{Owner: owner, srv: srv}
}

func GetPermContext(ctx context.Context) (PermContext, bool) {
	return scene.ContextFindValue[PermContext](ctx, permissionContextKey)
}

func SetPermContext(ctx context.Context, owner string, srv PermissionService) context.Context {
	return scene.ContextSetValue[PermContext](ctx, permissionContextKey, NewPermContext(owner, srv))
}

func (c *PermContext) HasPermission(perm *Permission) bool {
	if c.Owner == "" {
		return false
	}
	return c.srv.HasPermission(c.Owner, perm)
}

func (c *PermContext) HasPermissionStr(perm string) bool {
	if c.Owner == "" {
		return false
	}
	return c.srv.HasPermissionStr(c.Owner, perm)
}

func (c *PermContext) ListPermissions() []*Permission {
	if c.Owner == "" {
		return []*Permission{}
	}
	return c.srv.ListPermissions(c.Owner)
}
