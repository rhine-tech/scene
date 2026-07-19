package authentication

import (
	"context"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/model"
)

type IAuthenticationRepository interface {
	scene.Named
	Authenticate(ctx context.Context, username string, password string) (userID string, err error)
	UserById(ctx context.Context, userId string) (User, error)
	UserByName(ctx context.Context, username string) (User, error)
	UserByEmail(ctx context.Context, email string) (User, error)

	AddUser(ctx context.Context, user User) (User, error)
	DeleteUser(ctx context.Context, userId string) error
	UpdateUser(ctx context.Context, user User) error
	ListUsers(ctx context.Context, offset, limit int64) (model.PaginationResult[User], error)
}

// IAccessTokenRepository 定义了 AccessToken 的持久化存储接口
type IAccessTokenRepository interface {
	scene.Named
	// CreateToken 创建并存储一个新的 AccessToken
	CreateToken(ctx context.Context, token AccessToken) (AccessToken, error)
	// GetTokenByValue 通过令牌字符串查找 AccessToken
	GetTokenByValue(ctx context.Context, token string) (AccessToken, error)
	// ListTokensByUser 分页列出某个用户的所有 AccessToken
	ListTokensByUser(ctx context.Context, userId string, offset, limit int64) (model.PaginationResult[AccessToken], error)
	// ListTokens 分页列出系统中的所有 AccessToken
	ListTokens(ctx context.Context, offset, limit int64) (model.PaginationResult[AccessToken], error)
	// DeleteToken 删除一个 AccessToken
	DeleteToken(ctx context.Context, token string) error
}
