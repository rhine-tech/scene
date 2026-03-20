package repository

import (
	"context"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/cache"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/logger"
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

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	val, err := r.ds.Get(ctx, key)
	if err != nil {
		return nil, false, nil
	}
	return []byte(val), true, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration, _ ...string) error {
	if err := r.ds.Set(ctx, key, value, expiration); err != nil {
		return cache.ErrCacheWrite.WithDetail(err)
	}
	return nil
}

func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		if err := r.ds.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisCache) InvalidateTags(_ context.Context, _ ...string) error {
	return nil
}
