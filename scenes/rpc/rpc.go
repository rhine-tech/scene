package rpc

import (
	"context"
	"errors"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/utils"
	"net"
	"net/rpc"
)

type RpcApplication interface {
	scene.Application
	RegisterService(server *rpc.Server) error
}

type rpcContainer struct {
	apps     []RpcApplication
	addr     string
	server   *rpc.Server
	listener net.Listener
	stopSig  chan int
	log      logger.ILogger
}

func NewRpcContainer(
	addr string,
	apps []RpcApplication,
	opts ...RpcOption) scene.ApplicationContainer {
	return &rpcContainer{
		addr:    addr,
		server:  rpc.NewServer(),
		apps:    apps,
		stopSig: make(chan int),
		log:     registry.Logger.WithPrefix((&rpcContainer{}).Name().Identifier()),
	}
}

func (r *rpcContainer) Name() scene.ImplName {
	return scene.NewSceneImplNameNoVer("rpc", "Container")
}

func (r *rpcContainer) Start() error {
	for _, app := range r.apps {
		// todo: handle register service error
		_ = app.RegisterService(r.server)
	}
	var err error
	if !utils.IsValidAddress(r.addr) {
		r.log.Errorf("invalid address: %s", r.addr)
		return errors.New("invalid address " + r.addr)
	}
	r.listener, err = net.Listen("tcp", r.addr)
	if err != nil {
		return nil
	}
	go func() {
		for {
			conn, err := r.listener.Accept()
			if err != nil {
				select {
				case <-r.stopSig:
					return
				default:
					// log error log
					continue
				}
			}
			go r.server.ServeConn(conn)
		}
	}()
	r.log.Infof("rpc server started, listened at %s", r.addr)
	return nil
}

func (r *rpcContainer) Stop(ctx context.Context) error {
	if r.listener != nil {
		close(r.stopSig)
		return r.listener.Close()
	}
	return nil
}

func (r *rpcContainer) ListAppNames() []string {
	return []string{}
}
