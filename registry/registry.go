package registry

import (
	"fmt"
)

// Set registers a process-level value used before a module graph is active.
// It is intended for global dependencies such as configuration.
func Set[T any](value T, name ...string) T {
	return SetIn(defaultScope, value, name...)
}

// SetIn registers a global value in scope.
func SetIn[T any](scope *Scope, value T, name ...string) T {
	scope.set(dependencyName[T](name), value)
	return value
}

func Provide[T any](name ...string) T {
	return ProvideIn[T](defaultScope, name...)
}

// ProvideIn resolves T from scope and panics when it is unavailable.
func ProvideIn[T any](scope *Scope, name ...string) T {
	value, ok := LookupIn[T](scope, name...)
	if ok {
		return value
	}
	panic(fmt.Sprintf("no dependency registered for %s", dependencyName[T](name)))
}

// Lookup resolves a dependency without panicking when it is unavailable.
func Lookup[T any](name ...string) (T, bool) {
	return LookupIn[T](defaultScope, name...)
}

// LookupIn resolves a dependency from scope without panicking when it is
// unavailable.
func LookupIn[T any](scope *Scope, name ...string) (T, bool) {
	key := dependencyName[T](name)
	if value, ok := scope.lookup(key); ok {
		return value.(T), true
	}
	var zero T
	return zero, false
}

// Use returns val when it is usable, otherwise it resolves T from the active
// module graph or the global values.
func Use[T any](val T) T {
	return UseIn(defaultScope, val)
}

// UseIn returns val when it is usable, otherwise it resolves T from scope.
func UseIn[T any](scope *Scope, val T) T {
	if canUse(val) {
		return val
	}
	return ProvideIn[T](scope)
}
