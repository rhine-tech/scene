package registry

import "reflect"

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

// Use is a function to get instance from registry
// if val is not nil, then return itself
// if val is nil, then return the instance from registry
func Use[T any](val T) T {
	if canUse(val) {
		return val
	}
	return AcquireSingleton[T](val)
}

func canUse[T any](val T) bool {
	rv := reflect.ValueOf(val)
	actualKind := reflect.TypeOf(new(T)).Elem().Kind()
	if actualKind == reflect.Struct {
		return !rv.IsZero()
	}
	if actualKind == reflect.Interface {
		return rv.IsValid()
	}
	if actualKind == reflect.Ptr {
		return !rv.IsNil()
	}
	return rv.IsValid()
	//if actualVal.Kind() == reflect.Interface || actualVal.Kind() == reflect.Ptr {
	//	return rv.IsValid()
	//}
	//if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Slice || rv.Kind() == reflect.Map || rv.Kind() == reflect.Chan || rv.Kind() == reflect.Func {
	//	if rv.IsNil() {
	//		return false
	//	}
	//}
	//return rv.IsValid()
}

// Load is a shortcut for TryInject(val)
// Load will also check if val can add to registry
func Load[T any](val T) T {
	for _, registrant := range registrants {
		registrant(val)
	}
	return TryInject(val)
}

// Inject is the shortcut for TryInject(val)
func Inject[T any](val T) T {
	return TryInject(val)
}

func Validate() {
	AcquireInfrastructure()
}
