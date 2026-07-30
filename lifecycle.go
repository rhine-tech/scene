package scene

type Disposable interface {
	Dispose() error
}

type Setupable interface {
	Setup() error
}

type Defaultable[T any] interface {
	Default() T
}
