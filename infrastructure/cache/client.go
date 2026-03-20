package cache

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type Client struct {
	store       ICache
	codec       Codec
	metrics     Metrics
	jitterRatio float64

	group singleflight.Group
	rngMu sync.Mutex
	rng   *rand.Rand
}

type ClientOption func(*Client)

func WithCodec(codec Codec) ClientOption {
	return func(c *Client) {
		if codec != nil {
			c.codec = codec
		}
	}
}

func WithMetrics(metrics Metrics) ClientOption {
	return func(c *Client) {
		if metrics != nil {
			c.metrics = metrics
		}
	}
}

// WithTTLJitter configures random ttl jitter ratio in [0, 1].
// Example: 0.1 means final ttl is randomized in [0.9x, 1.1x].
func WithTTLJitter(ratio float64) ClientOption {
	return func(c *Client) {
		switch {
		case ratio <= 0:
			c.jitterRatio = 0
		case ratio > 1:
			c.jitterRatio = 1
		default:
			c.jitterRatio = ratio
		}
	}
}

func NewClient(store ICache, opts ...ClientOption) *Client {
	client := &Client{
		store:       store,
		codec:       JSONCodec{},
		metrics:     NopMetrics{},
		jitterRatio: 0.1,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(client)
		}
	}
	return client
}

type Loader[T any] func(ctx context.Context) (T, error)

type GetOrLoadPolicy[T any] struct {
	TTL       time.Duration
	Tags      []string
	Cacheable func(v T) bool
}

// GetOrLoad returns cached value first. On miss, it loads from source once per key
// (singleflight), writes it into cache, and returns the loaded value.
func GetOrLoad[T any](
	ctx context.Context,
	client *Client,
	key string,
	policy GetOrLoadPolicy[T],
	load Loader[T],
) (T, error) {
	var zero T
	if client == nil || client.store == nil {
		return zero, ErrInvalidCacheClient
	}
	if key == "" {
		return zero, ErrInvalidCacheKey
	}
	if load == nil {
		return zero, ErrInvalidLoader
	}

	if value, hit, ok := tryRead[T](ctx, client, key); ok {
		if hit {
			return value, nil
		}
	}

	res, err, _ := client.group.Do(key, func() (any, error) {
		if value, hit, ok := tryRead[T](ctx, client, key); ok && hit {
			return value, nil
		}
		loadStart := time.Now()
		value, loadErr := load(ctx)
		client.metrics.RecordLoad(time.Since(loadStart), loadErr)
		if loadErr != nil {
			return zero, loadErr
		}
		shouldCache := policy.Cacheable == nil || policy.Cacheable(value)
		if shouldCache && policy.TTL > 0 {
			raw, marshalErr := client.codec.Marshal(value)
			if marshalErr != nil {
				return zero, ErrCacheEncode.WithDetail(marshalErr)
			}
			setStart := time.Now()
			setErr := client.store.Set(ctx, key, raw, client.withJitter(policy.TTL), policy.Tags...)
			client.metrics.RecordSet(time.Since(setStart), setErr)
			if setErr != nil {
				return zero, ErrCacheWrite.WrapIfNot(setErr)
			}
		}
		return value, nil
	})
	if err != nil {
		return zero, err
	}
	casted, ok := res.(T)
	if !ok {
		return zero, ErrCacheDecode.WithDetailStr("singleflight type assertion failed")
	}
	return casted, nil
}

func Delete(ctx context.Context, client *Client, keys ...string) error {
	if client == nil || client.store == nil {
		return ErrInvalidCacheClient
	}
	if len(keys) == 0 {
		return nil
	}
	start := time.Now()
	err := client.store.Delete(ctx, keys...)
	client.metrics.RecordDelete(time.Since(start), err)
	if err != nil {
		return ErrCacheDelete.WrapIfNot(err)
	}
	return nil
}

func tryRead[T any](ctx context.Context, client *Client, key string) (value T, hit bool, ok bool) {
	var zero T
	start := time.Now()
	raw, found, err := client.store.Get(ctx, key)
	client.metrics.RecordGet(found, time.Since(start), err)
	if err != nil {
		return zero, false, false
	}
	if !found {
		return zero, false, true
	}
	if err = client.codec.Unmarshal(raw, &value); err != nil {
		client.metrics.RecordDecodeError(err)
		return zero, false, false
	}
	return value, true, true
}

func (c *Client) withJitter(ttl time.Duration) time.Duration {
	if ttl <= 0 || c.jitterRatio <= 0 {
		return ttl
	}
	c.rngMu.Lock()
	r := c.rng.Float64()*2 - 1
	c.rngMu.Unlock()
	factor := 1 + r*c.jitterRatio
	result := time.Duration(float64(ttl) * factor)
	if result <= 0 {
		return time.Millisecond
	}
	return result
}
