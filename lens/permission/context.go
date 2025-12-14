package permission

import "github.com/rhine-tech/scene"

type PermContext struct {
	Owner string
	srv   PermissionService
}

func NewPermContext(owner string, srv PermissionService) PermContext {
	return PermContext{Owner: owner, srv: srv}
}

func GetPermContext(ctx scene.Context) (PermContext, bool) {
	return scene.ContextFindValue[PermContext](ctx)
}

func SetPermContext(ctx scene.Context, owner string, srv PermissionService) {
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

func (c *PermContext) ListPermissions() []*Permission {
	if c.Owner == "" {
		return []*Permission{}
	}
	return c.srv.ListPermissions(string(c.Owner))
}
