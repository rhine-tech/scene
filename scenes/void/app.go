package void

import (
	"context"
	"github.com/rhine-tech/scene"
)

type VoidApp interface {
	scene.Application
	// Run starts the app's background behavior.
	// It should not block; long-running work should be started in background goroutines.
	Run() error
	// Stop asks the app to stop its background behavior.
	// Implementations should respect ctx cancellation when possible.
	Stop(ctx context.Context) error
}
