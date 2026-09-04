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

func (c *PermContext) HasPermission(ctx context.Context, perm *Permission) (bool, error) {
	if c.Owner == "" {
		return false, nil
	}
	return c.srv.HasPermission(ctx, c.Owner, perm)
}

func (c *PermContext) HasPermissionStr(ctx context.Context, perm string) (bool, error) {
	if c.Owner == "" {
		return false, nil
	}
	return c.srv.HasPermissionStr(ctx, c.Owner, perm)
}

func (c *PermContext) ListPermissions(ctx context.Context) ([]*Permission, error) {
	if c.Owner == "" {
		return []*Permission{}, nil
	}
	return c.srv.ListPermissions(ctx, c.Owner)
}
