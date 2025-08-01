package authentication

import "github.com/rhine-tech/scene"

// IsLoginInCtx return userId and if user has logged in
func IsLoginInCtx(ctx scene.Context) (string, bool) {
	actx, ok := GetAuthContext(ctx)
	if !ok {
		return "", false
	}
	return actx.UserID, actx.IsLogin()
}
