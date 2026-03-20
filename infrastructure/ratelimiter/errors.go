package ratelimiter

import "errors"

var (
	ErrInvalidLimit  = errors.New("ratelimiter: limit must be greater than zero")
	ErrInvalidWindow = errors.New("ratelimiter: window must be greater than zero")
	ErrEmptyKey      = errors.New("ratelimiter: key cannot be empty")
	ErrInvalidN      = errors.New("ratelimiter: n must be greater than zero and less than or equal to limit")
)
