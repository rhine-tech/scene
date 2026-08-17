package scene

import (
	"context"

	"github.com/rhine-tech/scene/registry"
)

// Scene is the delivery (controller) layer container,
// contains application from each module
type Scene interface {
	Named

	Start() error                   // start container
	Stop(ctx context.Context) error // stop container

	ListAppNames() []string // return application names
}

// Module Component

type Application interface {
	Name() ImplName // return scene
}

// SceneFactory builds one Scene from applications of the matching type.
type SceneFactory[T Application] interface {
	Build(scope *registry.Scope, applications []T) (Scene, error)
}

// SceneDefinition is the type-erased scene declaration consumed by an Engine.
// Use WithScene to create one from a typed SceneFactory.
type SceneDefinition struct {
	build func(*ModuleLoader) (Scene, error)
}

// WithScene declares a Scene. The factory receives only applications that
// implement T, preserving their module declaration order.
func WithScene[T Application](factory SceneFactory[T]) SceneDefinition {
	return SceneDefinition{
		build: func(loader *ModuleLoader) (Scene, error) {
			return factory.Build(loader.Scope(), Applications[T](loader.Applications()))
		},
	}
}

// Build builds the declared Scene from an initialized ModuleLoader.
func (d SceneDefinition) Build(loader *ModuleLoader) (Scene, error) {
	return d.build(loader)
}

// Applications returns the applications that implement T.
func Applications[T Application](applications []Application) []T {
	selected := make([]T, 0)
	for _, application := range applications {
		if typed, ok := application.(T); ok {
			selected = append(selected, typed)
		}
	}
	return selected
}
