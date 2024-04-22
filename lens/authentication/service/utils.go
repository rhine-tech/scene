package service

import "github.com/rhine-tech/scene/lens/middlewares/authentication"

func omitPassword(user authentication.User, err error) (authentication.User, error) {
	user.Password = ""
	return user, err
}
