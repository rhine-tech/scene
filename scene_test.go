package scene

import (
	"testing"

	"github.com/rhine-tech/scene/registry"
	"github.com/stretchr/testify/require"
)

type sceneFactoryApplication struct {
	name string
}

func (a *sceneFactoryApplication) Name() ImplName {
	return NewSceneImplNameNoVer("test", a.name)
}

type otherSceneFactoryApplication struct{}

func (*otherSceneFactoryApplication) Name() ImplName {
	return NewSceneImplNameNoVer("other", "app")
}

type sceneFactoryModule struct {
	ModuleFactory
	applications []Application
}

func (f sceneFactoryModule) Apps() []Application {
	return f.applications
}

type capturingSceneFactory struct {
	scope        **registry.Scope
	applications *[]*sceneFactoryApplication
}

func (f capturingSceneFactory) Build(
	scope *registry.Scope,
	applications []*sceneFactoryApplication,
) (Scene, error) {
	*f.scope = scope
	*f.applications = append([]*sceneFactoryApplication(nil), applications...)
	return nil, nil
}

func TestWithSceneSelectsMatchingApplicationsInDeclarationOrder(t *testing.T) {
	first := &sceneFactoryApplication{name: "first"}
	second := &sceneFactoryApplication{name: "second"}
	scope := registry.NewScope()
	loader := NewModuleLoaderIn(scope, ModuleFactoryArray{
		sceneFactoryModule{applications: []Application{first, new(otherSceneFactoryApplication)}},
		sceneFactoryModule{applications: []Application{second}},
	})
	require.NoError(t, loader.Init())
	t.Cleanup(func() { require.NoError(t, loader.TearDown()) })

	var receivedScope *registry.Scope
	var receivedApplications []*sceneFactoryApplication
	definition := WithScene(capturingSceneFactory{
		scope:        &receivedScope,
		applications: &receivedApplications,
	})
	_, err := definition.Build(loader)

	require.NoError(t, err)
	require.Same(t, scope, receivedScope)
	require.Equal(t, []*sceneFactoryApplication{first, second}, receivedApplications)
}
