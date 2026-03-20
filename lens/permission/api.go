package permission

import (
	"context"
)

func HasPermissionInCtx(ctx context.Context, perm *Permission) bool {
	pctx, ok := GetPermContext(ctx)
	if !ok {
		return false
	}
	if !pctx.HasPermission(perm) {
		return false
	}
	return true
}
