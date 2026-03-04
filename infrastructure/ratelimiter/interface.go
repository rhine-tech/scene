package ratelimiter

import (
	"context"
	"github.com/rhine-tech/scene"
	"time"
)

const Lens scene.InfraName = "ratelimiter"

// Limiter is a verbose and implementation-agnostic rate limiter contract.
type Limiter interface {
	scene.Named
	Allow(ctx context.Context, key string) (Decision, error)
	AllowN(ctx context.Context, key string, n int) (Decision, error)
}

// Decision is the verbose result of a rate limit attempt.
type Decision struct {
	Allowed bool

	// Limit is the quota limit of the current policy.
	Limit int
	// Remaining is the estimated units left after this decision.
	Remaining int

	// RetryAfter indicates how long caller should wait before retrying.
	// It is zero when Allowed is true.
	RetryAfter time.Duration
}
