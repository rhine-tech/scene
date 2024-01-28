package repository

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/cache"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"time"
)

type RedisCache struct {
	ds  datasource.RedisDataSource `aperture:""`
	log logger.ILogger             `aperture:""`
}

func NewRedisCache(ds datasource.RedisDataSource) cache.ICache {
	return &RedisCache{ds: ds}
}

func (r *RedisCache) RepoImplName() scene.ImplName {
	return scene.NewRepoImplName("cache", "ICache", "redis")
}

func (r *RedisCache) Status() error {
	return r.ds.Status()
}

func (r *RedisCache) Setup() error {
	r.log = r.log.WithPrefix(r.RepoImplName().Identifier())
	if err := r.Status(); err != nil {
		r.log.Error("setup redis cache failed")
		return err
	}
	r.log.Info("setup redis cache succeed")
	return nil
}

func (r *RedisCache) Get(key cache.CacheKey) (string, bool) {
	val, err := r.ds.Get(registry.EmptyContext, string(key))
	if err != nil {
		return val, false
	}
	return val, true
}

func (r *RedisCache) Set(key cache.CacheKey, value string, expiration time.Duration) error {
	if err := r.ds.Set(registry.EmptyContext, string(key), value, expiration); err != nil {
		return cache.ErrFailedToSetCache.WithDetail(err)
	}
	return nil
}

func (r *RedisCache) Delete(key cache.CacheKey) error {
	return r.ds.Delete(registry.EmptyContext, string(key))
}
