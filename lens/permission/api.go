package permission

import "github.com/rhine-tech/scene"

func HasPermissionInCtx(ctx scene.Context, perm *Permission) bool {
	pctx, ok := GetPermContext(ctx)
	if !ok {
		return false
	}
	if !pctx.HasPermission(perm) {
		return false
	}
	return true
}
