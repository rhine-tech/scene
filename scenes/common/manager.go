package common

import (
	"fmt"
	"github.com/rhine-tech/scene"
)

type commonStaticApplicationManagerImpl[T scene.Application] struct {
	apps map[string]T
}

func NewAppManager[T scene.Application](apps ...T) scene.ApplicationManager[T] {
	m := &commonStaticApplicationManagerImpl[T]{
		apps: make(map[string]T),
	}
	_ = m.LoadApps(apps...)
	return m
}

func (c *commonStaticApplicationManagerImpl[T]) Name() string {
	return "scene.app-manager.common"
}

func (c *commonStaticApplicationManagerImpl[T]) LoadApps(apps ...T) error {
	for _, app := range apps {
		_ = c.LoadApp(app)
	}
	return nil
}

func (c *commonStaticApplicationManagerImpl[T]) LoadApp(app T) error {
	id := app.Name().Identifier()
	_, ok := c.apps[id]
	if ok {
		panic(fmt.Sprintf("app %s already exists", string(id)))
		//return "", errcode.AppAlreadyExists.WithDetailStr(string(id))
	}
	c.apps[id] = app
	return nil
}

func (c *commonStaticApplicationManagerImpl[T]) GetApp(appID string) T {
	app, ok := c.apps[appID]
	if !ok {
		return *new(T)
	}
	return app
}

func (c *commonStaticApplicationManagerImpl[T]) ListAppNames() []string {
	ids := make([]string, 0, len(c.apps))
	for id := range c.apps {
		ids = append(ids, id)
	}
	return ids
}

func (c *commonStaticApplicationManagerImpl[T]) ListApps() []T {
	apps := make([]T, 0, len(c.apps))
	for _, app := range c.apps {
		apps = append(apps, app)
	}
	return apps
}
