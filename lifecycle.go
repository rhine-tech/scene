package scene

// Lifecycle is initialized when its module is loaded and released when the
// module is unloaded.
type Lifecycle interface {
	Setup() error
	TearDown() error
}

type Defaultable[T any] interface {
	Default() T
}
