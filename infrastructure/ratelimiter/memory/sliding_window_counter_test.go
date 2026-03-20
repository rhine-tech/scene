package memory

import (
	"context"
	"testing"
	"time"

	"github.com/rhine-tech/scene/infrastructure/ratelimiter"
)

func TestNewSlidingWindowCounterLimiterConfigValidation(t *testing.T) {
	if _, err := NewSlidingWindowCounterLimiter(0, time.Second); err != ratelimiter.ErrInvalidLimit {
		t.Fatalf("expected ErrInvalidLimit, got %v", err)
	}
	if _, err := NewSlidingWindowCounterLimiter(1, 0); err != ratelimiter.ErrInvalidWindow {
		t.Fatalf("expected ErrInvalidWindow, got %v", err)
	}
}

func TestSlidingWindowCounterLimiterAllowAndShift(t *testing.T) {
	l, err := NewSlidingWindowCounterLimiter(2, time.Second)
	if err != nil {
		t.Fatalf("create sliding counter failed: %v", err)
	}
	base := time.Unix(1700000400, 0)
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

	l.nowFunc = func() time.Time { return base.Add(2 * time.Second) }
	d, e = l.Allow(context.Background(), "k")
	if e != nil || !d.Allowed {
		t.Fatalf("new window request should pass, decision=%+v err=%v", d, e)
	}
}
