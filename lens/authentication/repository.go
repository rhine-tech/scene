package authentication

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/model"
)

type IAuthenticationRepository interface {
	scene.Named
	Authenticate(username string, password string) (userID string, err error)
	UserById(userId string) (User, error)
	UserByName(username string) (User, error)
	UserByEmail(email string) (User, error)

	AddUser(user User) (User, error)
	DeleteUser(userId string) error
	UpdateUser(user User) error
}

// IAccessTokenRepository 定义了 AccessToken 的持久化存储接口
type IAccessTokenRepository interface {
	scene.Named
	// CreateToken 创建并存储一个新的 AccessToken
	CreateToken(token AccessToken) (AccessToken, error)
	// GetTokenByValue 通过令牌字符串查找 AccessToken
	GetTokenByValue(token string) (AccessToken, error)
	// ListTokensByUser 分页列出某个用户的所有 AccessToken
	ListTokensByUser(userId string, offset, limit int64) (model.PaginationResult[AccessToken], error)
	// ListTokens 分页列出系统中的所有 AccessToken
	ListTokens(offset, limit int64) (model.PaginationResult[AccessToken], error)
	// DeleteToken 删除一个 AccessToken
	DeleteToken(token string) error
}
