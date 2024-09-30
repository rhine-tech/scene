package service

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/authentication"
)

type authenticationManageService struct {
	logger logger.ILogger                          `aperture:""`
	repo   authentication.AuthenticationRepository `aperture:""`
}

func (a *authenticationManageService) Setup() error {
	a.logger = a.logger.WithPrefix(a.SrvImplName().Identifier())
	//if err := a.repo.Status(); err != nil {
	//	a.logger.Errorf("repo init failed: %v", err)
	//	return err
	//}
	a.logger.Info("setup success")
	return nil
}

func (a *authenticationManageService) SrvImplName() scene.ImplName {
	return authentication.Lens.ImplName("AuthenticationManageService", "default")
}

func NewAuthenticationService(
	logger logger.ILogger, repo authentication.AuthenticationRepository) authentication.AuthenticationManageService {
	s := &authenticationManageService{
		logger: logger,
		repo:   repo,
	}
	return s
}

func (a *authenticationManageService) UserById(userId string) (authentication.User, error) {
	return omitPassword(a.repo.UserById(userId))
}

func (a *authenticationManageService) UserByName(username string) (authentication.User, error) {
	return omitPassword(a.repo.UserByName(username))
}

func (a *authenticationManageService) UserByEmail(email string) (authentication.User, error) {
	return omitPassword(a.UserByEmail(email))
}

func (a *authenticationManageService) Authenticate(username string, password string) (userID string, err error) {
	uid, err := a.repo.Authenticate(username, password)
	if err != nil {
		a.logger.Warnf("failed to authenticate user %s with password %s: %v", username, password, err)
		return "", authentication.ErrAuthenticationFailed
	}
	return uid, nil
}

func (a *authenticationManageService) AddUser(username, password string) (authentication.User, error) {
	return a.repo.AddUser(username, password)
}

func (a *authenticationManageService) DeleteUser(userId string) error {
	return a.repo.DeleteUser(userId)
}

func (a *authenticationManageService) UpdateUser(user authentication.User) error {
	return a.repo.UpdateUser(user)
}
