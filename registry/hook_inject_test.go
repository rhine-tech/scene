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

func TestInject_Hook_OptionalTag_WithRegisteredInstance_ShouldRunHook(t *testing.T) {
	container := NewContainer("hook-optional")
	Register[hookTestIface](container, &hookTestImpl{v: "origin"})
	holder := hookOptionalHolder{}
	container.Load(&holder)

	scope := NewScope()
	scope.RegisterInjectHooks(InjectHook{
		Interface: reflect.TypeFor[hookTestIface]().String(),
		Hook: func(_ string, _ reflect.Value, _ reflect.Value, instance *interface{}) {
			*instance = hookTestIface(&hookTestImpl{v: "hooked"})
		},
	})

	require.NoError(t, scope.Build(container))
	require.NotNil(t, holder.dep)
	// Optional injection hooks use the interface type as their lookup key.
	require.Equal(t, "hooked", holder.dep.Value())
}

func TestInject_Hook_ExplicitTag_WithRegisteredInstance_ShouldRunHook(t *testing.T) {
	container := NewContainer("hook-explicit")
	Register[hookTestIface](container, &hookTestImpl{v: "origin"}, "hook-explicit")
	holder := hookExplicitHolder{}
	container.Load(&holder)

	scope := NewScope()
	scope.RegisterInjectHookFunc(
		"hook-explicit",
		func(_ string, _ reflect.Value, _ reflect.Value, instance *interface{}) {
			*instance = hookTestIface(&hookTestImpl{v: "hooked"})
		},
	)

	require.NoError(t, scope.Build(container))
	require.NotNil(t, holder.dep)
	require.Equal(t, "hooked", holder.dep.Value())
}

func TestInject_Hook_NoHookRegistered_ShouldKeepInjectedInstance(t *testing.T) {
	container := NewContainer("hook-missing")
	origin := &hookTestImpl{v: "origin"}
	Register[hookTestIface](container, origin)
	holder := hookNoRegistrationHolder{}
	container.Load(&holder)

	require.NoError(t, NewScope().Build(container))
	require.Same(t, origin, holder.dep)
	require.Equal(t, "origin", holder.dep.Value())
}

func TestInject_Hook_MissingOptionalDependencyDoesNotRunHook(t *testing.T) {
	container := NewContainer("hook-missing-optional")
	holder := hookOptionalHolder{}
	container.Load(&holder)

	called := false
	scope := NewScope()
	scope.RegisterInjectHookFunc(
		reflect.TypeFor[hookTestIface]().String(),
		func(_ string, _ reflect.Value, _ reflect.Value, _ *interface{}) {
			called = true
		},
	)

	require.NoError(t, scope.Build(container))
	require.False(t, called)
	require.Nil(t, holder.dep)
}

func TestContainerInjectRunsOnce(t *testing.T) {
	container := NewContainer("inject-once")
	Register[hookTestIface](container, &hookTestImpl{v: "origin"})
	holder := hookNoRegistrationHolder{}
	container.Load(&holder)

	called := 0
	hook := func(_ string, _ reflect.Value, _ reflect.Value, _ *interface{}) {
		called++
	}
	container.Inject(hook)
	container.Inject(hook)

	require.Equal(t, 1, called)
}

func TestScopeResetPreservesGlobalsAndHooks(t *testing.T) {
	scope := NewScope()
	global := &hookTestImpl{v: "global"}
	SetIn[hookTestIface](scope, global)

	hookCalls := 0
	scope.RegisterInjectHookFunc(
		reflect.TypeFor[hookTestIface]().String(),
		func(_ string, _ reflect.Value, _ reflect.Value, _ *interface{}) {
			hookCalls++
		},
	)

	build := func(name string) {
		container := NewContainer(name)
		Register[hookTestIface](container, &hookTestImpl{v: name})
		container.Load(&hookNoRegistrationHolder{})
		require.NoError(t, scope.Build(container))
	}

	build("first")
	scope.Reset()
	require.Same(t, global, ProvideIn[hookTestIface](scope))
	build("second")
	require.Equal(t, 2, hookCalls)
}
