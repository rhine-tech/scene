package registry

import (
	"reflect"
	"unsafe"
)

func getInterfaceName[T any]() string {
	return reflect.TypeOf(new(T)).Elem().String()
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
