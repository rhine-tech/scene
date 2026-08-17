package registry

import (
	"reflect"
	"unsafe"
)

// core logic

const InjectTag = "aperture"
const EmbedValue = "embed"
const OptionalValue = "optional"

type injectionField struct {
	name      string
	optional  bool
	object    reflect.Value
	field     reflect.Value
	fieldName string
}

func inject(container *Container, fields []injectionField, hooks []InjectHookFunc) {
	for _, entry := range fields {
		// if field is nil, and it's an Interface or Ptr, inject it.
		if !entry.field.IsNil() {
			continue
		}
		instance, exists := container.lookup(entry.name)
		// if not exists, we have to check if this field is optional or not
		if !exists {
			// if this inject is optional, continue without panic
			if entry.optional {
				continue
			}
			// default should panic if optional tag is not specified.
			panic("scene registry: no instance found for " + entry.name + " when injecting " + entry.fieldName)
		}

		// run hooks
		for _, hook := range hooks {
			hook(entry.name, entry.object, entry.field, &instance)
		}
		setUnexportedField(entry.field, instance)
	}
}

func walkInjectionFields(indirectVal reflect.Value) []injectionField {
	if indirectVal.Kind() != reflect.Struct {
		panic("scene registry: inject on not injectable " + indirectVal.Type().String())
	}
	fields := make([]injectionField, 0)
	typ := indirectVal.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if tagValue, ok := field.Tag.Lookup(InjectTag); ok {
			fieldVal := indirectVal.Field(i)
			// if field is nil, and it's an Interface or Ptr, inject it.
			if tagValue != EmbedValue && (fieldVal.Kind() == reflect.Interface || fieldVal.Kind() == reflect.Ptr) && fieldVal.IsNil() {
				//fmt.Println("injecting", field.Type.String(), "for", getInterfaceName[T]())
				var lookupName string
				// set lookup name if tagValue is empty or specified as optional
				if tagValue == "" || tagValue == OptionalValue {
					lookupName = field.Type.String()
				} else {
					lookupName = tagValue
				}
				fields = append(fields, injectionField{
					name:      lookupName,
					optional:  tagValue == OptionalValue,
					object:    indirectVal.Addr(),
					field:     fieldVal,
					fieldName: field.Name,
				})
				continue
			}
			// if field is Anonymous field and has the tag or has the embed tag, inject the embed field
			if field.Anonymous || tagValue == EmbedValue {
				//fmt.Println("injecting embed", field.Type.String(), "for", getInterfaceName[T]())
				// The value to recurse into is the field we just processed.
				targetForRecursion := fieldVal
				// make targetForRecursion modifiable unexport to export
				targetForRecursion = reflect.NewAt(fieldVal.Type(), unsafe.Pointer(fieldVal.UnsafeAddr())).Elem()

				// if it is an interface, get value of the interface
				if targetForRecursion.Kind() == reflect.Interface {
					if targetForRecursion.Kind() == reflect.Interface {
						if targetForRecursion.IsNil() {
							panic("scene registry: failed to inject into a nil interface when injecting " + field.Name)
						}
						targetForRecursion = targetForRecursion.Elem()
					}
				}
				// If it's a pointer, we need to get the element it points to.
				// TODO: Maybe add support for dereferencing multiple pointer levels.
				if targetForRecursion.Kind() == reflect.Ptr {
					if targetForRecursion.IsNil() {
						// This can happen if the DI failed to find an instance,
						// or if it was nil to begin with and not injected.
						// In this case, we cannot recurse.
						panic("scene registry: failed to inject into a nil struct when injecting " + field.Name)
					}
					targetForRecursion = targetForRecursion.Elem()
				}
				fields = append(fields, walkInjectionFields(targetForRecursion)...)
			}
		}
	}
	return fields
}

func requirementsFromInjectionFields(fields []injectionField) []requirement {
	requirements := make([]requirement, 0, len(fields))
	seen := make(map[requirement]struct{}, len(fields))
	for _, field := range fields {
		candidate := requirement{
			name:     field.name,
			optional: field.optional,
		}
		if _, exists := seen[candidate]; exists {
			continue
		}
		seen[candidate] = struct{}{}
		requirements = append(requirements, candidate)
	}
	return requirements
}

func injectableStruct(value reflect.Value) (reflect.Value, bool) {
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Ptr) {
		if value.IsNil() {
			return reflect.Value{}, false
		}
		value = value.Elem()
	}
	return value, value.IsValid() && value.Kind() == reflect.Struct
}
