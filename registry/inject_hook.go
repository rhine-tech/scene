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

func runHooks(ifaceName string, obj reflect.Value, field reflect.Value, iface string, instance *interface{}) {
	ifaceHooks, ok := hooks[ifaceName]
	if !ok {
		return
	}
	for _, hook := range ifaceHooks {
		hook(obj, field, iface, instance)
	}
	return
}

// InjectHookFunc is a function that is called when a new interface is registered.
// field is the field that is being injected, and instance is the instance that
// is being injected, instance can be nil. You can set instance to a new value
type InjectHookFunc func(
	obj reflect.Value,
	field reflect.Value,
	iface string,
	instance *interface{})

type InjectHook struct {
	Interface string
	Hook      InjectHookFunc
}
