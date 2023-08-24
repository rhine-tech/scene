package registry

import (
	"reflect"
)

func getInterfaceName[T any]() string {
	return reflect.TypeOf(new(T)).Elem().String()
}
