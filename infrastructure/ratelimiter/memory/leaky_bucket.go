package memory

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/ratelimiter"
)

type leakyBucketState struct {
	level      float64
	lastLeaked time.Time
}

// LeakyBucketLimiter implements leaky bucket algorithm.
type LeakyBucketLimiter struct {
	limit   int
	window  time.Duration
	mu      sync.Mutex
	nowFunc func() time.Time
	buckets map[string]leakyBucketState
}

func NewLeakyBucketLimiter(limit int, window time.Duration) (*LeakyBucketLimiter, error) {
	if _, err := normalizeInput(limit, window, "bootstrap", 1); err != nil {
		return nil, err
	}
	return &LeakyBucketLimiter{
		limit:   limit,
		window:  window,
		nowFunc: time.Now,
		buckets: make(map[string]leakyBucketState),
	}, nil
}

func (l *LeakyBucketLimiter) ImplName() scene.ImplName {
	return ratelimiter.Lens.ImplName("Limiter", "leaky-bucket")
}

func (l *LeakyBucketLimiter) Allow(ctx context.Context, key string) (ratelimiter.Decision, error) {
	return l.AllowN(ctx, key, 1)
}

func (l *LeakyBucketLimiter) AllowN(_ context.Context, key string, n int) (ratelimiter.Decision, error) {
	trimmed, err := normalizeInput(l.limit, l.window, key, n)
	if err != nil {
		return ratelimiter.Decision{}, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.nowFunc()
	state, ok := l.buckets[trimmed]
	if !ok {
		state = leakyBucketState{
			level:      0,
			lastLeaked: now,
		}
	}

	leakRate := float64(l.limit) / l.window.Seconds()
	elapsed := now.Sub(state.lastLeaked).Seconds()
	if elapsed > 0 {
		state.level = math.Max(0, state.level-elapsed*leakRate)
		state.lastLeaked = now
	}

	nextLevel := state.level + float64(n)
	if nextLevel > float64(l.limit) {
		l.buckets[trimmed] = state
		overflow := nextLevel - float64(l.limit)
		retryAfter := time.Duration((overflow / leakRate) * float64(time.Second))
		if retryAfter < 0 {
			retryAfter = 0
		}
		return ratelimiter.Decision{
			Allowed:    false,
			Limit:      l.limit,
			Remaining:  int(math.Floor(float64(l.limit) - state.level)),
			RetryAfter: retryAfter,
		}, nil
	}

	state.level = nextLevel
	l.buckets[trimmed] = state
	return ratelimiter.Decision{
		Allowed:    true,
		Limit:      l.limit,
		Remaining:  int(math.Floor(float64(l.limit) - state.level)),
		RetryAfter: 0,
	}, nil
}

