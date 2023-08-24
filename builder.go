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

type Buildable interface {
	Init() LensInit
	Apps() []any
}

type Builder struct {
}

func (b Builder) Init() LensInit {
	return nil
}

func (b Builder) Apps() []any {
	return nil
}

type BuildableArray []Buildable

func BuildInitArray(builders BuildableArray) InitArray {
	var inits InitArray
	for _, builder := range builders {
		init := builder.Init()
		if init != nil {
			inits = append(inits, init)
		}
	}
	return inits
}

func BuildApps[T Application](builders BuildableArray) []T {
	var apps []T
	for _, builder := range builders {
		for _, app := range builder.Apps() {
			//fmt.Println(app.(AppInit[T]))
			if init, ok := app.(func() T); ok {
				if init != nil {
					apps = append(apps, init())
				}
			}
		}
	}
	return apps
}
