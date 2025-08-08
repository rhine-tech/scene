package scene

type Engine interface {
	Run() error
	Start() error
	Stop()

	ListContainers() []Scene
	GetContainer(name string) Scene
	//AddContainer(container ApplicationContainer) error
	//
	//StopContainer(name string) error
	//StartContainer(name string) error
}
