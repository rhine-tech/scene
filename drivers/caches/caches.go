package caches

import (
	"encoding/json"
	"github.com/rhine-tech/scene/lens/infrastructure/cache"
)

type Cache[Model any] struct {
	prefix string
	icache cache.ICache
}

func UseCache[Model any](prefix string, icache cache.ICache) *Cache[Model] {
	return &Cache[Model]{prefix: prefix, icache: icache}
}

func (c *Cache[Model]) getKey(key string) cache.CacheKey {
	return cache.CacheKey(c.prefix + "_" + key)
}

func (c *Cache[Model]) Get(key string) (Model, bool) {
	var val Model
	strVal, ok := c.icache.Get(c.getKey(key))
	if !ok {
		return val, false
	}
	return val, json.Unmarshal([]byte(strVal), &val) == nil
}

func (c *Cache[Model]) Set(key string, value Model) error {
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.icache.Set(c.getKey(key), string(val))
}

func (c *Cache[Model]) Delete(key string) error {
	return c.icache.Delete(c.getKey(key))
}
