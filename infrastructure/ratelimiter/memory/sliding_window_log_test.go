package memory

import (
	"context"
	"testing"
	"time"

	"github.com/rhine-tech/scene/infrastructure/ratelimiter"
)

func TestNewSlidingWindowLogLimiterConfigValidation(t *testing.T) {
	if _, err := NewSlidingWindowLogLimiter(0, time.Second); err != ratelimiter.ErrInvalidLimit {
		t.Fatalf("expected ErrInvalidLimit, got %v", err)
	}
	if _, err := NewSlidingWindowLogLimiter(1, 0); err != ratelimiter.ErrInvalidWindow {
		t.Fatalf("expected ErrInvalidWindow, got %v", err)
	}
}

func TestSlidingWindowLogLimiterAllow(t *testing.T) {
	limiter, err := NewSlidingWindowLogLimiter(2, time.Second)
	if err != nil {
		t.Fatalf("create limiter failed: %v", err)
	}

	got, err := limiter.Allow(context.Background(), "u:1")
	if err != nil || !got.Allowed {
		t.Fatalf("first allow should pass, decision=%+v err=%v", got, err)
	}
	got, err = limiter.Allow(context.Background(), "u:1")
	if err != nil || !got.Allowed {
		t.Fatalf("second allow should pass, decision=%+v err=%v", got, err)
	}
	got, err = limiter.Allow(context.Background(), "u:1")
	if err != nil {
		t.Fatalf("third allow should not return error, got %v", err)
	}
	if got.Allowed {
		t.Fatalf("third allow should be blocked")
	}
	if got.RetryAfter <= 0 {
		t.Fatalf("blocked decision should provide retry_after, got %+v", got)
	}
}

func TestSlidingWindowLogLimiterWindowReset(t *testing.T) {
	base := time.Unix(1700000000, 0)
	limiter, err := NewSlidingWindowLogLimiter(1, time.Second)
	if err != nil {
		t.Fatalf("create limiter failed: %v", err)
	}
	limiter.nowFunc = func() time.Time { return base }

	got, err := limiter.Allow(context.Background(), "k")
	if err != nil || !got.Allowed {
		t.Fatalf("first allow should pass, decision=%+v err=%v", got, err)
	}

	got, err = limiter.Allow(context.Background(), "k")
	if err != nil {
		t.Fatalf("second allow should not return error, got %v", err)
	}
	if got.Allowed {
		t.Fatalf("second allow should be blocked in same window")
	}

	limiter.nowFunc = func() time.Time { return base.Add(2 * time.Second) }
	got, err = limiter.Allow(context.Background(), "k")
	if err != nil || !got.Allowed {
		t.Fatalf("allow after window should pass, decision=%+v err=%v", got, err)
	}
}

func TestSlidingWindowLogLimiterPerKeyIsolation(t *testing.T) {
	limiter, err := NewSlidingWindowLogLimiter(1, time.Second)
	if err != nil {
		t.Fatalf("create limiter failed: %v", err)
	}

	got, err := limiter.Allow(context.Background(), "user:a")
	if err != nil || !got.Allowed {
		t.Fatalf("user:a first should pass, decision=%+v err=%v", got, err)
	}
	got, err = limiter.Allow(context.Background(), "user:b")
	if err != nil || !got.Allowed {
		t.Fatalf("user:b first should pass, decision=%+v err=%v", got, err)
	}
}

func TestSlidingWindowLogLimiterEmptyKey(t *testing.T) {
	limiter, err := NewSlidingWindowLogLimiter(1, time.Second)
	if err != nil {
		t.Fatalf("create limiter failed: %v", err)
	}

	if _, err = limiter.Allow(context.Background(), "   "); err != ratelimiter.ErrEmptyKey {
		t.Fatalf("expected ErrEmptyKey, got %v", err)
	}
}

func TestSlidingWindowLogLimiterAmount(t *testing.T) {
	limiter, err := NewSlidingWindowLogLimiter(3, time.Second)
	if err != nil {
		t.Fatalf("create limiter failed: %v", err)
	}

	got, err := limiter.AllowN(context.Background(), "k", 2)
	if err != nil || !got.Allowed {
		t.Fatalf("amount consume should pass, decision=%+v err=%v", got, err)
	}
	if got.Remaining != 1 {
		t.Fatalf("expected remaining=1, got %d", got.Remaining)
	}

	got, err = limiter.AllowN(context.Background(), "k", 2)
	if err != nil {
		t.Fatalf("second amount consume should not return error, got %v", err)
	}
	if got.Allowed {
		t.Fatalf("second amount consume should be blocked")
	}
}

func TestSlidingWindowLogLimiterInvalidN(t *testing.T) {
	limiter, err := NewSlidingWindowLogLimiter(2, time.Second)
	if err != nil {
		t.Fatalf("create limiter failed: %v", err)
	}

	if _, err = limiter.AllowN(context.Background(), "k", 0); err != ratelimiter.ErrInvalidN {
		t.Fatalf("expected ErrInvalidN when n=0, got %v", err)
	}
	if _, err = limiter.AllowN(context.Background(), "k", 3); err != ratelimiter.ErrInvalidN {
		t.Fatalf("expected ErrInvalidN when n>limit, got %v", err)
	}
}
