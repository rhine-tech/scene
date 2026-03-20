package memory

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/ratelimiter"
)

type slidingCounterState struct {
	windowID  int64
	prevCount int
	currCount int
}

// SlidingWindowCounterLimiter implements sliding window counter algorithm.
type SlidingWindowCounterLimiter struct {
	limit   int
	window  time.Duration
	mu      sync.Mutex
	nowFunc func() time.Time
	counts  map[string]slidingCounterState
}

func NewSlidingWindowCounterLimiter(limit int, window time.Duration) (*SlidingWindowCounterLimiter, error) {
	if _, err := normalizeInput(limit, window, "bootstrap", 1); err != nil {
		return nil, err
	}
	return &SlidingWindowCounterLimiter{
		limit:   limit,
		window:  window,
		nowFunc: time.Now,
		counts:  make(map[string]slidingCounterState),
	}, nil
}

func (l *SlidingWindowCounterLimiter) ImplName() scene.ImplName {
	return ratelimiter.Lens.ImplName("Limiter", "sliding-window-counter")
}

func (l *SlidingWindowCounterLimiter) Allow(ctx context.Context, key string) (ratelimiter.Decision, error) {
	return l.AllowN(ctx, key, 1)
}

func (l *SlidingWindowCounterLimiter) AllowN(_ context.Context, key string, n int) (ratelimiter.Decision, error) {
	trimmed, err := normalizeInput(l.limit, l.window, key, n)
	if err != nil {
		return ratelimiter.Decision{}, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.nowFunc()
	windowNs := l.window.Nanoseconds()
	windowID := now.UnixNano() / windowNs

	state, ok := l.counts[trimmed]
	if !ok {
		state = slidingCounterState{windowID: windowID}
	}

	if windowID > state.windowID {
		shift := windowID - state.windowID
		if shift == 1 {
			state.prevCount = state.currCount
		} else {
			state.prevCount = 0
		}
		state.currCount = 0
		state.windowID = windowID
	}

	windowStart := time.Unix(0, windowID*windowNs)
	elapsed := now.Sub(windowStart)
	overlapRatio := 1 - (float64(elapsed) / float64(l.window))
	if overlapRatio < 0 {
		overlapRatio = 0
	}

	estimated := float64(state.currCount) + float64(state.prevCount)*overlapRatio
	requested := float64(n)
	if estimated+requested > float64(l.limit) {
		nextWindowAt := windowStart.Add(l.window)
		retryAfter := nextWindowAt.Sub(now)
		if retryAfter < 0 {
			retryAfter = 0
		}
		remaining := int(math.Floor(float64(l.limit) - estimated))
		if remaining < 0 {
			remaining = 0
		}
		l.counts[trimmed] = state
		return ratelimiter.Decision{
			Allowed:    false,
			Limit:      l.limit,
			Remaining:  remaining,
			RetryAfter: retryAfter,
		}, nil
	}

	state.currCount += n
	l.counts[trimmed] = state

	newEstimated := float64(state.currCount) + float64(state.prevCount)*overlapRatio
	remaining := int(math.Floor(float64(l.limit) - newEstimated))
	if remaining < 0 {
		remaining = 0
	}

	return ratelimiter.Decision{
		Allowed:    true,
		Limit:      l.limit,
		Remaining:  remaining,
		RetryAfter: 0,
	}, nil
}
