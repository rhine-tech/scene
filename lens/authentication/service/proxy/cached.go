package proxy

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/cache"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/model"
)

const (
	userCacheTTL = 30 * time.Second
	listCacheTTL = 15 * time.Second
)

type CachedAuthenticationService struct {
	base  authentication.IAuthenticationService `aperture:"embed"`
	cache cache.ITaggedCache                    `aperture:"optional"`
	log   logger.ILogger                        `aperture:"optional"`

	cacheCli *cache.Client
}

func NewCachedAuthenticationService(base authentication.IAuthenticationService) authentication.IAuthenticationService {
	return &CachedAuthenticationService{base: base}
}

func (c *CachedAuthenticationService) SrvImplName() scene.ImplName {
	return authentication.Lens.ImplName("IAuthenticationService", "cached")
}

func (c *CachedAuthenticationService) Setup() error {
	if c.cache == nil {
		if c.log != nil {
			c.log.Warnf("cache.ITaggedCache not found, all request will directly pass to service")
		}
		return nil
	}

	c.cacheCli = cache.NewClient(c.cache)
	if c.log != nil {
		c.log.Infof("cache.ITaggedCache found, using cache with %s", c.cache.ImplName().Identifier())
	}
	return nil
}

func (c *CachedAuthenticationService) AddUser(username, password string) (authentication.User, error) {
	user, err := c.base.AddUser(username, password)
	if err != nil {
		return authentication.User{}, err
	}
	c.invalidateUsers(user.UserID)
	return user, nil
}

func (c *CachedAuthenticationService) DeleteUser(userID string) error {
	if err := c.base.DeleteUser(userID); err != nil {
		return err
	}
	c.invalidateUsers(userID)
	return nil
}

func (c *CachedAuthenticationService) UpdateUser(user authentication.User) error {
	if err := c.base.UpdateUser(user); err != nil {
		return err
	}
	c.invalidateUsers(user.UserID)
	return nil
}

func (c *CachedAuthenticationService) Authenticate(username string, password string) (string, error) {
	return c.base.Authenticate(username, password)
}

func (c *CachedAuthenticationService) AuthenticateByToken(token string) (string, error) {
	return c.base.AuthenticateByToken(token)
}

func (c *CachedAuthenticationService) HasUser(userID string) (bool, error) {
	_, err := c.UserById(userID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, authentication.ErrUserNotFound) {
		return false, nil
	}
	return false, err
}

func (c *CachedAuthenticationService) UserById(userID string) (authentication.User, error) {
	if c.cache == nil {
		return c.base.UserById(userID)
	}
	return c.getOrLoadUser(fmt.Sprintf("authentication:user:id:v1:%s", userID), func() (authentication.User, error) {
		return c.base.UserById(userID)
	})
}

func (c *CachedAuthenticationService) UserByName(username string) (authentication.User, error) {
	if c.cache == nil {
		return c.base.UserByName(username)
	}
	return c.getOrLoadUser(fmt.Sprintf("authentication:user:name:v1:%s", username), func() (authentication.User, error) {
		return c.base.UserByName(username)
	})
}

func (c *CachedAuthenticationService) UserByEmail(email string) (authentication.User, error) {
	if c.cache == nil {
		return c.base.UserByEmail(email)
	}
	return c.getOrLoadUser(fmt.Sprintf("authentication:user:email:v1:%s", email), func() (authentication.User, error) {
		return c.base.UserByEmail(email)
	})
}

func (c *CachedAuthenticationService) ListUsers(offset, limit int64) (model.PaginationResult[authentication.User], error) {
	if c.cacheCli == nil {
		return c.base.ListUsers(offset, limit)
	}

	key := fmt.Sprintf("authentication:user:list:v1:%d:%d", offset, limit)
	val, err := cache.GetOrLoad(context.Background(), c.cacheCli, key, cache.GetOrLoadPolicy[model.PaginationResult[authentication.User]]{
		TTL:  listCacheTTL,
		Tags: []string{authenticationUserListTag()},
	}, func(_ context.Context) (model.PaginationResult[authentication.User], error) {
		return c.base.ListUsers(offset, limit)
	})
	if err != nil {
		if c.log != nil {
			c.log.WarnW("authentication user list cache load failed, fallback to source", "offset", offset, "limit", limit, "error", err)
		}
		return c.base.ListUsers(offset, limit)
	}
	return val, nil
}

func (c *CachedAuthenticationService) getOrLoadUser(
	key string,
	load func() (authentication.User, error),
) (authentication.User, error) {
	ctx := context.Background()
	codec := cache.JSONCodec{}

	raw, hit, err := c.cache.Get(ctx, key)
	if err == nil && hit {
		var user authentication.User
		decodeErr := codec.Unmarshal(raw, &user)
		if decodeErr == nil {
			return user, nil
		}
		if c.log != nil {
			c.log.WarnW("failed to decode cached authentication user, fallback to source", "key", key, "error", decodeErr)
		}
	} else if err != nil && c.log != nil {
		c.log.WarnW("failed to read cached authentication user, fallback to source", "key", key, "error", err)
	}

	user, err := load()
	if err != nil {
		return authentication.User{}, err
	}
	if user.UserID == "" {
		return user, nil
	}

	raw, err = codec.Marshal(user)
	if err != nil {
		if c.log != nil {
			c.log.WarnW("failed to encode authentication user for cache", "key", key, "userID", user.UserID, "error", err)
		}
		return user, nil
	}

	if err = c.cache.SetWithTags(ctx, key, raw, userCacheTTL, authenticationUserTag(user.UserID)); err != nil && c.log != nil {
		c.log.WarnW("failed to write authentication user cache", "key", key, "userID", user.UserID, "error", err)
	}
	return user, nil
}

func (c *CachedAuthenticationService) invalidateUsers(userIDs ...string) {
	if c.cache == nil {
		return
	}

	tags := []string{authenticationUserListTag()}
	for _, userID := range userIDs {
		if userID == "" {
			continue
		}
		tags = append(tags, authenticationUserTag(userID))
	}
	if err := c.cache.InvalidateTags(context.Background(), tags...); err != nil && c.log != nil {
		c.log.WarnW("failed to invalidate authentication cache", "tags", tags, "error", err)
	}
}

func authenticationUserTag(userID string) string {
	return fmt.Sprintf("authentication:user:%s", userID)
}

func authenticationUserListTag() string {
	return "authentication:user:list"
}
