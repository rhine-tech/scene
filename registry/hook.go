package registry

import "reflect"

var hooks = make(map[string][]InjectHookFunc)

func RegisterInjectHooks(hooks ...InjectHook) {
	for _, hook := range hooks {
		RegisterInjectHookFunc(hook.Interface, hook.Hook)
	}
}

func RegisterInjectHookFunc(ifaceName string, hook InjectHookFunc) {
	if _, ok := hooks[ifaceName]; !ok {
		hooks[ifaceName] = make([]InjectHookFunc, 0)
	}
	hooks[ifaceName] = append(hooks[ifaceName], hook)
}

func runHooks(ifaceName string, obj reflect.Value, field reflect.Value, instance *interface{}) {
	ifaceHooks, ok := hooks[ifaceName]
	if !ok {
		return
	}
	for _, hook := range ifaceHooks {
		hook(obj, field, ifaceName, instance)
	}
	return
}

// InjectHookFunc runs during dependency injection when a field is about to be set.
// It is NOT called during registration.
//
// Parameters:
//   - obj:   the target object currently being injected (addressable reflect.Value)
//   - field: the target field reflect.Value that will receive the dependency
//   - iface: the lookup key/name used for this injection (usually interface type string or custom tag)
//   - instance: pointer to the resolved dependency instance; hook may replace it before field assignment
//
// If no dependency was resolved, this hook is not executed.
type InjectHookFunc func(
	obj reflect.Value,
	field reflect.Value,
	iface string,
	instance *interface{})

type InjectHook struct {
	Interface string
	Hook      InjectHookFunc
}
