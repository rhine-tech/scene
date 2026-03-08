package void

import (
	"context"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
)

type voidContainer struct {
	apps []VoidApp
	log  logger.ILogger
}

// NewVoidContainer creates a container for background-style delivery apps.
// Void apps should start background work in Run and return quickly.
func NewVoidContainer(
	apps []VoidApp,
	opts ...VoidOption) scene.Scene {
	return &voidContainer{
		apps: apps,
		log:  registry.Logger.WithPrefix((&voidContainer{}).ImplName().Identifier()),
	}
}

func (a *voidContainer) ImplName() scene.ImplName {
	return scene.NewSceneImplNameNoVer("void", "Scene")
}

func (a *voidContainer) Start() error {
	for _, app := range a.apps {
		err := app.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

// Stop forwards the shutdown context to all apps and returns the first stop error.
func (a *voidContainer) Stop(ctx context.Context) error {
	var stopErr error
	for _, app := range a.apps {
		if err := app.Stop(ctx); err != nil && stopErr == nil {
			stopErr = err
		}
	}
	return stopErr
}

func (a *voidContainer) ListAppNames() []string {
	names := make([]string, 0, len(a.apps))
	for _, app := range a.apps {
		names = append(names, app.Name().Identifier())
	}
	return names
}
