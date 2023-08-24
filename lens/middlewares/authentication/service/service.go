package service

import (
	"github.com/aynakeya/scene/lens/infrastructure/logger"
	"github.com/aynakeya/scene/lens/middlewares/authentication"
)

type authenticationManageService struct {
	logger logger.ILogger
	repo   authentication.AuthenticationManageRepository
}

func NewAuthenticationService(
	logger logger.ILogger, repo authentication.AuthenticationManageRepository) authentication.AuthenticationManageService {
	s := &authenticationManageService{
		logger: logger.WithPrefix("Service.authentication"),
		repo:   repo,
	}
	if err := s.repo.Status(); err != nil {
		s.logger.Errorf("authenticationManageService repo init failed: %v", err)
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
	return a.repo.Authenticate(username, password)
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

func (a *authenticationManageService) SrvImplName() string {
	return "authenticationManageService"
}
