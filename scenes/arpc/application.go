package arpc

import (
	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene"
)

type ARpcApp interface {
	scene.Application
	RegisterService(handler arpc.Handler) error
}
