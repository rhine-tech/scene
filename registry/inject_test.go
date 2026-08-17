package registry

import (
	"testing"

	"github.com/stretchr/testify/require"
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

func TestInject_Flat(t *testing.T) {
	container := NewContainer("flat")
	a := &StructA{}
	Register[IFaceA](container, a)
	b := StructB{}
	container.Load(&b)
	container.Inject()

	require.Nil(t, b.aNil)
	require.NotNil(t, b.a)
	require.Equal(t, a.A(), b.a.A())
}

type StructC struct {
	StructB `aperture:""`
	a       IFaceA `aperture:""`
	aNil    IFaceA
}

func TestInject_Anonymous_Embed(t *testing.T) {
	container := NewContainer("anonymous-embed")
	a := &StructA{}
	Register[IFaceA](container, a)
	c := StructC{}
	container.Load(&c)
	container.Inject()

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

func TestInject_Anonymous_Point_Embed(t *testing.T) {
	require.Panics(t, func() {
		container := NewContainer("nil-anonymous-pointer-embed")
		container.Load(&StructAnonymousPointerEmbed{})
	})

	container := NewContainer("anonymous-pointer-embed")
	a := &StructA{}
	Register[IFaceA](container, a)
	c := StructAnonymousPointerEmbed{StructB: &StructB{}}
	container.Load(&c)
	container.Inject()

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

func TestInject_Pointer_Embed(t *testing.T) {
	require.Panics(t, func() {
		container := NewContainer("nil-pointer-embed")
		container.Load(&StructEmbed{})
	})

	container := NewContainer("pointer-embed")
	a := &StructA{}
	Register[IFaceA](container, a)
	c := StructEmbed{bp: &StructB{}}
	container.Load(&c)
	container.Inject()

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

type IFaceB interface {
	B() string
}

type StructImplB struct {
	a   IFaceA `aperture:""`
	Val string
}

func (s *StructImplB) B() string {
	return s.a.A()
}

type StructEmbedIFace struct {
	iface IFaceB `aperture:"embed"`
}

func TestInject_Interface_Embed(t *testing.T) {
	require.Panics(t, func() {
		container := NewContainer("nil-interface-embed")
		container.Load(&StructEmbedIFace{})
	})

	container := NewContainer("interface-embed")
	a := &StructA{}
	Register[IFaceA](container, a)
	c := StructEmbedIFace{iface: IFaceB(&StructImplB{Val: "BBB"})}
	container.Load(&c)
	container.Inject()

	require.Equal(t, a.A(), c.iface.B())
	require.Equal(t, "BBB", c.iface.(*StructImplB).Val)
}

type directPointerHolder struct {
	dep *StructA `aperture:""`
}

func TestInject_DirectPointer(t *testing.T) {
	container := NewContainer("direct-pointer")
	dependency := &StructA{}
	Register[*StructA](container, dependency)
	holder := directPointerHolder{}
	container.Load(&holder)
	container.Inject()

	require.Same(t, dependency, holder.dep)
}

func TestInject_RequiredDoesNotOverridePreset(t *testing.T) {
	container := NewContainer("required-preset")
	preset := &StructA{}
	holder := StructB{a: preset}
	Register[IFaceA](container, &StructA{})
	container.Load(&holder)
	container.Inject()

	require.Same(t, preset, holder.a)
}

func TestInject_MissingRequiredDependencyPanics(t *testing.T) {
	container := NewContainer("required-missing")
	holder := StructB{}
	container.Load(&holder)

	require.Panics(t, func() {
		container.Inject()
	})
}
