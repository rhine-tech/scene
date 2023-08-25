package grpc

import (
	"github.com/aynakeya/scene"
	"google.golang.org/grpc"
)

const SceneName = "scene.app-container.grpc"

type GrpcApplication interface {
	scene.Application
	Create(server *grpc.Server) error
}
