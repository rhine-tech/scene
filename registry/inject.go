package registry

import (
	"reflect"
)

var lazyLoads []any

var UseLazyInject = false

const InjectTag = "aperture"

func TryInject[T any](injectable T) T {
	val := reflect.ValueOf(injectable)
	indirectVal := reflect.Indirect(val) // In case injectable is a pointer
	if indirectVal.Kind() != reflect.Struct {
		// not injectable
		return injectable
	}
	typ := indirectVal.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if tagValue, ok := field.Tag.Lookup(InjectTag); ok {
			fieldVal := indirectVal.Field(i)
			// if field is nil, inject it. otherwise keep the value
			// field value can only be Interface or Ptr
			if (fieldVal.Kind() == reflect.Interface || fieldVal.Kind() == reflect.Ptr) && fieldVal.IsNil() {
				//fmt.Println("injecting", field.Type.String(), "for", getInterfaceName[T]())
				if tagValue == "" {
					tagValue = field.Type.String()
				}
				instance, exists := singletonRegistry[tagValue]
				if !exists {
					panic("scene registry: no instance found for " + tagValue + " when injecting " + field.Name)
				}
				setUnexportedField(fieldVal, instance)
			}
		}
	}
	return injectable
}

func LazyInject() {
	for _, val := range lazyLoads {
		TryInject(val)
	}
}

// WithLazyInjection will do delay inject until all interface registered
// proc is the function register all interface
func WithLazyInjection(proc func()) {
	UseLazyInject = true
	proc()
	for _, val := range lazyLoads {
		TryInject(val)
	}
	UseLazyInject = false
	lazyLoads = make([]any, 0)
}
