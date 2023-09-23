package scene

import "context"

type AppContainerStatus int

const (
	AppContainerStatusStopped AppContainerStatus = iota
	AppContainerStatusRunning
	AppContainerStatusError
)

type ApplicationFactory[T Application] interface {
	Name() string        // return factory name
	Create(app T) error  // create application
	Destroy(app T) error // not used for now
}

type ApplicationManager[T Application] interface {
	Name() string             // return registry name
	LoadApp(app T) error      // load application
	LoadApps(apps ...T) error // load applications
	GetApp(appID AppName) T   // return application
	ListAppNames() []AppName  // return application names
	ListApps() []T            // return list of applications
}

type ApplicationContainer interface {
	Name() string // return container name

	Start() error                   // start container
	Stop(ctx context.Context) error // stop container
	Status() AppContainerStatus     // return container status

	GetAppInfo(appID AppName) Application // return application info
	ListAppNames() []AppName              // return application names
}
