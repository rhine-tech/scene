package grpc

import (
	"context"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	scommon "github.com/rhine-tech/scene/scenes/common"
	"google.golang.org/grpc"
	greflection "google.golang.org/grpc/reflection"
	"net"
)

type grpcContainer struct {
	manager scene.ApplicationManager[GrpcApplication]
	server  *grpc.Server
	status  scene.AppContainerStatus
	addr    string
	logger  logger.ILogger
}

func NewContainer(
	addr string,
	apps ...GrpcApplication) scene.ApplicationContainer {
	return &grpcContainer{
		manager: scommon.NewAppManager(apps...),
		addr:    addr,
		logger:  registry.Logger.WithPrefix(SceneName),
	}
}

func (g *grpcContainer) Name() string {
	return SceneName
}

func (h *grpcContainer) ListAppNames() []scene.AppName {
	return h.manager.ListAppNames()
}

func (g *grpcContainer) GetAppInfo(appID scene.AppName) scene.Application {
	return g.manager.GetApp(appID)
}

func (h *grpcContainer) Status() scene.AppContainerStatus {
	return h.status
}

func (g *grpcContainer) Start() error {
	if g.status == scene.AppContainerStatusRunning {
		return nil
	}
	g.server = grpc.NewServer()
	greflection.Register(g.server)
	g.status = scene.AppContainerStatusRunning
	for _, app := range g.manager.ListApps() {
		if err := app.Create(g.server); err != nil {
			g.logger.Errorf("create app %s error: %s\n", app.Name(), err)
		} else {
			g.logger.Infof("create app %s success", app.Name())
		}
	}
	go func() {
		lis, err := net.Listen("tcp", g.addr)
		if err != nil {
			g.logger.Errorf("listen %s error: %s\n", g.addr, err)
			g.status = scene.AppContainerStatusError
			return
		}
		g.logger.Infof("http server started, listen on %s", g.addr)
		if err := g.server.Serve(lis); err != nil {
			g.logger.Errorf("listen: %s\n", err)
			g.status = scene.AppContainerStatusError
		} else {
			g.status = scene.AppContainerStatusStopped
		}
	}()
	return nil
}

func (g *grpcContainer) Stop(ctx context.Context) error {
	if g.status != scene.AppContainerStatusRunning {
		return nil
	}
	g.server.GracefulStop()
	g.status = scene.AppContainerStatusStopped
	return nil
}
