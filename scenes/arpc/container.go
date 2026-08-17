package arpc

import (
	"context"
	"errors"
	"net"

	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/utils"
)

type arpcContainer struct {
	apps     []ARpcApp
	addr     string
	server   *arpc.Server
	listener net.Listener
	done     chan error
	log      logger.ILogger
}

// Factory builds an ARPC scene from ARPC applications.
type Factory struct {
	Addr    string
	Options []ServerOption
}

var _ scene.SceneFactory[ARpcApp] = Factory{}

// NewFactory declares an ARPC scene.
func NewFactory(addr string, options ...ServerOption) scene.SceneDefinition {
	return scene.WithScene[ARpcApp](Factory{
		Addr:    addr,
		Options: options,
	})
}

func (f Factory) Build(scope *registry.Scope, apps []ARpcApp) (scene.Scene, error) {
	server := arpc.NewServer()
	for _, option := range f.Options {
		if err := option(scope, server); err != nil {
			return nil, err
		}
	}
	log, exists := registry.LookupIn[logger.ILogger](scope)
	if !exists {
		log = logger.NoopLogger{}
	}
	return &arpcContainer{
		addr:   f.Addr,
		server: server,
		apps:   apps,
		log:    log.WithPrefix((&arpcContainer{}).ImplName().Identifier()),
	}, nil
}

func (a *arpcContainer) ImplName() scene.ImplName {
	return scene.NewSceneImplNameNoVer("arpc", "Scene")
}

func (a *arpcContainer) Start() error {
	if !utils.IsValidAddress(a.addr) {
		a.log.Errorf("invalid address: %s", a.addr)
		return errors.New("invalid address " + a.addr)
	}
	for _, app := range a.apps {
		// todo: handle register service error
		_ = app.RegisterService(a.server.Handler)
	}
	listener, err := net.Listen("tcp", a.addr)
	if err != nil {
		return err
	}
	a.listener = listener
	a.done = make(chan error, 1)
	a.server.Handler.SetLogTag("[Server]")
	a.log.Infof("arpc server started, listened at %s", a.addr)
	go func() {
		err := a.server.Serve(listener)
		if err != nil && !errors.Is(err, net.ErrClosed) {
			a.log.Errorf("failed to serve: %v", err)
		}
		a.done <- err
	}()
	return nil
}

func (a *arpcContainer) Stop(ctx context.Context) error {
	if a.listener == nil {
		return nil
	}
	closeErr := a.listener.Close()
	if errors.Is(closeErr, net.ErrClosed) {
		closeErr = nil
	}
	select {
	case serveErr := <-a.done:
		if serveErr != nil && !errors.Is(serveErr, net.ErrClosed) {
			return serveErr
		}
		return closeErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *arpcContainer) ListAppNames() []string {
	names := make([]string, 0, len(a.apps))
	for _, app := range a.apps {
		names = append(names, app.Name().Identifier())
	}
	return names
}
