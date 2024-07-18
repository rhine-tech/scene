package utils

func Must[T any](val T, err error) T {
	return val
}
