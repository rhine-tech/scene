package memory

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/ratelimiter"
	"golang.org/x/time/rate"
)

// TimeRateLimiter wraps golang.org/x/time/rate limiter per key.
type TimeRateLimiter struct {
	limit   int
	window  time.Duration
	nowFunc func() time.Time

	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

func NewTimeRateLimiter(limit int, window time.Duration) (*TimeRateLimiter, error) {
	if _, err := normalizeInput(limit, window, "bootstrap", 1); err != nil {
		return nil, err
	}
	return &TimeRateLimiter{
		limit:    limit,
		window:   window,
		nowFunc:  time.Now,
		limiters: make(map[string]*rate.Limiter),
	}, nil
}

func (l *TimeRateLimiter) ImplName() scene.ImplName {
	return ratelimiter.Lens.ImplName("Limiter", "go-time-rate")
}

func (l *TimeRateLimiter) Allow(ctx context.Context, key string) (ratelimiter.Decision, error) {
	return l.AllowN(ctx, key, 1)
}

func (l *TimeRateLimiter) AllowN(_ context.Context, key string, n int) (ratelimiter.Decision, error) {
	trimmed, err := normalizeInput(l.limit, l.window, key, n)
	if err != nil {
		return ratelimiter.Decision{}, err
	}

	now := l.nowFunc()
	limiter := l.getOrCreateLimiter(trimmed)

	reservation := limiter.ReserveN(now, n)
	if !reservation.OK() {
		return ratelimiter.Decision{
			Allowed:    false,
			Limit:      l.limit,
			Remaining:  tokensToRemaining(limiter.TokensAt(now)),
			RetryAfter: l.window,
		}, nil
	}

	delay := reservation.DelayFrom(now)
	if delay > 0 {
		reservation.CancelAt(now)
		return ratelimiter.Decision{
			Allowed:    false,
			Limit:      l.limit,
			Remaining:  tokensToRemaining(limiter.TokensAt(now)),
			RetryAfter: delay,
		}, nil
	}

	return ratelimiter.Decision{
		Allowed:    true,
		Limit:      l.limit,
		Remaining:  tokensToRemaining(limiter.TokensAt(now)),
		RetryAfter: 0,
	}, nil
}

func (l *TimeRateLimiter) getOrCreateLimiter(key string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	existing, ok := l.limiters[key]
	if ok {
		return existing
	}
	perSecond := rate.Limit(float64(l.limit) / l.window.Seconds())
	created := rate.NewLimiter(perSecond, l.limit)
	l.limiters[key] = created
	return created
}

func tokensToRemaining(tokens float64) int {
	remaining := int(math.Floor(tokens))
	if remaining < 0 {
		return 0
	}
	return remaining
}

