package registry

import (
	"reflect"
)

// InjectHookFunc runs during dependency injection when a field is about to be set.
// It is NOT called during registration.
//
// Parameters:
//   - iface: the lookup key/name used for this injection (usually interface type string or custom tag)
//   - obj:   the target object currently being injected (addressable reflect.Value)
//   - field: the target field reflect.Value that will receive the dependency
//   - instance: pointer to the resolved dependency instance; hook may replace it before field assignment
//
// If no dependency was resolved, this hook is not executed.
type InjectHookFunc func(
	iface string,
	obj reflect.Value,
	field reflect.Value,
	instance *interface{})

type InjectHook struct {
	Interface string
	Hook      InjectHookFunc
}

func RegisterInjectHooks(hooks ...InjectHook) {
	defaultScope.RegisterInjectHooks(hooks...)
}

func (s *Scope) RegisterInjectHooks(hooks ...InjectHook) {
	for _, hook := range hooks {
		s.RegisterInjectHookFunc(hook.Interface, hook.Hook)
	}
}

func RegisterInjectHookFunc(ifaceName string, hook InjectHookFunc) {
	defaultScope.RegisterInjectHookFunc(ifaceName, hook)
}

func (s *Scope) RegisterInjectHookFunc(ifaceName string, hook InjectHookFunc) {
	s.lock.Lock()
	s.hooks[ifaceName] = append(s.hooks[ifaceName], hook)
	s.lock.Unlock()
}

func (s *Scope) runInjectHooks(ifaceName string, obj reflect.Value, field reflect.Value, instance *interface{}) {
	for _, hook := range s.injectionHooks(ifaceName) {
		hook(ifaceName, obj, field, instance)
	}
}
