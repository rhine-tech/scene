package registry

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type hookTestIface interface {
	Value() string
}

type hookTestImpl struct {
	v string
}

func (h *hookTestImpl) Value() string {
	return h.v
}

type hookOptionalHolder struct {
	dep hookTestIface `aperture:"optional"`
}

type hookExplicitHolder struct {
	dep hookTestIface `aperture:"hook-explicit"`
}

type hookNoRegistrationHolder struct {
	dep hookTestIface `aperture:""`
}

func resetInjectHooksForTest() {
	defaultContainer.lock.Lock()
	defaultContainer.hooks = make(map[string][]InjectHookFunc)
	defaultContainer.lock.Unlock()
}

func TestInject_Hook_OptionalTag_WithRegisteredInstance_ShouldRunHook(t *testing.T) {
	resetDependenciesForTest()
	resetInjectHooksForTest()

	Register[hookTestIface](&hookTestImpl{v: "origin"})
	RegisterInjectHooks(InjectHook{
		Interface: reflect.TypeOf((*hookTestIface)(nil)).Elem().String(),
		Hook: func(obj reflect.Value, field reflect.Value, iface string, instance *interface{}) {
			*instance = hookTestIface(&hookTestImpl{v: "hooked"})
		},
	})

	holder := hookOptionalHolder{}
	Inject(&holder)

	require.NotNil(t, holder.dep)
	// Optional injection hooks use the interface type as their lookup key.
	require.Equal(t, "hooked", holder.dep.Value())
}

func TestInject_Hook_ExplicitTag_WithRegisteredInstance_ShouldRunHook(t *testing.T) {
	resetDependenciesForTest()
	resetInjectHooksForTest()

	Register[hookTestIface](&hookTestImpl{v: "origin"}, "hook-explicit")
	RegisterInjectHookFunc("hook-explicit", func(obj reflect.Value, field reflect.Value, iface string, instance *interface{}) {
		*instance = hookTestIface(&hookTestImpl{v: "hooked"})
	})

	holder := hookExplicitHolder{}
	Inject(&holder)

	require.NotNil(t, holder.dep)
	require.Equal(t, "hooked", holder.dep.Value())
}

func TestInject_Hook_NoHookRegistered_ShouldKeepInjectedInstance(t *testing.T) {
	resetDependenciesForTest()
	resetInjectHooksForTest()

	Register[hookTestIface](&hookTestImpl{v: "origin"})

	holder := hookNoRegistrationHolder{}
	Inject(&holder)

	require.NotNil(t, holder.dep)
	require.Equal(t, "origin", holder.dep.Value())
}
