package repository

import (
	"github.com/rhine-tech/scene/lens/infrastructure/cache"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource"
	"github.com/rhine-tech/scene/registry"
)

type RedisCache struct {
	ds datasource.RedisDataSource
}

func NewRedisCache(ds datasource.RedisDataSource) cache.ICache {
	return &RedisCache{ds: ds}
}

func (r *RedisCache) Get(key cache.CacheKey) (string, bool) {
	val, err := r.ds.Get(registry.EmptyContext, string(key))
	if err != nil {
		return val, false
	}
	return val, true
}

func (r *RedisCache) Set(key cache.CacheKey, value string) error {
	if err := r.ds.Set(registry.EmptyContext, string(key), value, -1); err != nil {
		return cache.ErrFailedToSetCache.WithDetail(err)
	}
	return nil
}
