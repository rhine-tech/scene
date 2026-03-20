package authentication

import (
	"context"
	"github.com/rhine-tech/scene"
)

type AuthContext struct {
	UserID string
}

type authContext struct{}

var authContextKey = authContext{}

func NewAuthContext(userID string) AuthContext {
	return AuthContext{
		UserID: userID}
}

func GetAuthContext(ctx context.Context) (AuthContext, bool) {
	return scene.ContextFindValue[AuthContext](ctx, authContextKey)
}

func SetAuthContext(ctx context.Context, userID string) context.Context {
	return scene.ContextSetValue[AuthContext](ctx, authContextKey, NewAuthContext(userID))
}

func (c *AuthContext) IsLogin() bool {
	return c.UserID != ""
}
