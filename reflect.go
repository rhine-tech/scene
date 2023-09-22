package scene

import "reflect"

func GetInterfaceName[T any]() string {
	return reflect.TypeOf(new(T)).Elem().String()
}
