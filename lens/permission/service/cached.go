package service

import (
	"context"
	"fmt"
	"time"

	"github.com/rhine-tech/scene"
	scache "github.com/rhine-tech/scene/infrastructure/cache"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/permission"
)

const permissionListCacheTTL = 30 * time.Second

type CachedPermissionService struct {
	Base      permission.PermissionService `aperture:"embed"`
	CacheRepo scache.ICache                `aperture:"optional"`
	Logger    logger.ILogger               `aperture:""`

	cacheCli *scache.Client
}

func NewCachedPermissionService(base permission.PermissionService) *CachedPermissionService {
	return &CachedPermissionService{Base: base}
}

func (c *CachedPermissionService) SrvImplName() scene.ImplName {
	return permission.Lens.ImplName("PermissionService", "cached")
}

func (c *CachedPermissionService) Setup() error {
	if c.CacheRepo != nil {
		c.cacheCli = scache.NewClient(c.CacheRepo)
		c.Logger.Infof("cache.ICache found, using cache with %s ", c.CacheRepo.ImplName().Identifier())
	} else {
		c.Logger.Warnf("cache.ICache not found, all request will directly pass to service")
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
		return c.Base.ListPermissions(owner)
	}
	key := fmt.Sprintf("permission:list:v1:%s", owner)
	tag := permissionOwnerTag(owner)
	val, err := scache.GetOrLoad(context.Background(), c.cacheCli, key, scache.GetOrLoadPolicy[[]*permission.Permission]{
		TTL:  permissionListCacheTTL,
		Tags: []string{tag},
	}, func(_ context.Context) ([]*permission.Permission, error) {
		return c.Base.ListPermissions(owner), nil
	})
	if err != nil {
		if c.Logger != nil {
			c.Logger.WarnW("permission cache load failed, fallback to source", "owner", owner, "error", err)
		}
		return c.Base.ListPermissions(owner)
	}
	return val
}

func (c *CachedPermissionService) AddPermission(owner string, perm string) error {
	if err := c.Base.AddPermission(owner, perm); err != nil {
		return err
	}
	c.invalidateOwner(owner)
	return nil
}

func (c *CachedPermissionService) RemovePermission(owner string, perm string) error {
	if err := c.Base.RemovePermission(owner, perm); err != nil {
		return err
	}
	c.invalidateOwner(owner)
	return nil
}

func (c *CachedPermissionService) invalidateOwner(owner string) {
	if c.CacheRepo == nil {
		return
	}
	tag := permissionOwnerTag(owner)
	if err := c.CacheRepo.InvalidateTags(context.Background(), tag); err != nil && c.Logger != nil {
		c.Logger.WarnW("failed to invalidate permission cache", "tag", tag, "error", err)
	}
}

func permissionOwnerTag(owner string) string {
	return fmt.Sprintf("permission:owner:%s", owner)
}
