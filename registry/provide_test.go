package registry

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type provideTestIface interface {
	Value() string
}

type provideImpl struct {
	v string
}

func (s *provideImpl) Value() string {
	return s.v
}

func resetDependenciesForTest() {
	defaultContainer.lock.Lock()
	defaultContainer.dependencies = make(map[string]interface{})
	defaultContainer.lock.Unlock()
}

func TestRegisterAndProvide(t *testing.T) {
	resetDependenciesForTest()
	impl := &provideImpl{v: "a"}
	Register[provideTestIface](impl)

	got := Provide[provideTestIface]()
	require.Same(t, impl, got)
	require.Equal(t, "a", got.Value())
}

func TestRegisterAndProvideByName(t *testing.T) {
	resetDependenciesForTest()
	impl := &provideImpl{v: "by-name"}
	Register[provideTestIface](impl, "test:dependency")

	got := Provide[provideTestIface]("test:dependency")
	require.Same(t, impl, got)
	require.Equal(t, "by-name", got.Value())
}

func TestRegisterOverwrite(t *testing.T) {
	resetDependenciesForTest()
	first := &provideImpl{v: "first"}
	second := &provideImpl{v: "second"}

	Register[provideTestIface](first)
	Register[provideTestIface](second)

	got := Provide[provideTestIface]()
	require.Same(t, second, got)
	require.Equal(t, "second", got.Value())
}

func TestProvidePanicWhenMissing(t *testing.T) {
	resetDependenciesForTest()
	require.Panics(t, func() {
		_ = Provide[provideTestIface]()
	})
}

func TestUseFallbackToRegisteredDependency(t *testing.T) {
	resetDependenciesForTest()
	impl := &provideImpl{v: "fallback"}
	Register[provideTestIface](impl)

	got := Use[provideTestIface](nil)
	require.Same(t, impl, got)
	require.Equal(t, "fallback", got.Value())
}

func TestUseReturnsProvidedDependency(t *testing.T) {
	resetDependenciesForTest()
	registered := &provideImpl{v: "registered"}
	provided := &provideImpl{v: "provided"}
	Register[provideTestIface](registered)

	got := Use[provideTestIface](provided)
	require.Same(t, provided, got)
}

func TestMustRegisterPanicsBeforeRegistration(t *testing.T) {
	resetDependenciesForTest()
	err := errors.New("register failed")

	require.PanicsWithValue(t, err, func() {
		MustRegister[provideTestIface](&provideImpl{}, err)
	})
	require.Panics(t, func() {
		_ = Provide[provideTestIface]()
	})
}

func TestRegisterAndProvideRejectMultipleNames(t *testing.T) {
	resetDependenciesForTest()
	require.Panics(t, func() {
		Register[provideTestIface](&provideImpl{}, "first", "second")
	})
	require.Panics(t, func() {
		_ = Provide[provideTestIface]("first", "second")
	})
}
