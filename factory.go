package scene

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/rhine-tech/scene/registry"
)

// AppInit constructs an application.
type AppInit[T Application] func() T

type IModuleDependencyProvider[T any] interface {
	Provide() T
}

// IModuleFactory declares the objects and applications owned by one module.
type IModuleFactory interface {
	Init(container *registry.Container)
	Apps() []Application
}

type IDefaultableModuleFactory[T any] interface {
	IModuleFactory
	Defaultable[T]
}

type ModuleFactory struct{}

func (ModuleFactory) Init(*registry.Container) {}

func (ModuleFactory) Apps() []Application {
	return nil
}

type ModuleFactoryArray []IModuleFactory

// ModuleLoader builds a flat list of module factories in one registry Scope.
type ModuleLoader struct {
	factories    ModuleFactoryArray
	scope        *registry.Scope
	containers   []*registry.Container
	applications []Application
	lifecycles   []Lifecycle
	setupCount   int
	initialized  bool
}

func NewModuleLoader(factories ModuleFactoryArray) *ModuleLoader {
	return NewModuleLoaderIn(registry.DefaultScope(), factories)
}

// NewModuleLoaderIn creates a module loader backed by scope.
func NewModuleLoaderIn(scope *registry.Scope, factories ModuleFactoryArray) *ModuleLoader {
	if scope == nil {
		panic("scene: module loader scope is nil")
	}
	return &ModuleLoader{
		factories: append(ModuleFactoryArray(nil), factories...),
		scope:     scope,
	}
}

// Init declares all modules, resolves their dependencies, and performs
// injection. It does not run lifecycle Setup.
func (l *ModuleLoader) Init() error {
	if l.initialized {
		return nil
	}
	containers := make([]*registry.Container, 0, len(l.factories))
	applications := make([]Application, 0)
	for index, factory := range l.factories {
		if factory == nil {
			return fmt.Errorf("scene: module factory %d is nil", index)
		}
		container := registry.NewContainer(moduleFactoryName(factory, index))
		factory.Init(container)
		for _, application := range factory.Apps() {
			container.Load(application)
			applications = append(applications, application)
		}
		containers = append(containers, container)
	}
	if err := l.scope.Build(containers...); err != nil {
		return err
	}
	// collecting lifecycles
	lifecycles := make([]Lifecycle, 0)
	seen := make(map[Lifecycle]struct{})
	for _, value := range l.scope.OrderedValues() {
		managed, ok := value.(Lifecycle)
		if !ok {
			continue
		}
		reflected := reflect.ValueOf(value)
		if reflected.Kind() == reflect.Ptr && !reflected.IsNil() {
			if _, exists := seen[managed]; exists {
				continue
			}
			seen[managed] = struct{}{}
		}
		lifecycles = append(lifecycles, managed)
	}
	// finishing up
	l.containers = containers
	l.applications = applications
	l.lifecycles = lifecycles
	l.initialized = true
	return nil
}

// Setup initializes module lifecycles in dependency order. Init must complete
// successfully before Setup is called.
func (l *ModuleLoader) Setup() error {
	if !l.initialized {
		return errors.New("scene: module loader is not initialized")
	}
	if l.setupCount != 0 {
		return nil
	}
	for index, managed := range l.lifecycles {
		if err := managed.Setup(); err != nil {
			setupErr := wrapLifecycleError("setup", managed, err)
			rollbackErr := wrapLifecycleError("tear down", managed, managed.TearDown())
			for previous := index - 1; previous >= 0; previous-- {
				previousManaged := l.lifecycles[previous]
				rollbackErr = errors.Join(
					rollbackErr,
					wrapLifecycleError("tear down", previousManaged, previousManaged.TearDown()),
				)
			}
			l.reset()
			return errors.Join(setupErr, rollbackErr)
		}
		l.setupCount = index + 1
	}
	return nil
}

// TearDown releases initialized modules in reverse dependency order.
func (l *ModuleLoader) TearDown() error {
	if l == nil || !l.initialized {
		return nil
	}
	var err error
	for index := l.setupCount - 1; index >= 0; index-- {
		managed := l.lifecycles[index]
		err = errors.Join(err, wrapLifecycleError("tear down", managed, managed.TearDown()))
	}
	l.reset()
	return err
}

func wrapLifecycleError(operation string, managed Lifecycle, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("scene: %s %T: %w", operation, managed, err)
}

func (l *ModuleLoader) reset() {
	l.scope.Reset()
	l.containers = nil
	l.applications = nil
	l.lifecycles = nil
	l.setupCount = 0
	l.initialized = false
}

// Containers returns the initialized module containers in factory order.
func (l *ModuleLoader) Containers() []*registry.Container {
	return append([]*registry.Container(nil), l.containers...)
}

// Applications returns the applications declared during Init in module order.
func (l *ModuleLoader) Applications() []Application {
	return append([]Application(nil), l.applications...)
}

// Scope returns the registry Scope owned by this loader.
func (l *ModuleLoader) Scope() *registry.Scope {
	return l.scope
}

func moduleFactoryName(factory IModuleFactory, index int) string {
	typeOf := reflect.TypeOf(factory)
	for typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
	}
	name := typeOf.String()
	if typeOf.PkgPath() != "" {
		name = typeOf.PkgPath() + "." + typeOf.Name()
	}
	return fmt.Sprintf("%s[%d]", name, index)
}
