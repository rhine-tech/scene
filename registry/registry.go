package registry

func Register[T any](val T, name ...string) T {
	return ContainerRegister(defaultContainer, val, name...)
}

func MustRegister[T any](val T, err error, name ...string) T {
	return ContainerMustRegister(defaultContainer, val, err, name...)
}

func Provide[T any](name ...string) T {
	return ContainerProvide[T](defaultContainer, name...)
}

// Use is a function to get instance from registry
// if val is not nil, then return itself
// if val is nil, then return the instance from registry
func Use[T any](val T) T {
	return ContainerUse(defaultContainer, val)
}

// Load is a shortcut for Inject(val)
// Load will also check if val can add to registry
func Load[T any](val T) T {
	return ContainerLoad(defaultContainer, val)
}

// Inject injects dependencies from the default container into val.
func Inject[T any](val T) T {
	return ContainerInject(defaultContainer, val)
}

func Validate() {
	AcquireInfrastructure()
}
