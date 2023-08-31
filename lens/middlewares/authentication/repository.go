package authentication

import "github.com/aynakeya/scene"

type AuthenticationRepository interface {
	scene.Repository
	Authenticate(username string, password string) (userID string, err error)
	UserById(userId string) (User, error)
	UserByName(username string) (User, error)
	UserByEmail(email string) (User, error)
}

type AuthenticationManageRepository interface {
	AuthenticationRepository
	AddUser(username, password string) (User, error)
	DeleteUser(userId string) error
	UpdateUser(user User) error
}

type UserInfoRepository interface {
	scene.Repository
	InfoById(userId string) (UserInfo, error)
}
