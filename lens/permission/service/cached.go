package service

import (
	"context"
	"fmt"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/cache"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/permission"
)

const permissionListCacheTTL = 30 * time.Second

type CachedPermissionService struct {
	base  permission.PermissionService `aperture:"embed"`
	cache cache.ICache                 `aperture:"optional"`
	log   logger.ILogger               `aperture:"optional"`

	cacheCli *cache.Client
}

func NewCachedPermissionService(base permission.PermissionService) permission.PermissionService {
	return &CachedPermissionService{base: base}
}

func (c *CachedPermissionService) SrvImplName() scene.ImplName {
	return permission.Lens.ImplName("PermissionService", "cached")
}

func (c *CachedPermissionService) Setup() error {
	if c.cache != nil {
		c.cacheCli = cache.NewClient(c.cache)
		c.log.Infof("cache.ICache found, using cache with %s ", c.cache.ImplName().Identifier())
	} else {
		c.log.Warnf("cache.ICache not found, all request will directly pass to service")
	}
	return nil
}

func (c *CachedPermissionService) HasPermission(owner string, perm *permission.Permission) bool {
	perms := c.ListPermissions(owner)
	tree := permission.BuildTree(perms...)
	return tree.HasPermission(perm)
}

func (c *CachedPermissionService) HasPermissionStr(owner string, perm string) bool {
	p1, err := permission.ParsePermission(perm)
	if err != nil {
		return false
	}
	return c.HasPermission(owner, p1)
}

func (c *CachedPermissionService) ListPermissions(owner string) []*permission.Permission {
	if c.cacheCli == nil {
		return c.base.ListPermissions(owner)
	}
	key := fmt.Sprintf("permission:list:v1:%s", owner)
	tag := permissionOwnerTag(owner)
	val, err := cache.GetOrLoad(context.Background(), c.cacheCli, key, cache.GetOrLoadPolicy[[]*permission.Permission]{
		TTL:  permissionListCacheTTL,
		Tags: []string{tag},
	}, func(_ context.Context) ([]*permission.Permission, error) {
		return c.base.ListPermissions(owner), nil
	})
	if err != nil {
		if c.log != nil {
			c.log.WarnW("permission cache load failed, fallback to source", "owner", owner, "error", err)
		}
		return c.base.ListPermissions(owner)
	}
	return val
}

func (c *CachedPermissionService) AddPermission(owner string, perm string) error {
	if err := c.base.AddPermission(owner, perm); err != nil {
		return err
	}
	c.invalidateOwner(owner)
	return nil
}

func (c *CachedPermissionService) RemovePermission(owner string, perm string) error {
	if err := c.base.RemovePermission(owner, perm); err != nil {
		return err
	}
	c.invalidateOwner(owner)
	return nil
}

func (c *CachedPermissionService) invalidateOwner(owner string) {
	if c.cache == nil {
		return
	}
	tag := permissionOwnerTag(owner)
	if err := c.cache.InvalidateTags(context.Background(), tag); err != nil && c.log != nil {
		c.log.WarnW("failed to invalidate permission cache", "tag", tag, "error", err)
	}
}

func permissionOwnerTag(owner string) string {
	return fmt.Sprintf("permission:owner:%s", owner)
}
