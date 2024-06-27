package cache

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/errcode"
	"time"
)

var _eg = errcode.NewErrorGroup(4, "cache")

const Lens scene.InfraName = "cache"

var (
	ErrFailedToSetCache = _eg.CreateError(0, "failed to set cache")
)

var NoExpiration = time.Duration(-1)

type CacheKey string

type ICache interface {
	scene.Repository
	Get(key CacheKey) (string, bool)
	Set(key CacheKey, value string, expiration time.Duration) error
	Delete(key CacheKey) error
}
