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

// NewVoidContainer create a void container
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

func (a *voidContainer) Stop(ctx context.Context) error {
	for _, app := range a.apps {
		_ = app.Stop()
	}
	return nil
}

func (a *voidContainer) ListAppNames() []string {
	names := make([]string, 0, len(a.apps))
	for _, app := range a.apps {
		names = append(names, app.Name().Identifier())
	}
	return names
}
