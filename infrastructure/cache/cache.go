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

type ICache interface {
	scene.Named
	Get(key string) (string, bool)
	Set(key string, value string, expiration time.Duration) error
	Delete(key string) error
}
