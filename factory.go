package scene

// LensInit is a function initialize a lens
// if error happens, it should panic
type LensInit func()

type InitArray []LensInit

func (inits InitArray) Inits() {
	for _, init := range inits {
		init()
	}
}

type AppInit[T Application] func() T

type IModuleDependencyProvider[T any] interface {
	Provide() T
}

type IModuleFactory interface {
	Init() LensInit
	Apps() []any
}

type IDefaultableModuleFactory[T any] interface {
	IModuleFactory
	Defaultable[T]
}

type ModuleFactory struct {
}

func (b ModuleFactory) Init() LensInit {
	return nil
}

func (b ModuleFactory) Apps() []any {
	return nil
}

type ModuleFactoryArray []IModuleFactory

func BuildInitArray(builders ModuleFactoryArray) InitArray {
	var inits InitArray
	for _, builder := range builders {
		init := builder.Init()
		if init != nil {
			inits = append(inits, init)
		}
	}
	return inits
}

func BuildApps[T Application](builders ModuleFactoryArray) []T {
	var apps []T
	for _, builder := range builders {
		for _, app := range builder.Apps() {
			// should be AppInit[T], but golang compiler complains about it
			// So use func() T instead
			if init, ok := app.(func() T); ok {
				if init != nil {
					apps = append(apps, init())
				}
			}
		}
	}
	return apps
}
