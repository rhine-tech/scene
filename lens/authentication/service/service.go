package service

import (
	"errors"
	"github.com/google/uuid"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/authentication"
	"strings"
)

type authenticationService struct {
	logger    logger.ILogger                           `aperture:""`
	userRepo  authentication.IAuthenticationRepository `aperture:""`
	tokenRepo authentication.IAccessTokenService       `aperture:""`
}

func (s *authenticationService) SrvImplName() scene.ImplName {
	return authentication.Lens.ImplName("IAuthenticationService", "default")
}

func (s *authenticationService) Setup() error {
	s.logger = s.logger.WithPrefix(s.SrvImplName().Identifier())
	return nil
}

// NewAuthenticationService 创建 IAuthenticationService 的实例
// 注意：构造函数现在需要两个 repository
func NewAuthenticationService(
	logger logger.ILogger,
	userRepo authentication.IAuthenticationRepository,
	tokenRepo authentication.IAccessTokenService) authentication.IAuthenticationService {
	return &authenticationService{
		logger:    logger,
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}
}

func (s *authenticationService) AddUser(username, password string) (authentication.User, error) {
	_, err := s.userRepo.UserByName(username)
	if err == nil {
		return authentication.User{}, authentication.ErrUserAlreadyExists
	}
	if !errors.Is(err, authentication.ErrUserAlreadyExists) {
		s.logger.ErrorW("failed to check user existence before adding", "username", username, "error", err)
		return authentication.User{}, authentication.ErrFailToAddUser.Wrap(err)
	}

	newUser := authentication.User{
		UserID:   strings.ReplaceAll(uuid.NewString(), "-", ""), // Service 层负责生成唯一ID
		Username: username,
		Password: password, // 注意：实际项目中应在此处加密密码
	}
	s.logger.InfoW("adding user", "username", username, "userId", newUser.UserID)

	createdUser, err := s.userRepo.AddUser(newUser)
	if err != nil {
		s.logger.ErrorW("failed to add user in repository", "username", username, "error", err)
		return authentication.User{}, authentication.ErrFailToAddUser.WrapIfNot(err)
	}
	return createdUser, nil
}

func (s *authenticationService) DeleteUser(userId string) error {
	s.logger.Warnf("deleting user %s", userId)
	err := s.userRepo.DeleteUser(userId)
	if err != nil {
		s.logger.ErrorW("failed to delete user in repository", "username", userId, "error", err)
		return authentication.ErrInternalError.WrapIfNot(err)
	}
	return nil
}

func (s *authenticationService) UpdateUser(user authentication.User) error {
	s.logger.Infof("updating user %s", user.UserID)
	err := s.userRepo.UpdateUser(user)
	if err != nil {
		s.logger.ErrorW("failed to update user in repository", "username", user.UserID, "error", err)
		return authentication.ErrInternalError.WrapIfNot(err)
	}
	return nil
}

func (s *authenticationService) Authenticate(username string, password string) (string, error) {
	s.logger.Debugf("authenticating user %s by password", username)
	userId, err := s.userRepo.Authenticate(username, password)
	if err != nil {
		s.logger.Debugf("failed to authenticate user", "username", username, "error", err)
	}
	return userId, nil
}

// AuthenticateByToken 通过 Access Token 进行身份验证
func (s *authenticationService) AuthenticateByToken(token string) (string, error) {
	s.logger.Debugf("authenticating by token")
	userId, valid, err := s.tokenRepo.Validate(token)
	if err != nil {
		if errors.Is(err, authentication.ErrTokenNotFound) {
			return "", authentication.ErrAuthenticationFailed
		}
		return "", err
	}
	if !valid {
		return "", authentication.ErrAuthenticationFailed
	}
	// 可选：在这里可以检查 token 是否过期或有其他状态
	return userId, nil
}

func (s *authenticationService) UserById(userId string) (authentication.User, error) {
	s.logger.Debugf("getting user by id %s", userId)
	user, err := s.userRepo.UserById(userId)
	if err != nil {
		s.logger.ErrorW("failed to get user by id", "userId", userId, "error", err)
		return authentication.User{}, authentication.ErrInternalError.WrapIfNot(err)
	}
	return user, nil
}

func (s *authenticationService) UserByName(username string) (authentication.User, error) {
	s.logger.Debugf("getting user by name %s", username)
	user, err := s.userRepo.UserByName(username)
	if err != nil {
		s.logger.ErrorW("failed to get user by name", "username", username, "error", err)
		return authentication.User{}, authentication.ErrInternalError.WrapIfNot(err)
	}
	return user, nil
}

func (s *authenticationService) UserByEmail(email string) (authentication.User, error) {
	s.logger.Debugf("getting user by email %s", email)
	user, err := s.userRepo.UserByEmail(email)
	if err != nil {
		s.logger.ErrorW("failed to get user by email", "email", email, "error", err)
		return authentication.User{}, authentication.ErrInternalError.WrapIfNot(err)
	}
	return user, nil
}
