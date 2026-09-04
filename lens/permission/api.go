package permission

import (
	"context"
)

func HasPermissionInCtx(ctx context.Context, perm *Permission) (bool, error) {
	pctx, ok := GetPermContext(ctx)
	if !ok {
		return false, nil
	}
	return pctx.HasPermission(ctx, perm)
}
