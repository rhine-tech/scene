package registry

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type containerTestDependency interface {
	Value() string
}

type containerTestDependencyImpl struct {
	value string
}

func (d *containerTestDependencyImpl) Value() string {
	return d.value
}

type containerTestTarget struct {
	dependency containerTestDependency `aperture:""`
}

type containerNamedTestTarget struct {
	dependency containerTestDependency `aperture:"container-named"`
}

func TestContainerIsolatesDependencies(t *testing.T) {
	first := NewContainer()
	second := NewContainer()

	ContainerRegister[containerTestDependency](first, &containerTestDependencyImpl{value: "first"})
	ContainerRegister[containerTestDependency](second, &containerTestDependencyImpl{value: "second"})

	firstTarget := new(containerTestTarget)
	secondTarget := new(containerTestTarget)
	ContainerInject(first, firstTarget)
	ContainerInject(second, secondTarget)

	require.Equal(t, "first", firstTarget.dependency.Value())
	require.Equal(t, "second", secondTarget.dependency.Value())

	firstDependency := ContainerProvide[containerTestDependency](first)
	require.Equal(t, "first", firstDependency.Value())
}

func TestContainerIsolatesNamedDependencies(t *testing.T) {
	first := NewContainer()
	second := NewContainer()

	ContainerRegister[containerTestDependency](
		first,
		&containerTestDependencyImpl{value: "first"},
		"container-named",
	)
	ContainerRegister[containerTestDependency](
		second,
		&containerTestDependencyImpl{value: "second"},
		"container-named",
	)

	firstTarget := new(containerNamedTestTarget)
	secondTarget := new(containerNamedTestTarget)
	ContainerInject(first, firstTarget)
	ContainerInject(second, secondTarget)

	require.Equal(t, "first", firstTarget.dependency.Value())
	require.Equal(t, "second", secondTarget.dependency.Value())
	require.Equal(
		t,
		"first",
		ContainerProvide[containerTestDependency](first, "container-named").Value(),
	)
}

func TestContainerIsolatesInjectHooks(t *testing.T) {
	withHook := NewContainer()
	withoutHook := NewContainer()
	ContainerRegister[containerTestDependency](withHook, &containerTestDependencyImpl{value: "origin"})
	ContainerRegister[containerTestDependency](withoutHook, &containerTestDependencyImpl{value: "origin"})

	withHook.RegisterInjectHooks(
		InjectHook{
			Interface: getInterfaceName[containerTestDependency](),
			Hook: func(_ reflect.Value, _ reflect.Value, _ string, instance *interface{}) {
				*instance = containerTestDependency(&containerTestDependencyImpl{value: "hooked"})
			},
		},
	)

	withHookTarget := new(containerTestTarget)
	withoutHookTarget := new(containerTestTarget)
	ContainerInject(withHook, withHookTarget)
	ContainerInject(withoutHook, withoutHookTarget)

	require.Equal(t, "hooked", withHookTarget.dependency.Value())
	require.Equal(t, "origin", withoutHookTarget.dependency.Value())
}

func TestContainerLoadDoesNotRegisterDependency(t *testing.T) {
	container := NewContainer()

	ContainerLoad[containerTestDependency](
		container,
		&containerTestDependencyImpl{value: "loaded"},
	)

	require.Panics(t, func() {
		_ = ContainerProvide[containerTestDependency](container)
	})
}

func TestContainerWithLazyInjection(t *testing.T) {
	container := NewContainer()
	target := new(containerTestTarget)

	container.WithLazyInjection(func() {
		ContainerLoad(container, target)
		require.Nil(t, target.dependency)
		ContainerRegister[containerTestDependency](
			container,
			&containerTestDependencyImpl{value: "lazy"},
		)
	})

	require.Equal(t, "lazy", target.dependency.Value())
}

func TestWithLazyInjectionUsesDefaultContainer(t *testing.T) {
	resetDependenciesForTest()
	defaultContainer.lock.Lock()
	defaultContainer.pendingInjections = nil
	defaultContainer.lock.Unlock()
	target := new(containerTestTarget)

	WithLazyInjection(func() {
		Load(target)
		require.Nil(t, target.dependency)
		Register[containerTestDependency](
			&containerTestDependencyImpl{value: "default container"},
		)
	})

	require.Equal(t, "default container", target.dependency.Value())
}

func TestContainerWithLazyInjectionAbortsOnPanic(t *testing.T) {
	container := NewContainer()
	target := new(containerTestTarget)

	require.PanicsWithValue(t, "boom", func() {
		container.WithLazyInjection(func() {
			ContainerLoad(container, target)
			panic("boom")
		})
	})
	require.Nil(t, target.dependency)

	ContainerRegister[containerTestDependency](
		container,
		&containerTestDependencyImpl{value: "after panic"},
	)
	ContainerLoad(container, target)
	require.Equal(t, "after panic", target.dependency.Value())
}

func TestContainerWithLazyInjectionRejectsNestedScopes(t *testing.T) {
	container := NewContainer()

	require.PanicsWithValue(
		t,
		"scene registry: lazy injection is already active for this container",
		func() {
			container.WithLazyInjection(func() {
				container.WithLazyInjection(func() {})
			})
		},
	)

	target := new(containerTestTarget)
	ContainerRegister[containerTestDependency](
		container,
		&containerTestDependencyImpl{value: "after nested scope"},
	)
	ContainerLoad(container, target)
	require.Equal(t, "after nested scope", target.dependency.Value())
}

func TestContainerWithLazyInjectionFinishesBeforeInjecting(t *testing.T) {
	container := NewContainer()
	target := new(containerTestTarget)
	registeredByHook := new(containerTestTarget)
	hookCalled := false

	container.RegisterInjectHookFunc(
		getInterfaceName[containerTestDependency](),
		func(_ reflect.Value, _ reflect.Value, _ string, _ *interface{}) {
			if hookCalled {
				return
			}
			hookCalled = true
			ContainerLoad(container, registeredByHook)
		},
	)

	container.WithLazyInjection(func() {
		ContainerLoad(container, target)
		ContainerRegister[containerTestDependency](
			container,
			&containerTestDependencyImpl{value: "registered by hook"},
		)
	})

	require.Equal(t, "registered by hook", registeredByHook.dependency.Value())
}

type containerLifecycleFirst interface {
	First()
}

type containerLifecycleSecond interface {
	Second()
}

type containerLifecycle struct{}

func (*containerLifecycle) First()         {}
func (*containerLifecycle) Second()        {}
func (*containerLifecycle) Setup() error   { return nil }
func (*containerLifecycle) Dispose() error { return nil }

func TestContainerDeduplicatesLifecycleInstance(t *testing.T) {
	container := NewContainer()
	lifecycle := new(containerLifecycle)

	ContainerRegister[containerLifecycleFirst](container, lifecycle)
	ContainerRegister[containerLifecycleSecond](container, lifecycle)

	require.Len(t, container.setupable.AcquireAll(), 1)
	require.Len(t, container.disposable.AcquireAll(), 1)
}
