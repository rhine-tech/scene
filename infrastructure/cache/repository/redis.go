package repository

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/cache"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/logger"
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

func (r *RedisCache) ImplName() scene.ImplName {
	return cache.Lens.ImplName("ICache", "redis")
}

func (r *RedisCache) Status() error {
	return r.ds.Status()
}

func (r *RedisCache) Setup() error {
	r.log = r.log.WithPrefix(r.ImplName().Identifier())
	if err := r.Status(); err != nil {
		r.log.Error("setup redis cache failed")
		return err
	}
	r.log.Info("setup redis cache succeed")
	return nil
}

func (r *RedisCache) Get(key string) (string, bool) {
	val, err := r.ds.Get(registry.EmptyContext, key)
	if err != nil {
		return val, false
	}
	return val, true
}

func (r *RedisCache) Set(key string, value string, expiration time.Duration) error {
	if err := r.ds.Set(registry.EmptyContext, key, value, expiration); err != nil {
		return cache.ErrFailedToSetCache.WithDetail(err)
	}
	return nil
}

func (r *RedisCache) Delete(key string) error {
	return r.ds.Delete(registry.EmptyContext, key)
}
