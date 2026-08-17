package registry

import (
	"fmt"
	"reflect"
)

type requirement struct {
	name     string
	optional bool
}

type declaration struct {
	injections   []injectionField
	requirements []requirement
	instance     any
}

// Load declares a value owned and injected by the container. The value
// cannot be resolved as a dependency.
func (c *Container) Load(value any) {
	c.addDeclaration(value, loaded, "")
}

// Register binds a value to name for injection inside this container.
func (c *Container) Register(value any, name string) {
	c.addDeclaration(value, registered, name)
}

// Export binds a value to name for injection inside this container and into
// other containers.
func (c *Container) Export(value any, name string) {
	c.addDeclaration(value, exported, name)
}

// Load declares value and returns it unchanged.
func Load[T any](container *Container, value T) T {
	container.Load(value)
	return value
}

// Register declares value and returns it unchanged. T is used as the default
// dependency key.
func Register[T any](container *Container, value T, name ...string) T {
	container.Register(value, dependencyName[T](name))
	return value
}

// Export declares value and returns it unchanged. T is used as the default
// dependency key.
func Export[T any](container *Container, value T, name ...string) T {
	container.Export(value, dependencyName[T](name))
	return value
}

func (c *Container) addDeclaration(value any, visibility visibility, name string) {
	if c.sealed {
		panic("scene registry: cannot modify a container after graph creation")
	}
	if visibility != loaded && name == "" {
		panic("scene registry: dependency name cannot be empty")
	}

	reflected := reflect.ValueOf(value)
	if !reflected.IsValid() || isNil(reflected) {
		panic("scene registry: nil value")
	}

	if visibility != loaded {
		if _, exists := c.locals[name]; exists {
			panic(fmt.Sprintf("scene registry: duplicate dependency %s in %s", name, c.name))
		}
	}

	decl := c.findDeclaration(value)
	if decl == nil {
		decl = &declaration{instance: value}
		if indirect, ok := injectableStruct(reflected); ok {
			decl.injections = walkInjectionFields(indirect)
			decl.requirements = requirementsFromInjectionFields(decl.injections)
		}
		c.declarations = append(c.declarations, decl)
	}

	if visibility != loaded {
		c.locals[name] = &binding{
			key:         name,
			visibility:  visibility,
			declaration: decl,
		}
	}
}

func (c *Container) findDeclaration(value any) *declaration {
	key, ok := identityOf(value)
	if !ok {
		return nil
	}
	for _, decl := range c.declarations {
		declarationKey, ok := identityOf(decl.instance)
		if ok && declarationKey == key {
			return decl
		}
	}
	return nil
}
