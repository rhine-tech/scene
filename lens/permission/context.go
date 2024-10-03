package permission

import "github.com/rhine-tech/scene"

type PermContext struct {
	Owner PermOwner
	srv   PermissionService
}

func NewPermContext(owner PermOwner, srv PermissionService) PermContext {
	return PermContext{Owner: owner, srv: srv}
}

func GetPermContext(ctx scene.Context) (PermContext, bool) {
	return scene.ContextFindValue[PermContext](ctx)
}

func SetPermContext(ctx scene.Context, owner PermOwner, srv PermissionService) {
	scene.ContextSetValue[PermContext](ctx, NewPermContext(owner, srv))
}

func (c *PermContext) HasPermission(perm *Permission) bool {
	if c.Owner == "" {
		return false
	}
	return c.srv.HasPermission(string(c.Owner), perm)
}

func (c *PermContext) HasPermissionStr(perm string) bool {
	if c.Owner == "" {
		return false
	}
	return c.srv.HasPermissionStr(string(c.Owner), perm)
}

func (c *PermContext) ListPermissions() PermissionSet {
	if c.Owner == "" {
		return PermissionSet{}
	}
	return c.srv.ListPermissions(string(c.Owner))
}
