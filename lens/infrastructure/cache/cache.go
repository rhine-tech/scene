package cache

import "github.com/rhine-tech/scene/errcode"

var _eg = errcode.NewErrorGroup(2, "cache")

var (
	ErrFailedToSetCache = _eg.CreateError(0, "failed to set cache")
)

type CacheKey string

type ICache interface {
	Get(key CacheKey) (string, bool)
	Set(key CacheKey, value string) error
}
