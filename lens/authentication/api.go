package authentication

import (
	"context"
)

// IsLoginInCtx return userId and if user has logged in
func IsLoginInCtx(ctx context.Context) (string, bool) {
	actx, ok := GetAuthContext(ctx)
	if !ok {
		return "", false
	}
	return actx.UserID, actx.IsLogin()
}
