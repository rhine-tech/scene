package registry

func Register[T any](val T) T {
	for _, registrant := range registrants {
		registrant(val)
	}
	RegisterSingleton[T](val)
	TryInject(val)
	return val
}

func MustRegister[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return Register(val)
}

// Use is a shortcut for AcquireSingleton[T](val)
func Use[T any](val T) T {
	return AcquireSingleton[T](val)
}

// Load is a shortcut for TryInject(val)
func Load[T any](val T) T {
	return TryInject(val)
}

func Validate() {
	AcquireInfrastructure()
}
