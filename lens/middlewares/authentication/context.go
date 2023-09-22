package authentication

type AuthContext struct {
	UserID   string
	Username string
}

func (c *AuthContext) IsLogin() bool {
	return c.UserID != ""
}
