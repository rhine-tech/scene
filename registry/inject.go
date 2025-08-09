package registry

import (
	"reflect"
)

var lazyLoads []any

var UseLazyInject = false

const InjectTag = "aperture"
const EmbedValue = "embed"

func inject[T any](indirectVal reflect.Value) {
	if indirectVal.Kind() != reflect.Struct {
		panic("scene registry: inject on not injectable " + indirectVal.Type().String())
		return
	}
	typ := indirectVal.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if tagValue, ok := field.Tag.Lookup(InjectTag); ok {
			fieldVal := indirectVal.Field(i)
			// if field is nil, and it's an Interface or Ptr, inject it.
			if tagValue != EmbedValue && (fieldVal.Kind() == reflect.Interface || fieldVal.Kind() == reflect.Ptr) && fieldVal.IsNil() {
				//fmt.Println("injecting", field.Type.String(), "for", getInterfaceName[T]())
				if tagValue == "" {
					tagValue = field.Type.String()
				}
				instance, exists := singletonRegistry[tagValue]
				if !exists {
					panic("scene registry: no instance found for " + tagValue + " when injecting " + field.Name)
				}
				// run hooks
				runHooks(tagValue, indirectVal.Addr(), fieldVal, tagValue, &instance)
				setUnexportedField(fieldVal, instance)
				continue
			}
			// if field is Anonymous field and has the tag or has the embed tag, inject the embed field
			if field.Anonymous || tagValue == EmbedValue {
				// fmt.Println("injecting embed", field.Type.String(), "for", getInterfaceName[T]())
				// The value to recurse into is the field we just processed.
				targetForRecursion := fieldVal

				// if it is an interface, get value of the interface
				if targetForRecursion.Kind() == reflect.Interface {
					targetForRecursion = targetForRecursion.Elem()
				}

				// If it's a pointer, we need to get the element it points to.
				if targetForRecursion.Kind() == reflect.Ptr {
					if targetForRecursion.IsNil() {
						// This can happen if the DI failed to find an instance,
						// or if it was nil to begin with and not injected.
						// In this case, we cannot recurse.
						panic("scene registry: failed to inject into a nil struct when injecting " + field.Name)
					}
					targetForRecursion = targetForRecursion.Elem()
				}
				// Finally, make the recursive call on the embedded struct value.
				inject[T](targetForRecursion)
			}
		}
	}
	return
}

func TryInject[T any](injectable T) T {
	val := reflect.ValueOf(injectable)
	indirectVal := reflect.Indirect(val) // In case injectable is a pointer
	inject[T](indirectVal)
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
