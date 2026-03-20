package memory

import (
	"context"
	"testing"
	"time"

	"github.com/rhine-tech/scene/infrastructure/ratelimiter"
)

func TestNewTimeRateLimiterConfigValidation(t *testing.T) {
	if _, err := NewTimeRateLimiter(0, time.Second); err != ratelimiter.ErrInvalidLimit {
		t.Fatalf("expected ErrInvalidLimit, got %v", err)
	}
	if _, err := NewTimeRateLimiter(1, 0); err != ratelimiter.ErrInvalidWindow {
		t.Fatalf("expected ErrInvalidWindow, got %v", err)
	}
}

func TestTimeRateLimiterAllowAndRefill(t *testing.T) {
	l, err := NewTimeRateLimiter(2, time.Second)
	if err != nil {
		t.Fatalf("create time rate limiter failed: %v", err)
	}
	base := time.Unix(1700000500, 0)
	now := base
	l.nowFunc = func() time.Time { return now }

	for i := 0; i < 2; i++ {
		d, e := l.Allow(context.Background(), "k")
		if e != nil || !d.Allowed {
			t.Fatalf("request %d should pass, decision=%+v err=%v", i+1, d, e)
		}
	}

	d, e := l.Allow(context.Background(), "k")
	if e != nil {
		t.Fatalf("third request should not return error, got %v", e)
	}
	if d.Allowed {
		t.Fatalf("third request should be blocked, decision=%+v", d)
	}
	if d.RetryAfter <= 0 {
		t.Fatalf("blocked decision should provide retry_after, got %+v", d)
	}

	now = base.Add(600 * time.Millisecond)
	d, e = l.Allow(context.Background(), "k")
	if e != nil || !d.Allowed {
		t.Fatalf("after refill request should pass, decision=%+v err=%v", d, e)
	}
}

func TestTimeRateLimiterPerKeyIsolation(t *testing.T) {
	l, err := NewTimeRateLimiter(1, time.Second)
	if err != nil {
		t.Fatalf("create time rate limiter failed: %v", err)
	}
	base := time.Unix(1700000600, 0)
	l.nowFunc = func() time.Time { return base }

	d, e := l.Allow(context.Background(), "user:a")
	if e != nil || !d.Allowed {
		t.Fatalf("user:a first should pass, decision=%+v err=%v", d, e)
	}
	d, e = l.Allow(context.Background(), "user:b")
	if e != nil || !d.Allowed {
		t.Fatalf("user:b first should pass, decision=%+v err=%v", d, e)
	}
}

func TestTimeRateLimiterInvalidNAndKey(t *testing.T) {
	l, err := NewTimeRateLimiter(2, time.Second)
	if err != nil {
		t.Fatalf("create time rate limiter failed: %v", err)
	}

	if _, err = l.Allow(context.Background(), " "); err != ratelimiter.ErrEmptyKey {
		t.Fatalf("expected ErrEmptyKey, got %v", err)
	}
	if _, err = l.AllowN(context.Background(), "k", 0); err != ratelimiter.ErrInvalidN {
		t.Fatalf("expected ErrInvalidN for n=0, got %v", err)
	}
	if _, err = l.AllowN(context.Background(), "k", 3); err != ratelimiter.ErrInvalidN {
		t.Fatalf("expected ErrInvalidN for n>limit, got %v", err)
	}
}
