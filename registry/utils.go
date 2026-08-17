package registry

import (
	"reflect"
	"unsafe"
)

func getInterfaceName[T any]() string {
	return reflect.TypeOf(new(T)).Elem().String()
}

func dependencyName[T any](name []string) string {
	switch len(name) {
	case 0:
		return getInterfaceName[T]()
	case 1:
		if name[0] == "" {
			return getInterfaceName[T]()
		}
		return name[0]
	default:
		panic("scene registry: dependency accepts at most one name")
	}
}

// https://stackoverflow.com/questions/42664837/how-to-access-unexported-struct-fields
func setUnexportedField(field reflect.Value, value interface{}) {
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
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

type instanceKey struct {
	typeOf  reflect.Type
	pointer uintptr
}

func identityOf(value any) (instanceKey, bool) {
	reflected := reflect.ValueOf(value)
	if !reflected.IsValid() {
		return instanceKey{}, false
	}
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice:
		if reflected.IsNil() {
			return instanceKey{}, false
		}
		return instanceKey{typeOf: reflected.Type(), pointer: reflected.Pointer()}, true
	default:
		return instanceKey{}, false
	}
}

func isNil(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
