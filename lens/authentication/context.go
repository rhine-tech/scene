package authentication

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

func (c *AuthContext) IsLogin() bool {
	return c.UserID != ""
}

func (c *AuthContext) UserInfo() (UserInfo, error) {
	if c.srv == nil || !c.IsLogin() {
		return UserInfo{}, ErrNotLogin
	}
	return c.srv.InfoById(c.UserID)
}
