package repository

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/cache"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/logger"
)

const redisTaggedValuePrefix = "__scene_cache_tagged_v1__"

type redisTaggedValue struct {
	Value      []byte           `json:"value"`
	TagVersion map[string]int64 `json:"tag_versions,omitempty"`
}

type RedisCache struct {
	ds  datasource.RedisDataSource `aperture:""`
	log logger.ILogger             `aperture:""`
}

func NewRedisCache(ds datasource.RedisDataSource) cache.ITaggedCache {
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
	if strings.HasPrefix(val, redisTaggedValuePrefix) {
		tagged, err := decodeRedisTaggedValue([]byte(strings.TrimPrefix(val, redisTaggedValuePrefix)))
		if err != nil {
			_ = r.ds.Delete(ctx, key)
			return nil, false, nil
		}
		valid, err := r.redisTagVersionsMatch(ctx, tagged.TagVersion)
		if err != nil || !valid {
			_ = r.ds.Delete(ctx, key)
			return nil, false, nil
		}
		return tagged.Value, true, nil
	}
	return []byte(val), true, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if err := r.ds.Set(ctx, key, value, expiration); err != nil {
		return cache.ErrCacheWrite.WithDetail(err)
	}
	return nil
}

func (r *RedisCache) SetWithTags(ctx context.Context, key string, value []byte, expiration time.Duration, tags ...string) error {
	cleanTags := cleanCacheTags(tags)
	if len(cleanTags) == 0 {
		return r.Set(ctx, key, value, expiration)
	}
	versions, err := r.redisTagVersions(ctx, cleanTags)
	if err != nil {
		return cache.ErrCacheRead.WithDetail(err)
	}
	raw, err := json.Marshal(redisTaggedValue{
		Value:      value,
		TagVersion: versions,
	})
	if err != nil {
		return cache.ErrCacheEncode.WithDetail(err)
	}
	if err := r.ds.Set(ctx, key, append([]byte(redisTaggedValuePrefix), raw...), expiration); err != nil {
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

func (r *RedisCache) InvalidateTags(ctx context.Context, tags ...string) error {
	for _, tag := range cleanCacheTags(tags) {
		if _, err := r.ds.Incr(ctx, redisTagVersionKey(tag)); err != nil {
			return cache.ErrCacheWrite.WithDetail(err)
		}
	}
	return nil
}

func (r *RedisCache) redisTagVersions(ctx context.Context, tags []string) (map[string]int64, error) {
	keys := make([]string, len(tags))
	for i, tag := range tags {
		keys[i] = redisTagVersionKey(tag)
	}
	values, err := r.ds.MGet(ctx, keys...)
	if err != nil {
		return nil, err
	}
	versions := make(map[string]int64, len(tags))
	for i, tag := range tags {
		versions[tag] = parseRedisTagVersion(values[i])
	}
	return versions, nil
}

func (r *RedisCache) redisTagVersionsMatch(ctx context.Context, expected map[string]int64) (bool, error) {
	if len(expected) == 0 {
		return true, nil
	}
	tags := make([]string, 0, len(expected))
	for tag := range expected {
		tags = append(tags, tag)
	}
	current, err := r.redisTagVersions(ctx, tags)
	if err != nil {
		return false, err
	}
	for tag, version := range expected {
		if current[tag] != version {
			return false, nil
		}
	}
	return true, nil
}

func decodeRedisTaggedValue(raw []byte) (redisTaggedValue, error) {
	var value redisTaggedValue
	err := json.Unmarshal(raw, &value)
	return value, err
}

func redisTagVersionKey(tag string) string {
	return "scene:cache:tagver:" + tag
}

func cleanCacheTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return result
}

func parseRedisTagVersion(raw string) int64 {
	if raw == "" {
		return 0
	}
	version, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return version
}
