package scene

type Engine interface {
	Run() error
	Start()
	Stop()

	ListContainers() []ApplicationContainer
	GetContainer(name string) ApplicationContainer
	//AddContainer(container ApplicationContainer) error
	//
	//StopContainer(name string) error
	//StartContainer(name string) error
}
