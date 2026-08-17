package cmd

import (
	"testing"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/registry"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type commandLifecycle struct {
	events *[]string
}

func (l *commandLifecycle) Setup() error {
	*l.events = append(*l.events, "setup")
	return nil
}

func (l *commandLifecycle) TearDown() error {
	*l.events = append(*l.events, "teardown")
	return nil
}

type commandApplication struct {
	events *[]string
}

func (*commandApplication) Name() scene.ImplName {
	return scene.NewSceneImplNameNoVer("cmd-test", "command")
}

func (a *commandApplication) Command(root *cobra.Command) error {
	root.AddCommand(&cobra.Command{
		Use: "run",
		Run: func(*cobra.Command, []string) {
			*a.events = append(*a.events, "run")
		},
	})
	return nil
}

type nonCommandApplication struct{}

func (*nonCommandApplication) Name() scene.ImplName {
	return scene.NewSceneImplNameNoVer("cmd-test", "other")
}

type commandModuleFactory struct {
	scene.ModuleFactory
	events *[]string
}

func (f commandModuleFactory) Init(container *registry.Container) {
	container.Load(&commandLifecycle{events: f.events})
}

func (f commandModuleFactory) Apps() []scene.Application {
	return []scene.Application{
		&commandApplication{events: f.events},
		new(nonCommandApplication),
	}
}

func TestContainerOwnsModuleLifecycleAndSelectsCommandApps(t *testing.T) {
	events := make([]string, 0)
	scope := registry.NewScope()
	loader := scene.NewModuleLoaderIn(scope, scene.ModuleFactoryArray{
		commandModuleFactory{events: &events},
	})
	container := NewContainer("test", "test commands", loader)
	container.RootCommand().SetArgs([]string{"run"})

	require.Equal(t, []string{
		scene.NewSceneImplNameNoVer("cmd-test", "command").Identifier(),
	}, container.ListAppNames())
	require.NoError(t, container.Execute())
	require.Equal(t, []string{"setup", "run", "teardown"}, events)
	require.Empty(t, scope.OrderedValues())
}
