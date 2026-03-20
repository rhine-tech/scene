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

func TestTryInject_Optional_WithRegisteredDependency(t *testing.T) {
	Register[optionalIface](&optionalImpl{})
	holder := optionalHolder{}
	require.NotPanics(t, func() {
		TryInject(&holder)
	})
	require.NotNil(t, holder.dep)
	require.Equal(t, "optional", holder.dep.Val())
}

func TestTryInject_Optional_MissingDependencyNoPanic(t *testing.T) {
	holder := optionalMissingHolder{}
	require.NotPanics(t, func() {
		TryInject(&holder)
	})
	require.Nil(t, holder.dep)
}

func TestTryInject_Optional_DoesNotOverridePreset(t *testing.T) {
	preset := &optionalImpl{}
	holder := optionalPresetHolder{dep: preset}
	Register[optionalIface](&optionalImpl{})
	TryInject(&holder)
	require.Same(t, preset, holder.dep)
}
