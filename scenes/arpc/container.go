package arpc

import (
	"context"
	"errors"
	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/utils"
	"net"
	"time"
)

type arpcContainer struct {
	apps     []ARpcApp
	addr     string
	server   *arpc.Server
	listener net.Listener
	stopSig  chan int
	log      logger.ILogger
}

// NewARpcContainer create a arpc container
func NewARpcContainer(
	addr string,
	apps []ARpcApp,
	opts ...ARpcOption) scene.Scene {
	server := arpc.NewServer()
	for _, opt := range opts {
		if err := opt(server); err != nil {
			panic(err)
		}
	}
	return &arpcContainer{
		addr:    addr,
		server:  server,
		apps:    apps,
		stopSig: make(chan int),
		log:     registry.Logger.WithPrefix((&arpcContainer{}).ImplName().Identifier()),
	}
}

func (a *arpcContainer) ImplName() scene.ImplName {
	return scene.NewSceneImplNameNoVer("arpc", "Scene")
}

func (a *arpcContainer) Start() error {
	for _, app := range a.apps {
		// todo: handle register service error
		_ = app.RegisterService(a.server)
	}
	var err error
	if !utils.IsValidAddress(a.addr) {
		a.log.Errorf("invalid address: %s", a.addr)
		return errors.New("invalid address " + a.addr)
	}
	a.listener, err = net.Listen("tcp", a.addr)
	if err != nil {
		return nil
	}
	a.server.Handler.SetLogTag("[Server]")
	errCh := make(chan error, 1)
	go func() {
		serveErr := a.server.Serve(a.listener)
		if serveErr != nil {
			errCh <- serveErr
		}
	}()
	select {
	case serveErr := <-errCh:
		a.log.Errorf("failed to serve: %v", serveErr)
		return serveErr
	case <-time.After(1 * time.Second):
		a.log.Infof("arpc server started, listened at %s", a.addr)
		return nil
	}
}

func (a *arpcContainer) Stop(ctx context.Context) error {
	return a.server.Stop()
}

func (a *arpcContainer) ListAppNames() []string {
	names := make([]string, 0, len(a.apps))
	for _, app := range a.apps {
		names = append(names, app.Name().Identifier())
	}
	return names
}
