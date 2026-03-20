package registry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type singletonTestIface interface {
	Value() string
}

type singletonImpl struct {
	v string
}

func (s *singletonImpl) Value() string {
	return s.v
}

func resetSingletonRegistryForTest() {
	singletonLock.Lock()
	singletonRegistry = make(map[string]interface{})
	singletonLock.Unlock()
}

func TestRegisterSingletonAndAcquireSingleton(t *testing.T) {
	resetSingletonRegistryForTest()
	impl := &singletonImpl{v: "a"}
	RegisterSingleton[singletonTestIface](impl)

	got := AcquireSingleton[singletonTestIface](nil)
	require.Same(t, impl, got)
	require.Equal(t, "a", got.Value())
}

func TestRegisterSingletonByNameAndAcquireSingletonByName(t *testing.T) {
	resetSingletonRegistryForTest()
	impl := &singletonImpl{v: "by-name"}
	RegisterSingletonByName("test:singleton", impl)

	got := AcquireSingletonByName[singletonTestIface]("test:singleton")
	require.Same(t, impl, got)
	require.Equal(t, "by-name", got.Value())
}

func TestRegisterSingletonOverwrite(t *testing.T) {
	resetSingletonRegistryForTest()
	first := &singletonImpl{v: "first"}
	second := &singletonImpl{v: "second"}

	RegisterSingleton[singletonTestIface](first)
	RegisterSingleton[singletonTestIface](second)

	got := AcquireSingleton[singletonTestIface](nil)
	require.Same(t, second, got)
	require.Equal(t, "second", got.Value())
}

func TestAcquireSingletonPanicWhenMissing(t *testing.T) {
	resetSingletonRegistryForTest()
	require.Panics(t, func() {
		_ = AcquireSingleton[singletonTestIface](nil)
	})
}

func TestUseFallbackToSingleton(t *testing.T) {
	resetSingletonRegistryForTest()
	impl := &singletonImpl{v: "fallback"}
	RegisterSingleton[singletonTestIface](impl)

	got := Use[singletonTestIface](nil)
	require.Same(t, impl, got)
	require.Equal(t, "fallback", got.Value())
}
