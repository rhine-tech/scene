package authentication

import "github.com/rhine-tech/scene"

type AuthContext struct {
	UserID string
	srv    UserInfoService
}

func NewAuthContext(
	userID string,
	srv UserInfoService) AuthContext {
	return AuthContext{
		UserID: userID, srv: srv}
}

func GetAuthContext(ctx scene.Context) (AuthContext, bool) {
	return scene.ContextFindValue[AuthContext](ctx)
}

func SetAuthContext(ctx scene.Context, userID string, srv UserInfoService) {
	scene.ContextSetValue[AuthContext](ctx, NewAuthContext(userID, srv))
}

func (c *AuthContext) IsLogin() bool {
	return c.UserID != ""
}

func (c *AuthContext) UserInfo() (UserInfo, error) {
	if c.srv == nil || !c.IsLogin() {
		return UserInfo{}, ErrNotLogin
	}
	return c.srv.InfoById(c.UserID)
}
