package scene

import "context"

// Scene is the delivery (controller) layer container,
// contains application from each module
type Scene interface {
	Named

	Start() error                   // start container
	Stop(ctx context.Context) error // stop container

	// GetAppInfo(appID string) Application // return application info
	ListAppNames() []string // return application names
}
