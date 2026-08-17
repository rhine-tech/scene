package registry

import (
	"reflect"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

type testDependency interface {
	Value() string
}

type testDependencyImpl struct {
	value string
}

func (d *testDependencyImpl) Value() string { return d.value }

type testLocalDependency interface {
	Local() string
}

type testLocalDependencyImpl struct{}

func (*testLocalDependencyImpl) Local() string { return "local" }

type testOptionalDependency interface {
	Optional() string
}

type testTarget struct {
	local    testLocalDependency    `aperture:""`
	external testDependency         `aperture:"external"`
	optional testOptionalDependency `aperture:"optional"`
}

type testExport interface {
	Exported()
}

type testExportImpl struct{}

func (*testExportImpl) Exported() {}

func TestContainerMetadataUsesInjectionEntries(t *testing.T) {
	container := NewContainer("consumer")
	target := new(testTarget)
	Register[testLocalDependency](container, new(testLocalDependencyImpl))
	container.Load(target)
	Export[testExport](container, new(testExportImpl))

	require.Equal(t, []string{"external"}, container.Requires(false))
	require.Equal(t, []string{reflect.TypeFor[testOptionalDependency]().String()}, container.Requires(true))
	require.Equal(t, []string{reflect.TypeFor[testExport]().String()}, container.Provides())
}

func TestContainerRegisterAndExportRequireName(t *testing.T) {
	require.PanicsWithValue(t, "scene registry: dependency name cannot be empty", func() {
		NewContainer("module").Register(new(testDependencyImpl), "")
	})
	require.PanicsWithValue(t, "scene registry: dependency name cannot be empty", func() {
		NewContainer("module").Export(new(testDependencyImpl), "")
	})
}

type firstBoundDependency interface {
	First()
}

type secondBoundDependency interface {
	Second()
}

type multiBoundDependency struct{}

func (*multiBoundDependency) First()  {}
func (*multiBoundDependency) Second() {}

func TestContainerProvidesSortedByKey(t *testing.T) {
	container := NewContainer("module")
	dependency := new(multiBoundDependency)
	Export[secondBoundDependency](container, dependency)
	Export[firstBoundDependency](container, dependency)

	want := []string{
		reflect.TypeFor[secondBoundDependency]().String(),
		reflect.TypeFor[firstBoundDependency]().String(),
	}
	slices.Sort(want)
	require.Equal(t, want, container.Provides())
}

func TestMultipleBindingsShareOneOwnedValue(t *testing.T) {
	container := NewContainer("module")
	dependency := new(multiBoundDependency)
	Export[firstBoundDependency](container, dependency)
	Export[secondBoundDependency](container, dependency)

	scope := NewScope()
	require.NoError(t, scope.Build(container))
	require.Equal(t, []any{dependency}, scope.OrderedValues())
	require.Same(t, dependency, UseIn[firstBoundDependency](scope, nil))
	require.Same(t, dependency, UseIn[secondBoundDependency](scope, nil))
}

func TestScopeInjectsContainer(t *testing.T) {
	provider := NewContainer("provider")
	external := &testDependencyImpl{value: "external"}
	Export[testDependency](provider, external, "external")

	consumer := NewContainer("consumer")
	target := new(testTarget)
	Register[testLocalDependency](consumer, new(testLocalDependencyImpl))
	consumer.Load(target)

	scope := NewScope()
	require.NoError(t, scope.Build(consumer, provider))
	require.Equal(t, "local", target.local.Local())
	require.Same(t, external, target.external)
	require.Nil(t, target.optional)
}

type testOptionalDependencyImpl struct{}

func (*testOptionalDependencyImpl) Optional() string { return "optional" }

func TestScopeInjectsOptionalDependencyWhenExported(t *testing.T) {
	provider := NewContainer("provider")
	optional := new(testOptionalDependencyImpl)
	Export[testOptionalDependency](provider, optional)

	consumer := NewContainer("consumer")
	target := new(testTarget)
	Register[testLocalDependency](consumer, new(testLocalDependencyImpl))
	Export[testDependency](consumer, &testDependencyImpl{}, "external")
	consumer.Load(target)

	scope := NewScope()
	require.NoError(t, scope.Build(consumer, provider))
	require.Same(t, optional, target.optional)
}

type declaredFieldTarget struct {
	external testDependency         `aperture:"external"`
	optional testOptionalDependency `aperture:"optional"`
}

func TestDeclaredValueContributesInjectionRequirements(t *testing.T) {
	external := &testDependencyImpl{value: "external"}
	optional := new(testOptionalDependencyImpl)
	provider := NewContainer("provider")
	Export[testDependency](provider, external, "external")
	Export[testOptionalDependency](provider, optional)

	consumer := NewContainer("consumer")
	target := new(declaredFieldTarget)
	consumer.Load(target)
	require.Equal(t, []string{"external"}, consumer.Requires(false))
	require.Equal(t, []string{reflect.TypeFor[testOptionalDependency]().String()}, consumer.Requires(true))

	scope := NewScope()
	require.NoError(t, scope.Build(consumer, provider))
	require.Same(t, external, target.external)
	require.Same(t, optional, target.optional)
}

func TestRegisterIsNotVisibleToAnotherContainer(t *testing.T) {
	provider := NewContainer("provider")
	Register[testDependency](provider, new(testDependencyImpl))
	consumer := NewContainer("consumer")
	consumer.Load(&struct {
		dependency testDependency `aperture:""`
	}{})

	err := NewScope().Build(provider, consumer)
	require.ErrorContains(t, err, "requires missing dependency")
}

func TestLoadIsNotResolvableInsideContainer(t *testing.T) {
	container := NewContainer("module")
	container.Load(new(testDependencyImpl))
	container.Load(&struct {
		dependency testDependency `aperture:""`
	}{})

	err := NewScope().Build(container)
	require.ErrorContains(t, err, "requires missing dependency")
}

func TestScopeRejectsDuplicateExports(t *testing.T) {
	first := NewContainer("first")
	Export[testDependency](first, new(testDependencyImpl))
	second := NewContainer("second")
	Export[testDependency](second, new(testDependencyImpl))

	err := NewScope().Build(first, second)
	require.ErrorContains(t, err, "exported by both")
}

type cycleA interface{ A() }
type cycleB interface{ B() }

type cycleAImpl struct {
	b cycleB `aperture:""`
}

func (*cycleAImpl) A() {}

type cycleBImpl struct {
	a cycleA `aperture:""`
}

func (*cycleBImpl) B() {}

func TestFieldInjectionCycleInsideContainer(t *testing.T) {
	container := NewContainer("module")
	a := new(cycleAImpl)
	b := new(cycleBImpl)
	Register[cycleA](container, a)
	Register[cycleB](container, b)

	require.Empty(t, container.Requires(false))
	scope := NewScope()
	require.NoError(t, scope.Build(container))
	require.Same(t, b, a.b)
	require.Same(t, a, b.a)
}

func TestContainerInject(t *testing.T) {
	container := NewContainer("module")
	dependency := new(testDependencyImpl)
	target := &struct {
		dependency testDependency `aperture:""`
	}{}
	Register[testDependency](container, dependency)
	container.Load(target)

	container.Inject()

	require.Same(t, dependency, target.dependency)
}

func TestContainersConnectWithoutScope(t *testing.T) {
	provider := NewContainer("provider")
	dependency := new(testDependencyImpl)
	Export[testDependency](provider, dependency, "external")
	Register[testLocalDependency](provider, new(testLocalDependencyImpl), "local")

	consumer := NewContainer("consumer")
	target := new(declaredFieldTarget)
	consumer.Load(target)

	exported, ok := provider.Resolve("external")
	require.True(t, ok)
	consumer.Import("external", exported)
	consumer.Inject()

	require.Same(t, dependency, target.external)
	_, ok = provider.Resolve("local")
	require.False(t, ok)
}

func TestScopeBuildDoesNotRecoverInjectionPanic(t *testing.T) {
	container := NewContainer("module")
	Register[testDependency](container, new(testDependencyImpl))
	container.Load(&struct {
		dependency testDependency `aperture:""`
	}{})
	scope := NewScope()
	scope.RegisterInjectHookFunc(
		reflect.TypeFor[testDependency]().String(),
		func(string, reflect.Value, reflect.Value, *interface{}) {
			panic("boom")
		},
	)

	require.PanicsWithValue(t, "boom", func() {
		_ = scope.Build(container)
	})
}

func TestScopeHookRegisteredBeforeBuildIsUsed(t *testing.T) {
	provider := NewContainer("provider")
	Export[testDependency](provider, &testDependencyImpl{value: "original"}, "external")
	consumer := NewContainer("consumer")
	target := new(testTarget)
	Register[testLocalDependency](consumer, new(testLocalDependencyImpl))
	consumer.Load(target)

	scope := NewScope()
	scope.RegisterInjectHookFunc("external", func(_ string, _ reflect.Value, _ reflect.Value, instance *interface{}) {
		*instance = testDependency(&testDependencyImpl{value: "hooked"})
	})

	require.NoError(t, scope.Build(provider, consumer))
	require.Equal(t, "hooked", target.external.Value())
}

func TestSetAndUseGlobalDependency(t *testing.T) {
	scope := NewScope()
	dependency := &testDependencyImpl{value: "global"}
	SetIn[testDependency](scope, dependency)
	require.Same(t, dependency, UseIn[testDependency](scope, nil))
	_, exists := LookupIn[testDependency](NewScope())
	require.False(t, exists)
}

type nestedInjection interface {
	Nested()
}

type nestedInjectionImpl struct {
	dependency testDependency `aperture:""`
}

func (*nestedInjectionImpl) Nested() {}

type embeddedInjectionTarget struct {
	nested nestedInjection `aperture:"embed"`
}

func TestInjectionTraversesEmbeddedField(t *testing.T) {
	dependency := &testDependencyImpl{value: "cached"}
	provider := NewContainer("provider")
	Export[testDependency](provider, dependency)

	nested := new(nestedInjectionImpl)
	target := &embeddedInjectionTarget{nested: nested}
	consumer := NewContainer("consumer")
	consumer.Load(target)

	scope := NewScope()
	require.NoError(t, scope.Build(provider, consumer))
	require.Same(t, dependency, nested.dependency)
}

type embeddedPointerTarget struct {
	*nestedInjectionImpl `aperture:"embed"`
}

func TestWalkInjectionFieldsRejectsNilEmbed(t *testing.T) {
	consumer := NewContainer("consumer")
	require.PanicsWithValue(t,
		"scene registry: failed to inject into a nil struct when injecting nestedInjectionImpl",
		func() {
			consumer.Load(new(embeddedPointerTarget))
		},
	)
}

func TestGenericDeclarationsReturnValuesAndUseTypeParameterAsKey(t *testing.T) {
	dependency := &testDependencyImpl{value: "generic"}
	provider := NewContainer("provider")
	exported := Export[testDependency](provider, dependency)
	require.Same(t, dependency, exported)
	require.Equal(t, []string{reflect.TypeFor[testDependency]().String()}, provider.Provides())

	consumer := NewContainer("consumer")
	target := Load(consumer, &struct {
		dependency testDependency `aperture:""`
	}{})

	scope := NewScope()
	require.NoError(t, scope.Build(provider, consumer))
	require.Same(t, dependency, target.dependency)
}

func TestContainerCannotChangeAfterScopeBuild(t *testing.T) {
	container := NewContainer("sealed")
	require.NoError(t, NewScope().Build(container))

	require.Panics(t, func() {
		container.Load(new(testDependencyImpl))
	})
}
