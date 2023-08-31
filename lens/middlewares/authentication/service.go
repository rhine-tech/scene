package authentication

import (
	"github.com/aynakeya/scene"
	"net/http"
)

type AuthenticationService interface {
	scene.Service
	Authenticate(username string, password string) (userID string, err error)
	UserById(userId string) (User, error)
	UserByName(username string) (User, error)
	UserByEmail(email string) (User, error)
}

type AuthenticationManageService interface {
	AuthenticationService
	AddUser(username, password string) (User, error)
	DeleteUser(userId string) error
	UpdateUser(user User) error
}

type LoginStatusService interface {
	scene.Service
	Verify(request *http.Request) (status LoginStatus, err error)
	Login(userId string, resp http.ResponseWriter) (status LoginStatus, err error)
	Logout(resp http.ResponseWriter) (err error)
}

type UserInfoService interface {
	scene.Service
	InfoById(userId string) (UserInfo, error)
}
