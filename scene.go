package scene

import "context"

type AppContainerStatus int

const (
	AppContainerStatusStopped AppContainerStatus = iota
	AppContainerStatusRunning
	AppContainerStatusError
)

// ApplicationFactory
// Deprecated: no longer used, keep it for compatibility
type ApplicationFactory[T Application] interface {
	Name() string        // return factory name
	Create(app T) error  // create application
	Destroy(app T) error // not used for now
}

// ApplicationFactory
// Deprecated: no longer used, keep it for compatibility
type ApplicationManager[T Application] interface {
	Name() string             // return registry name
	LoadApp(app T) error      // load application
	LoadApps(apps ...T) error // load applications
	GetApp(appID string) T    // return application
	ListAppNames() []string   // return application names
	ListApps() []T            // return list of applications
}

type ApplicationContainer interface {
	Name() ImplName // return container name

	Start() error                   // start container
	Stop(ctx context.Context) error // stop container
	// Status() AppContainerStatus     // return container status

	// GetAppInfo(appID string) Application // return application info
	ListAppNames() []string // return application names
}
