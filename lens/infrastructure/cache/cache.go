package cache

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/errcode"
)

var _eg = errcode.NewErrorGroup(4, "cache")

var (
	ErrFailedToSetCache = _eg.CreateError(0, "failed to set cache")
)

type CacheKey string

type ICache interface {
	scene.Repository
	Get(key CacheKey) (string, bool)
	Set(key CacheKey, value string) error
	Delete(key CacheKey) error
}
