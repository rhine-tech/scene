package must

func Must[T any](val T, err error) T {
	return val
}

func Must3[T1 any, T2 any](val1 T1, val2 T2, err error) (T1, T2) {
	return val1, val2
}

// PMust panic if err is not nil
func PMust[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

// PMust3 panic if err is not nil
func PMust3[T1 any, T2 any](val1 T1, val2 T2, err error) (T1, T2) {
	if err != nil {
		panic(err)
	}
	return val1, val2
}
