package memory

import (
	"context"
	"sync"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/ratelimiter"
)

type fixedWindowState struct {
	windowID int64
	count    int
}

// FixedWindowLimiter implements fixed window counter algorithm.
type FixedWindowLimiter struct {
	limit   int
	window  time.Duration
	mu      sync.Mutex
	nowFunc func() time.Time
	counts  map[string]fixedWindowState
}

func NewFixedWindowLimiter(limit int, window time.Duration) (*FixedWindowLimiter, error) {
	if _, err := normalizeInput(limit, window, "bootstrap", 1); err != nil {
		return nil, err
	}
	return &FixedWindowLimiter{
		limit:   limit,
		window:  window,
		nowFunc: time.Now,
		counts:  make(map[string]fixedWindowState),
	}, nil
}

func (l *FixedWindowLimiter) ImplName() scene.ImplName {
	return ratelimiter.Lens.ImplName("Limiter", "fixed-window")
}

func (l *FixedWindowLimiter) Allow(ctx context.Context, key string) (ratelimiter.Decision, error) {
	return l.AllowN(ctx, key, 1)
}

func (l *FixedWindowLimiter) AllowN(_ context.Context, key string, n int) (ratelimiter.Decision, error) {
	trimmed, err := normalizeInput(l.limit, l.window, key, n)
	if err != nil {
		return ratelimiter.Decision{}, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.nowFunc()
	windowID := now.UnixNano() / l.window.Nanoseconds()
	state, ok := l.counts[trimmed]
	if !ok || state.windowID != windowID {
		state = fixedWindowState{
			windowID: windowID,
			count:    0,
		}
	}

	if state.count+n > l.limit {
		nextWindowAt := time.Unix(0, (windowID+1)*l.window.Nanoseconds())
		retryAfter := nextWindowAt.Sub(now)
		if retryAfter < 0 {
			retryAfter = 0
		}
		l.counts[trimmed] = state
		return ratelimiter.Decision{
			Allowed:    false,
			Limit:      l.limit,
			Remaining:  l.limit - state.count,
			RetryAfter: retryAfter,
		}, nil
	}

	state.count += n
	l.counts[trimmed] = state
	return ratelimiter.Decision{
		Allowed:    true,
		Limit:      l.limit,
		Remaining:  l.limit - state.count,
		RetryAfter: 0,
	}, nil
}

