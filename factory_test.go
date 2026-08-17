package scene

import (
	"errors"
	"reflect"
	"testing"

	"github.com/rhine-tech/scene/registry"
	"github.com/stretchr/testify/require"
)

type loaderDependency interface {
	Dependency()
}

type loaderLifecycle struct {
	name        string
	events      *[]string
	setupErr    error
	tearDownErr error
}

func (*loaderLifecycle) Dependency() {}

func (l *loaderLifecycle) Setup() error {
	*l.events = append(*l.events, "setup:"+l.name)
	return l.setupErr
}

func (l *loaderLifecycle) TearDown() error {
	*l.events = append(*l.events, "teardown:"+l.name)
	return l.tearDownErr
}

type loaderProviderFactory struct {
	ModuleFactory
	lifecycle *loaderLifecycle
}

func (f loaderProviderFactory) Init(container *registry.Container) {
	registry.Export[loaderDependency](container, f.lifecycle)
}

type loaderConsumerFactory struct {
	ModuleFactory
	events *[]string
}

type loaderConsumerLifecycle struct {
	dependency loaderDependency `aperture:""`
	events     *[]string
}

func (l *loaderConsumerLifecycle) Setup() error {
	*l.events = append(*l.events, "setup:consumer")
	return nil
}

func (l *loaderConsumerLifecycle) TearDown() error {
	*l.events = append(*l.events, "teardown:consumer")
	return nil
}

func (f loaderConsumerFactory) Init(container *registry.Container) {
	container.Load(&loaderConsumerLifecycle{events: f.events})
}

func TestModuleLoaderOwnsLifecycleOutsideRegistry(t *testing.T) {
	events := make([]string, 0)
	provider := &loaderLifecycle{name: "provider", events: &events}
	scope := registry.NewScope()
	loader := NewModuleLoaderIn(scope, ModuleFactoryArray{
		loaderConsumerFactory{events: &events},
		loaderProviderFactory{lifecycle: provider},
	})

	require.NoError(t, loader.Init())
	require.NoError(t, loader.Setup())
	require.Same(t, provider, registry.UseIn[loaderDependency](scope, nil))
	require.Equal(t, []string{"setup:provider", "setup:consumer"}, events)
	require.Equal(t,
		[]string{reflect.TypeFor[loaderDependency]().String()},
		loader.Containers()[0].Requires(false),
	)

	require.NoError(t, loader.TearDown())
	require.Equal(t, []string{
		"setup:provider",
		"setup:consumer",
		"teardown:consumer",
		"teardown:provider",
	}, events)
	_, exists := registry.LookupIn[loaderDependency](scope)
	require.False(t, exists)
}

type loaderLifecycleFactory struct {
	ModuleFactory
	lifecycle *loaderLifecycle
}

func (f loaderLifecycleFactory) Init(container *registry.Container) {
	container.Load(f.lifecycle)
}

func TestModuleLoaderRollsBackSetupFailure(t *testing.T) {
	events := make([]string, 0)
	want := errors.New("setup failed")
	scope := registry.NewScope()
	loader := NewModuleLoaderIn(scope, ModuleFactoryArray{
		loaderLifecycleFactory{lifecycle: &loaderLifecycle{name: "first", events: &events}},
		loaderLifecycleFactory{lifecycle: &loaderLifecycle{name: "second", events: &events, setupErr: want}},
	})
	t.Cleanup(func() { _ = loader.TearDown() })

	require.NoError(t, loader.Init())
	err := loader.Setup()
	require.ErrorIs(t, err, want)
	require.Equal(t, []string{
		"setup:first",
		"setup:second",
		"teardown:second",
		"teardown:first",
	}, events)
	require.Empty(t, scope.OrderedValues())
	require.Empty(t, loader.Containers())
}

func TestModuleLoaderReportsTearDownLifecycle(t *testing.T) {
	events := make([]string, 0)
	want := errors.New("tear down failed")
	loader := NewModuleLoaderIn(registry.NewScope(), ModuleFactoryArray{
		loaderLifecycleFactory{lifecycle: &loaderLifecycle{
			name:        "module",
			events:      &events,
			tearDownErr: want,
		}},
	})

	require.NoError(t, loader.Init())
	require.NoError(t, loader.Setup())
	err := loader.TearDown()
	require.ErrorIs(t, err, want)
}

func TestModuleLoaderRebuildsScopeAfterTearDown(t *testing.T) {
	events := make([]string, 0)
	scope := registry.NewScope()
	loader := NewModuleLoaderIn(scope, ModuleFactoryArray{
		loaderLifecycleFactory{lifecycle: &loaderLifecycle{name: "module", events: &events}},
	})

	require.NoError(t, loader.Init())
	require.NoError(t, loader.Setup())
	require.NoError(t, loader.TearDown())
	require.NoError(t, loader.Init())
	require.NoError(t, loader.Setup())
	require.NoError(t, loader.TearDown())
	require.Equal(t, []string{
		"setup:module",
		"teardown:module",
		"setup:module",
		"teardown:module",
	}, events)
}

func TestModuleLoaderSetupRequiresInit(t *testing.T) {
	events := make([]string, 0)
	scope := registry.NewScope()
	loader := NewModuleLoaderIn(scope, ModuleFactoryArray{
		loaderLifecycleFactory{lifecycle: &loaderLifecycle{name: "module", events: &events}},
	})

	require.EqualError(t, loader.Setup(), "scene: module loader is not initialized")
	require.Empty(t, events)
	require.Empty(t, scope.OrderedValues())
}

type loaderApplication struct {
	dependency loaderDependency `aperture:""`
}

func (*loaderApplication) Name() ImplName {
	return NewSceneImplNameNoVer("test", "app")
}

type loaderApplicationFactory struct {
	ModuleFactory
	application *loaderApplication
	calls       *int
}

func (f loaderApplicationFactory) Apps() []Application {
	(*f.calls)++
	return []Application{
		f.application,
	}
}

type loaderOwnedApplicationFactory struct {
	ModuleFactory
	application *loaderApplication
}

func (f loaderOwnedApplicationFactory) Init(container *registry.Container) {
	container.Load(f.application)
}

func TestModuleLoaderApplicationsUseDeclaredCatalog(t *testing.T) {
	events := make([]string, 0)
	provider := &loaderLifecycle{name: "provider", events: &events}
	declared := new(loaderApplication)
	ownedOnly := new(loaderApplication)
	appCalls := 0
	scope := registry.NewScope()
	loader := NewModuleLoaderIn(scope, ModuleFactoryArray{
		loaderProviderFactory{lifecycle: provider},
		loaderOwnedApplicationFactory{application: ownedOnly},
		loaderApplicationFactory{application: declared, calls: &appCalls},
	})
	t.Cleanup(func() { _ = loader.TearDown() })

	require.NoError(t, loader.Init())
	applications := Applications[*loaderApplication](loader.Applications())
	require.Len(t, applications, 1)
	require.Same(t, declared, applications[0])
	require.Same(t, provider, declared.dependency)
	require.Same(t, provider, ownedOnly.dependency)
	require.Equal(t, 1, appCalls)

	applications[0] = nil
	require.Same(t, declared, Applications[*loaderApplication](loader.Applications())[0])
	require.Equal(t, 1, appCalls)
}
