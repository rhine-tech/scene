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
	dep hookTestIface `aperture:"registry.hookTestIface"`
}

type hookNoRegistrationHolder struct {
	dep hookTestIface `aperture:""`
}

func resetInjectHooksForTest() {
	hooks = make(map[string][]InjectHookFunc)
}

func TestTryInject_Hook_OptionalTag_WithRegisteredInstance_ShouldRunHook(t *testing.T) {
	resetSingletonRegistryForTest()
	resetInjectHooksForTest()

	RegisterSingleton[hookTestIface](&hookTestImpl{v: "origin"})
	RegisterInjectHookFunc(reflect.TypeOf((*hookTestIface)(nil)).Elem().String(), func(obj reflect.Value, field reflect.Value, iface string, instance *interface{}) {
		*instance = hookTestIface(&hookTestImpl{v: "hooked"})
	})

	holder := hookOptionalHolder{}
	TryInject(&holder)

	require.NotNil(t, holder.dep)
	// This assertion reproduces the bug: current inject.go calls runHooks with tag value "optional",
	// so hook key "registry.hookTestIface" is not hit and value remains "origin".
	require.Equal(t, "hooked", holder.dep.Value())
}

func TestTryInject_Hook_ExplicitTag_WithRegisteredInstance_ShouldRunHook(t *testing.T) {
	resetSingletonRegistryForTest()
	resetInjectHooksForTest()

	RegisterSingleton[hookTestIface](&hookTestImpl{v: "origin"})
	RegisterInjectHookFunc("registry.hookTestIface", func(obj reflect.Value, field reflect.Value, iface string, instance *interface{}) {
		*instance = hookTestIface(&hookTestImpl{v: "hooked"})
	})

	holder := hookExplicitHolder{}
	TryInject(&holder)

	require.NotNil(t, holder.dep)
	require.Equal(t, "hooked", holder.dep.Value())
}

func TestTryInject_Hook_NoHookRegistered_ShouldKeepInjectedInstance(t *testing.T) {
	resetSingletonRegistryForTest()
	resetInjectHooksForTest()

	RegisterSingleton[hookTestIface](&hookTestImpl{v: "origin"})

	holder := hookNoRegistrationHolder{}
	TryInject(&holder)

	require.NotNil(t, holder.dep)
	require.Equal(t, "origin", holder.dep.Value())
}
