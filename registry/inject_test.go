package registry

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type IFaceA interface {
	A() string
}

type StructA struct{}

func (s *StructA) A() string {
	return "A"
}

type StructB struct {
	a    IFaceA `aperture:""`
	aNil IFaceA
}

func TestTryInject_Flat(t *testing.T) {
	a := &StructA{}
	Register[IFaceA](a)
	b := StructB{}
	Load(&b)
	require.Nil(t, b.aNil)
	require.NotNil(t, b.a)
	require.Equal(t, a.A(), b.a.A())
}

type StructC struct {
	StructB `aperture:""`
	a       IFaceA `aperture:""`
	aNil    IFaceA
}

func TestTryInject_Anonymous_Embed(t *testing.T) {
	a := &StructA{}
	Register[IFaceA](a)
	c := StructC{}
	TryInject(&c)
	require.Nil(t, c.aNil)
	require.NotNil(t, c.a)
	require.Equal(t, a.A(), c.a.A())
	require.Nil(t, c.StructB.aNil)
	require.NotNil(t, c.StructB.a)
	require.Equal(t, a.A(), c.StructB.a.A())
}

type StructAnonymousPointerEmbed struct {
	*StructB `aperture:"embed"`
	a        IFaceA `aperture:""`
	aNil     IFaceA
}

func TestTryInject_Anonymous_Point_Embed(t *testing.T) {
	a := &StructA{}
	Register[IFaceA](a)
	c := StructAnonymousPointerEmbed{}
	require.Panics(t, func() {
		TryInject(&c)
	})
	c = StructAnonymousPointerEmbed{StructB: &StructB{}}
	TryInject(&c)
	require.Nil(t, c.aNil)
	require.NotNil(t, c.a)
	require.Equal(t, a.A(), c.a.A())
	require.Nil(t, c.StructB.aNil)
	require.NotNil(t, c.StructB.a)
	require.Equal(t, a.A(), c.StructB.a.A())
}

type StructEmbed struct {
	bp   *StructB `aperture:"embed"`
	b    StructB  `aperture:"embed"`
	a    IFaceA   `aperture:""`
	aNil IFaceA
}

func TestTryInject_Pointer_Embed(t *testing.T) {
	a := &StructA{}
	Register[IFaceA](a)
	c := StructEmbed{}
	require.Panics(t, func() {
		TryInject(&c)
	})
	c = StructEmbed{bp: &StructB{}}
	TryInject(&c)
	require.Nil(t, c.aNil)
	require.NotNil(t, c.a)
	require.Equal(t, a.A(), c.a.A())
	require.Nil(t, c.b.aNil)
	require.NotNil(t, c.b.a)
	require.Equal(t, a.A(), c.b.a.A())
	require.Nil(t, c.bp.aNil)
	require.NotNil(t, c.bp.a)
	require.Equal(t, a.A(), c.bp.a.A())
}
