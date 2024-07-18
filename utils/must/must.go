package must

func Must[T any](val T, err error) T {
	return val
}

func Must3[T1 any, T2 any](val1 T1, val2 T2, err error) (T1, T2) {
	return val1, val2
}

func MPanic[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}
