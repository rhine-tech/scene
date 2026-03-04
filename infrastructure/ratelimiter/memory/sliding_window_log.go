package memory

import (
	"context"
	"sync"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/ratelimiter"
)

// SlidingWindowLogLimiter implements sliding window log algorithm.
type SlidingWindowLogLimiter struct {
	limit   int
	window  time.Duration
	mu      sync.Mutex
	nowFunc func() time.Time
	logs    map[string][]time.Time
}

func NewSlidingWindowLogLimiter(limit int, window time.Duration) (*SlidingWindowLogLimiter, error) {
	if _, err := normalizeInput(limit, window, "bootstrap", 1); err != nil {
		return nil, err
	}
	return &SlidingWindowLogLimiter{
		limit:   limit,
		window:  window,
		nowFunc: time.Now,
		logs:    make(map[string][]time.Time),
	}, nil
}

func (l *SlidingWindowLogLimiter) ImplName() scene.ImplName {
	return ratelimiter.Lens.ImplName("Limiter", "sliding-window-log")
}

func (l *SlidingWindowLogLimiter) Allow(ctx context.Context, key string) (ratelimiter.Decision, error) {
	return l.AllowN(ctx, key, 1)
}

func (l *SlidingWindowLogLimiter) AllowN(_ context.Context, key string, n int) (ratelimiter.Decision, error) {
	trimmed, err := normalizeInput(l.limit, l.window, key, n)
	if err != nil {
		return ratelimiter.Decision{}, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.nowFunc()
	cutoff := now.Add(-l.window)
	events := l.logs[trimmed]

	valid := events[:0]
	for _, t := range events {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid)+n > l.limit {
		l.logs[trimmed] = valid
		requiredExpire := len(valid) + n - l.limit
		retryAt := valid[requiredExpire-1].Add(l.window)
		retryAfter := retryAt.Sub(now)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return ratelimiter.Decision{
			Allowed:    false,
			Limit:      l.limit,
			Remaining:  l.limit - len(valid),
			RetryAfter: retryAfter,
		}, nil
	}

	for range n {
		valid = append(valid, now)
	}
	l.logs[trimmed] = valid
	return ratelimiter.Decision{
		Allowed:    true,
		Limit:      l.limit,
		Remaining:  l.limit - len(valid),
		RetryAfter: 0,
	}, nil
}

