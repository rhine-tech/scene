package memory

import (
	"strings"
	"time"

	"github.com/rhine-tech/scene/infrastructure/ratelimiter"
)

func normalizeInput(limit int, window time.Duration, key string, n int) (string, error) {
	if limit <= 0 {
		return "", ratelimiter.ErrInvalidLimit
	}
	if window <= 0 {
		return "", ratelimiter.ErrInvalidWindow
	}
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return "", ratelimiter.ErrEmptyKey
	}
	if n <= 0 || n > limit {
		return "", ratelimiter.ErrInvalidN
	}
	return trimmed, nil
}

