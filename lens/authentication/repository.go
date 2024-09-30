package authentication

import "github.com/rhine-tech/scene"

type AuthenticationRepository interface {
	scene.Named
	Authenticate(username string, password string) (userID string, err error)
	UserById(userId string) (User, error)
	UserByName(username string) (User, error)
	UserByEmail(email string) (User, error)

	AddUser(username, password string) (User, error)
	DeleteUser(userId string) error
	UpdateUser(user User) error
}

type UserInfoRepository interface {
	scene.Named
	InfoById(userId string) (UserInfo, error)
}
