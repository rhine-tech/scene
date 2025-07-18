package authentication

import "github.com/rhine-tech/scene"

type AuthContext struct {
	UserID string
}

func NewAuthContext(userID string) AuthContext {
	return AuthContext{
		UserID: userID}
}

func GetAuthContext(ctx scene.Context) (AuthContext, bool) {
	return scene.ContextFindValue[AuthContext](ctx)
}

func SetAuthContext(ctx scene.Context, userID string) {
	scene.ContextSetValue[AuthContext](ctx, NewAuthContext(userID))
}

func (c *AuthContext) IsLogin() bool {
	return c.UserID != ""
}
