package utils

import (
	"golang.org/x/exp/constraints"
)

func CopySlice[T any](s []T) []T {
	c := make([]T, len(s))
	copy(c, s)
	return c
}

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}
