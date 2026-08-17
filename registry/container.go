package registry

import (
	"slices"
)

// Container declares and owns the objects belonging to one module.
type Container struct {
	name         string
	declarations []*declaration
	locals       map[string]*binding
	imports      map[string]any
	sealed       bool
	injected     bool
}

func NewContainer(name ...string) *Container {
	if len(name) > 1 {
		panic("scene registry: container accepts at most one name")
	}
	containerName := "module"
	if len(name) == 1 && name[0] != "" {
		containerName = name[0]
	}
	return &Container{
		name:    containerName,
		locals:  make(map[string]*binding),
		imports: make(map[string]any),
	}
}

// Requires returns the required or optional dependencies that may be exported
// by other containers, in declaration order.
func (c *Container) Requires(optional bool) []string {
	required := make([]string, 0)
	seen := make(map[string]struct{})
	for _, decl := range c.declarations {
		for _, dependency := range decl.requirements {
			if dependency.optional != optional {
				continue
			}
			if _, local := c.locals[dependency.name]; local {
				continue
			}
			if _, exists := seen[dependency.name]; exists {
				continue
			}
			seen[dependency.name] = struct{}{}
			required = append(required, dependency.name)
		}
	}
	return required
}

// Provides returns the dependencies exported by this container, sorted by key.
func (c *Container) Provides() []string {
	provided := make([]string, 0, len(c.locals))
	for key, dependencyBinding := range c.locals {
		if dependencyBinding.visibility != exported {
			continue
		}
		provided = append(provided, key)
	}
	slices.Sort(provided)
	return provided
}

// Inject injects every declared value using this container's local and imported bindings.
func (c *Container) Inject(hooks ...InjectHookFunc) {
	if c.injected {
		return
	}
	for _, decl := range c.declarations {
		inject(c, decl.injections, hooks)
	}
	c.injected = true
}

func (c *Container) lookup(name string) (any, bool) {
	if local, ok := c.locals[name]; ok {
		return local.declaration.instance, true
	}
	imported, ok := c.imports[name]
	return imported, ok
}

// Import adds an external dependency to this container.
func (c *Container) Import(name string, imported any) {
	c.imports[name] = imported
}

// Resolve returns a dependency exported by this container.
func (c *Container) Resolve(name string) (any, bool) {
	dependencyBinding, ok := c.locals[name]
	if !ok || dependencyBinding.visibility != exported || dependencyBinding.declaration.instance == nil {
		return nil, false
	}
	return dependencyBinding.declaration.instance, true
}

func (c *Container) values() []any {
	values := make([]any, 0, len(c.declarations))
	for _, decl := range c.declarations {
		if decl.instance != nil {
			values = append(values, decl.instance)
		}
	}
	return values
}
