package engines

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/registry"
	"github.com/stretchr/testify/require"
)

type engineLifecycle struct {
	events *[]string
}

func (l *engineLifecycle) Setup() error {
	*l.events = append(*l.events, "setup")
	return nil
}

func (l *engineLifecycle) TearDown() error {
	*l.events = append(*l.events, "teardown")
	return nil
}

type engineModuleFactory struct {
	scene.ModuleFactory
	lifecycle *engineLifecycle
}

func (f engineModuleFactory) Init(container *registry.Container) {
	container.Load(f.lifecycle)
}

type engineApplication struct {
	name string
}

func (a *engineApplication) Name() scene.ImplName {
	return scene.NewSceneImplNameNoVer("test", a.name)
}

type engineApplicationFactory struct {
	scene.ModuleFactory
	applications []scene.Application
}

func (f engineApplicationFactory) Apps() []scene.Application {
	return f.applications
}

type engineScene struct {
	name     string
	events   *[]string
	startErr error
	started  chan<- struct{}
}

func (s *engineScene) ImplName() scene.ImplName {
	return scene.NewSceneImplNameNoVer("test", s.name)
}

func (s *engineScene) Start() error {
	*s.events = append(*s.events, "start:"+s.name)
	if s.started != nil {
		close(s.started)
	}
	return s.startErr
}

func (s *engineScene) Stop(context.Context) error {
	*s.events = append(*s.events, "stop:"+s.name)
	return nil
}

func (*engineScene) ListAppNames() []string { return nil }

type engineSceneFactory struct {
	name         string
	events       *[]string
	receivedApps *[]int
	buildErr     error
	startErr     error
	started      chan<- struct{}
}

func (f engineSceneFactory) Build(
	_ *registry.Scope,
	applications []*engineApplication,
) (scene.Scene, error) {
	*f.events = append(*f.events, "build:"+f.name)
	if f.receivedApps != nil {
		*f.receivedApps = append(*f.receivedApps, len(applications))
	}
	if f.buildErr != nil {
		return nil, f.buildErr
	}
	return &engineScene{name: f.name, events: f.events, startErr: f.startErr, started: f.started}, nil
}

func TestEngineDelegatesLifecycleToModuleLoader(t *testing.T) {
	events := make([]string, 0)
	receivedApps := make([]int, 0)
	loader := scene.NewModuleLoaderIn(registry.NewScope(), scene.ModuleFactoryArray{
		engineModuleFactory{lifecycle: &engineLifecycle{events: &events}},
		engineApplicationFactory{applications: []scene.Application{
			&engineApplication{name: "one"},
			&engineApplication{name: "two"},
		}},
	})
	engine := NewEngine(loader,
		scene.WithScene(engineSceneFactory{name: "first", events: &events, receivedApps: &receivedApps}),
		scene.WithScene(engineSceneFactory{name: "second", events: &events, receivedApps: &receivedApps}),
	)

	require.NoError(t, engine.Start())
	engine.Stop()
	require.Equal(t, []string{
		"setup",
		"build:first",
		"build:second",
		"start:first",
		"start:second",
		"stop:second",
		"stop:first",
		"teardown",
	}, events)
	require.Equal(t, []int{2, 2}, receivedApps)
}

func TestEngineRollsBackStartedScenesAndModules(t *testing.T) {
	events := make([]string, 0)
	loader := scene.NewModuleLoaderIn(registry.NewScope(), scene.ModuleFactoryArray{
		engineModuleFactory{lifecycle: &engineLifecycle{events: &events}},
	})
	want := errors.New("start failed")
	engine := NewEngine(loader,
		scene.WithScene(engineSceneFactory{name: "first", events: &events}),
		scene.WithScene(engineSceneFactory{name: "second", events: &events, startErr: want}),
	)

	require.ErrorIs(t, engine.Start(), want)
	require.Equal(t, []string{
		"setup",
		"build:first",
		"build:second",
		"start:first",
		"start:second",
		"stop:first",
		"teardown",
	}, events)
}

func TestEngineTearsDownModulesWhenSceneBuildFails(t *testing.T) {
	events := make([]string, 0)
	loader := scene.NewModuleLoaderIn(registry.NewScope(), scene.ModuleFactoryArray{
		engineModuleFactory{lifecycle: &engineLifecycle{events: &events}},
	})
	want := errors.New("build failed")
	engine := NewEngine(loader,
		scene.WithScene(engineSceneFactory{name: "first", events: &events}),
		scene.WithScene(engineSceneFactory{name: "second", events: &events, buildErr: want}),
	)

	require.ErrorIs(t, engine.Start(), want)
	require.Equal(t, []string{
		"setup",
		"build:first",
		"build:second",
		"teardown",
	}, events)
}

func TestEngineRunPreservesStartFailure(t *testing.T) {
	events := make([]string, 0)
	want := errors.New("build failed")
	loader := scene.NewModuleLoaderIn(registry.NewScope(), nil)
	engine := NewEngine(loader, scene.WithScene(engineSceneFactory{
		name:     "failed",
		events:   &events,
		buildErr: want,
	}))

	err := engine.Run()
	require.ErrorIs(t, err, errStartEngineFailed)
	require.ErrorIs(t, err, want)
}

func TestEngineRunStopsOnSIGTERM(t *testing.T) {
	events := make([]string, 0)
	started := make(chan struct{})
	loader := scene.NewModuleLoaderIn(registry.NewScope(), scene.ModuleFactoryArray{
		engineModuleFactory{lifecycle: &engineLifecycle{events: &events}},
	})
	engine := NewEngine(loader, scene.WithScene(engineSceneFactory{
		name:    "signal",
		events:  &events,
		started: started,
	}))
	done := make(chan error, 1)
	go func() {
		done <- engine.Run()
	}()

	<-started
	require.NoError(t, syscall.Kill(os.Getpid(), syscall.SIGTERM))
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("engine did not stop after SIGTERM")
	}
	require.Equal(t, []string{
		"setup",
		"build:signal",
		"start:signal",
		"stop:signal",
		"teardown",
	}, events)
}
