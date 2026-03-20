package cache

import (
	"context"
	"encoding/json"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/errcode"
	"time"
)

var _eg = errcode.NewErrorGroup(4, "cache")

const Lens scene.InfraName = "cache"

var (
	ErrInvalidCacheClient = _eg.CreateError(0, "invalid cache client")
	ErrInvalidCacheKey    = _eg.CreateError(1, "invalid cache key")
	ErrInvalidLoader      = _eg.CreateError(2, "invalid cache loader")
	ErrCacheRead          = _eg.CreateError(3, "cache read failed")
	ErrCacheWrite         = _eg.CreateError(4, "cache write failed")
	ErrCacheDelete        = _eg.CreateError(5, "cache delete failed")
	ErrCacheEncode        = _eg.CreateError(6, "cache encode failed")
	ErrCacheDecode        = _eg.CreateError(7, "cache decode failed")
)

var NoExpiration = time.Duration(-1)

type ICache interface {
	scene.Named
	Get(ctx context.Context, key string) (value []byte, hit bool, err error)
	Set(ctx context.Context, key string, value []byte, expiration time.Duration, tags ...string) error
	Delete(ctx context.Context, keys ...string) error
	InvalidateTags(ctx context.Context, tags ...string) error
}

type Codec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, dst any) error
}

type JSONCodec struct{}

func (JSONCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec) Unmarshal(data []byte, dst any) error {
	return json.Unmarshal(data, dst)
}
