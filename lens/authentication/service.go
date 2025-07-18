package authentication

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/model"
	"net/http"
)

type IAuthenticationService interface {
	scene.Service
	AddUser(username, password string) (User, error)
	DeleteUser(userId string) error
	UpdateUser(user User) error
	Authenticate(username string, password string) (userID string, err error)
	AuthenticateByToken(token string) (userID string, err error)
	UserById(userId string) (User, error)
	UserByName(username string) (User, error)
	UserByEmail(email string) (User, error)
}

type IAccessTokenService interface {
	scene.Service
	scene.WithContext[IAccessTokenService]
	// Create a new token for user
	Create(userId, name string, expireAt int64) (AccessToken, error)
	// ListByUser 分页列出某个用户的所有 AccessToken
	ListByUser(userId string, offset, limit int64) (model.PaginationResult[AccessToken], error)
	// List 分页列出系统中的所有 AccessToken
	List(offset, limit int64) (model.PaginationResult[AccessToken], error)
	// Validate token and return user ID
	Validate(token string) (userId string, valid bool, err error)
}

type HTTPLoginStatusVerifier interface {
	scene.Service
	Verify(request *http.Request) (status LoginStatus, err error)
	Login(userId string, resp http.ResponseWriter) (status LoginStatus, err error)
	Logout(resp http.ResponseWriter) (err error)
}

//type IExternalAccountService interface{}
