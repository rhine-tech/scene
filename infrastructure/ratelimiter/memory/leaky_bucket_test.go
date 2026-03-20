package memory

import (
	"context"
	"testing"
	"time"

	"github.com/rhine-tech/scene/infrastructure/ratelimiter"
)

func TestNewLeakyBucketLimiterConfigValidation(t *testing.T) {
	if _, err := NewLeakyBucketLimiter(0, time.Second); err != ratelimiter.ErrInvalidLimit {
		t.Fatalf("expected ErrInvalidLimit, got %v", err)
	}
	if _, err := NewLeakyBucketLimiter(1, 0); err != ratelimiter.ErrInvalidWindow {
		t.Fatalf("expected ErrInvalidWindow, got %v", err)
	}
}

func TestLeakyBucketLimiterAllowAndLeak(t *testing.T) {
	l, err := NewLeakyBucketLimiter(2, time.Second)
	if err != nil {
		t.Fatalf("create leaky bucket failed: %v", err)
	}
	base := time.Unix(1700000100, 0)
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

	l.nowFunc = func() time.Time { return base.Add(600 * time.Millisecond) }
	d, e = l.Allow(context.Background(), "k")
	if e != nil || !d.Allowed {
		t.Fatalf("after leak request should pass, decision=%+v err=%v", d, e)
	}
}
