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
