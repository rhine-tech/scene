package arpc

import (
	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene"
)

type ARpcApp interface {
	scene.Application
	RegisterService(server *arpc.Server) error
}
