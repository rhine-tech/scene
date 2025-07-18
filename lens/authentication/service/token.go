package service

import (
	"errors"
	"github.com/google/uuid"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/model"
	"strings"
	"time"
)

// accessTokenService provides the implementation for IAccessTokenService.
type accessTokenService struct {
	logger    logger.ILogger                        `aperture:""`
	tokenRepo authentication.IAccessTokenRepository `aperture:""`
	authSrv   authentication.IAuthenticationService `aperture:""`
}

// NewAccessTokenService creates a new instance of IAccessTokenService.
func NewAccessTokenService(
	repo authentication.IAccessTokenRepository,
	log logger.ILogger) authentication.IAccessTokenService {
	return &accessTokenService{
		logger:    log,
		tokenRepo: repo,
	}
}

func (s *accessTokenService) SrvImplName() scene.ImplName {
	return authentication.Lens.ImplName("IAccessTokenService", "default")
}

func (s *accessTokenService) Setup() error {
	s.logger = s.logger.WithPrefix(s.SrvImplName().Identifier())
	return nil
}

// Create generates a new unique token for a specified user.
func (s *accessTokenService) Create(userId, name string, expireAt int64) (authentication.AccessToken, error) {
	s.logger.InfoW("creating new access token", "userId", userId, "name", name)

	// check userId exists
	has, err := s.authSrv.HasUser(userId)
	if err != nil {
		return authentication.AccessToken{}, err
	}

	if !has {
		return authentication.AccessToken{}, authentication.ErrUserNotFound
	}

	// Create a new token object
	newToken := authentication.AccessToken{
		Token:     strings.ReplaceAll(uuid.NewString(), "-", ""), // Generate a unique token value
		UserID:    userId,
		Name:      name,
		CreatedAt: time.Now().Unix(),
		ExpireAt:  expireAt,
	}

	// Persist the token using the repository
	createdToken, err := s.tokenRepo.CreateToken(newToken)
	if err != nil {
		s.logger.ErrorW("failed to create token in repository", "userId", userId, "error", err)
		return authentication.AccessToken{}, err
	}

	s.logger.InfoW("successfully created access token", "userId", userId, "tokenId", createdToken.Token)
	return createdToken, nil
}

// ListByUser retrieves a paginated list of tokens for a specific user.
func (s *accessTokenService) ListByUser(userId string, offset, limit int64) (model.PaginationResult[authentication.AccessToken], error) {
	s.logger.Debugf("listing tokens for user %s with offset %d and limit %d", userId, offset, limit)

	result, err := s.tokenRepo.ListTokensByUser(userId, offset, limit)
	if err != nil {
		s.logger.ErrorW("failed to list tokens by user from repository", "userId", userId, "error", err)
		return model.PaginationResult[authentication.AccessToken]{}, authentication.ErrFailToGetToken.WrapIfNot(err)
	}

	return result, nil
}

// List retrieves a paginated list of all tokens in the system.
func (s *accessTokenService) List(offset, limit int64) (model.PaginationResult[authentication.AccessToken], error) {
	s.logger.Debugf("listing all tokens with offset %d and limit %d", offset, limit)

	result, err := s.tokenRepo.ListTokens(offset, limit)
	if err != nil {
		s.logger.ErrorW("failed to list all tokens from repository", "error", err)
		return model.PaginationResult[authentication.AccessToken]{}, authentication.ErrFailToGetToken.WrapIfNot(err)
	}

	return result, nil
}

// Delete removes a token, ensuring the user owns it first.
func (s *accessTokenService) Delete(tokenValue string) error {
	s.logger.InfoW("deleting token", "tokenValue", tokenValue)

	// Proceed with deletion
	if err := s.tokenRepo.DeleteToken(tokenValue); err != nil {
		s.logger.ErrorW("failed to delete token from repository", "tokenValue", tokenValue, "error", err)
		return authentication.ErrInternalError.WrapIfNot(err)
	}

	s.logger.InfoW("successfully deleted token", "tokenValue", tokenValue)
	return nil
}

// Validate checks if a token string is valid and not expired.
func (s *accessTokenService) Validate(tokenValue string) (userId string, valid bool, err error) {
	s.logger.Debugf("validating token")

	token, err := s.tokenRepo.GetTokenByValue(tokenValue)
	if err != nil {
		// If the token is not found in the repository
		if errors.Is(err, authentication.ErrTokenNotFound) {
			s.logger.Debugf("token not found", "token", tokenValue)
			return "", false, authentication.ErrTokenNotFound
		}
		// For other repository errors
		s.logger.ErrorW("failed to get token from repository for validation", "error", err)
		return "", false, authentication.ErrInternalError.WrapIfNot(err)
	}

	// Check if the token is expired
	if token.ExpireAt > 0 && time.Now().Unix() > token.ExpireAt {
		s.logger.Debugf("token is expired", "token", tokenValue, "expireAt", token.ExpireAt)
		return token.UserID, false, nil
	}

	s.logger.Debugf("token is valid", "token", tokenValue, "userId", token.UserID)
	return token.UserID, true, nil
}

func (s *accessTokenService) WithSceneContext(ctx scene.Context) authentication.IAccessTokenService {
	return &accessTokenServiceCtx{s: s, ctx: ctx}
}
