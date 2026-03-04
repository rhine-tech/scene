package memory

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/ratelimiter"
)

type tokenBucketState struct {
	tokens     float64
	lastRefill time.Time
}

// TokenBucketLimiter implements token bucket algorithm.
type TokenBucketLimiter struct {
	limit   int
	window  time.Duration
	mu      sync.Mutex
	nowFunc func() time.Time
	buckets map[string]tokenBucketState
}

func NewTokenBucketLimiter(limit int, window time.Duration) (*TokenBucketLimiter, error) {
	if _, err := normalizeInput(limit, window, "bootstrap", 1); err != nil {
		return nil, err
	}
	return &TokenBucketLimiter{
		limit:   limit,
		window:  window,
		nowFunc: time.Now,
		buckets: make(map[string]tokenBucketState),
	}, nil
}

func (l *TokenBucketLimiter) ImplName() scene.ImplName {
	return ratelimiter.Lens.ImplName("Limiter", "token-bucket")
}

func (l *TokenBucketLimiter) Allow(ctx context.Context, key string) (ratelimiter.Decision, error) {
	return l.AllowN(ctx, key, 1)
}

func (l *TokenBucketLimiter) AllowN(_ context.Context, key string, n int) (ratelimiter.Decision, error) {
	trimmed, err := normalizeInput(l.limit, l.window, key, n)
	if err != nil {
		return ratelimiter.Decision{}, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.nowFunc()
	state, ok := l.buckets[trimmed]
	if !ok {
		state = tokenBucketState{
			tokens:     float64(l.limit),
			lastRefill: now,
		}
	}

	refillRate := float64(l.limit) / l.window.Seconds()
	if refillRate > 0 {
		elapsed := now.Sub(state.lastRefill).Seconds()
		if elapsed > 0 {
			state.tokens = math.Min(float64(l.limit), state.tokens+elapsed*refillRate)
			state.lastRefill = now
		}
	}

	need := float64(n)
	if state.tokens < need {
		l.buckets[trimmed] = state
		missing := need - state.tokens
		retryAfter := time.Duration((missing / refillRate) * float64(time.Second))
		if retryAfter < 0 {
			retryAfter = 0
		}
		return ratelimiter.Decision{
			Allowed:    false,
			Limit:      l.limit,
			Remaining:  int(math.Floor(state.tokens)),
			RetryAfter: retryAfter,
		}, nil
	}

	state.tokens -= need
	l.buckets[trimmed] = state
	return ratelimiter.Decision{
		Allowed:    true,
		Limit:      l.limit,
		Remaining:  int(math.Floor(state.tokens)),
		RetryAfter: 0,
	}, nil
}

