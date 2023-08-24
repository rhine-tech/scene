package service

import "github.com/aynakeya/scene/lens/middlewares/authentication"

func omitPassword(user authentication.User, err error) (authentication.User, error) {
	user.Password = ""
	return user, err
}
