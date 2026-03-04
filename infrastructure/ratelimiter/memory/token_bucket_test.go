package memory

import (
	"context"
	"testing"
	"time"

	"github.com/rhine-tech/scene/infrastructure/ratelimiter"
)

func TestNewTokenBucketLimiterConfigValidation(t *testing.T) {
	if _, err := NewTokenBucketLimiter(0, time.Second); err != ratelimiter.ErrInvalidLimit {
		t.Fatalf("expected ErrInvalidLimit, got %v", err)
	}
	if _, err := NewTokenBucketLimiter(1, 0); err != ratelimiter.ErrInvalidWindow {
		t.Fatalf("expected ErrInvalidWindow, got %v", err)
	}
}

func TestTokenBucketLimiterAllowAndRefill(t *testing.T) {
	l, err := NewTokenBucketLimiter(2, time.Second)
	if err != nil {
		t.Fatalf("create token bucket failed: %v", err)
	}
	base := time.Unix(1700000000, 0)
	l.nowFunc = func() time.Time { return base }

	for i := 0; i < 2; i++ {
		d, e := l.Allow(context.Background(), "k")
		if e != nil || !d.Allowed {
			t.Fatalf("request %d should pass, decision=%+v err=%v", i+1, d, e)
		}
	}

	d, e := l.Allow(context.Background(), "k")
	if e != nil || d.Allowed {
		t.Fatalf("third request should be blocked, decision=%+v err=%v", d, e)
	}
	if d.RetryAfter <= 0 {
		t.Fatalf("blocked decision should have retry_after, got %+v", d)
	}

	l.nowFunc = func() time.Time { return base.Add(600 * time.Millisecond) }
	d, e = l.Allow(context.Background(), "k")
	if e != nil || !d.Allowed {
		t.Fatalf("after refill request should pass, decision=%+v err=%v", d, e)
	}
}

