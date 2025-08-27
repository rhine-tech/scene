package void

import (
	"github.com/rhine-tech/scene"
)

type VoidApp interface {
	scene.Application
	Run() error // should not block
	Stop() error
}
