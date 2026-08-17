package registry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type optionalIface interface {
	Val() string
}

type optionalImpl struct{}

func (o *optionalImpl) Val() string {
	return "optional"
}

type optionalHolder struct {
	dep optionalIface `aperture:"optional"`
}

type optionalMissingIface interface {
	Missing() string
}

type optionalMissingHolder struct {
	dep optionalMissingIface `aperture:"optional"`
}

type optionalPresetHolder struct {
	dep optionalIface `aperture:"optional"`
}

func TestInject_Optional_WithRegisteredDependency(t *testing.T) {
	container := NewContainer("optional-registered")
	Register[optionalIface](container, &optionalImpl{})
	holder := optionalHolder{}
	container.Load(&holder)

	require.NotPanics(t, func() {
		container.Inject()
	})
	require.NotNil(t, holder.dep)
	require.Equal(t, "optional", holder.dep.Val())
}

func TestInject_Optional_MissingDependencyNoPanic(t *testing.T) {
	container := NewContainer("optional-missing")
	holder := optionalMissingHolder{}
	container.Load(&holder)

	require.NotPanics(t, func() {
		container.Inject()
	})
	require.Nil(t, holder.dep)
}

func TestInject_Optional_DoesNotOverridePreset(t *testing.T) {
	container := NewContainer("optional-preset")
	preset := &optionalImpl{}
	holder := optionalPresetHolder{dep: preset}
	Register[optionalIface](container, &optionalImpl{})
	container.Load(&holder)

	container.Inject()
	require.Same(t, preset, holder.dep)
}
