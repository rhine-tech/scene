package registry

import "reflect"

func RegisterInjectHooks(hooks ...InjectHook) {
	defaultContainer.RegisterInjectHooks(hooks...)
}

func (container *Container) RegisterInjectHooks(hooks ...InjectHook) {
	for _, hook := range hooks {
		container.RegisterInjectHookFunc(hook.Interface, hook.Hook)
	}
}

func RegisterInjectHookFunc(ifaceName string, hook InjectHookFunc) {
	defaultContainer.RegisterInjectHookFunc(ifaceName, hook)
}

func (container *Container) RegisterInjectHookFunc(ifaceName string, hook InjectHookFunc) {
	container.lock.Lock()
	if _, ok := container.hooks[ifaceName]; !ok {
		container.hooks[ifaceName] = make([]InjectHookFunc, 0)
	}
	container.hooks[ifaceName] = append(container.hooks[ifaceName], hook)
	container.lock.Unlock()
}

func runHooks(container *Container, ifaceName string, obj reflect.Value, field reflect.Value, instance *interface{}) {
	container.lock.RLock()
	ifaceHooks, ok := container.hooks[ifaceName]
	ifaceHooks = append([]InjectHookFunc(nil), ifaceHooks...)
	container.lock.RUnlock()
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
