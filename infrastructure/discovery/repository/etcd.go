package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/discovery"
	"github.com/rhine-tech/scene/infrastructure/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

type etcdRegister struct {
	uri     string
	client  *clientv3.Client
	session *concurrency.Session
	manager endpoints.Manager
	nodes   []*discovery.Node
	log     logger.ILogger `aperture:""`

	watchers   map[string]context.CancelFunc
	watcherMux sync.Mutex
}

func (e *etcdRegister) ImplName() scene.ImplName {
	return scene.NewSceneImplNameNoVer("discovery", "EtcdRegister")
}

func (e *etcdRegister) Dispose() error {
	for _, node := range e.nodes {
		err := e.Deregister(node)
		if err != nil {
			e.log.Warnf("deregister node %s error: %s", node, err)
		}
	}
	e.watcherMux.Lock()
	for key, cancel := range e.watchers {
		cancel()
		delete(e.watchers, key)
	}
	e.watcherMux.Unlock()
	return e.session.Close()
}

func (e *etcdRegister) Setup() error {
	e.log = e.log.WithPrefix("discovery.etcd")
	cli, err := clientv3.NewFromURL(e.uri)
	if err != nil {
		return err
	}
	e.client = cli
	e.session, err = concurrency.NewSession(e.client)
	if err != nil {
		return err
	}
	go func() {
		<-e.session.Done()
	}()
	e.watchers = map[string]context.CancelFunc{}
	return nil
}

func (e *etcdRegister) formatKey(node *discovery.Node) string {
	return fmt.Sprintf("%s/%s", node.Key, node.Server)
}

func (e *etcdRegister) Register(key string, endpoint discovery.Endpoint) (node *discovery.Node, err error) {
	e.manager, err = endpoints.NewManager(e.client, key)
	if err != nil {
		return nil, err
	}
	node = &discovery.Node{
		Key:    key,
		Server: endpoint.Addr,
	}
	err = e.manager.AddEndpoint(context.TODO(),
		e.formatKey(node),
		endpoints.Endpoint{Addr: endpoint.Addr},
		clientv3.WithLease(e.session.Lease()))
	if err != nil {
		return nil, err
	}
	e.nodes = append(e.nodes, node)
	return node, nil
}

func (e *etcdRegister) Deregister(node *discovery.Node) error {
	em, err := endpoints.NewManager(e.client, node.Key)
	if err != nil {
		return err
	}
	return em.DeleteEndpoint(context.TODO(),
		e.formatKey(node),
		clientv3.WithLease(e.session.Lease()))
}

func (e *etcdRegister) Resolve(ctx context.Context, key string) ([]discovery.Endpoint, error) {
	em, err := endpoints.NewManager(e.client, key)
	if err != nil {
		return nil, err
	}
	data, err := em.List(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]discovery.Endpoint, 0, len(data))
	for _, ep := range data {
		results = append(results, discovery.Endpoint{Addr: ep.Addr})
	}
	return results, nil
}

func (e *etcdRegister) Watch(ctx context.Context, key string, handler discovery.EndpointWatchHandler) (context.CancelFunc, error) {
	em, err := endpoints.NewManager(e.client, key)
	if err != nil {
		return nil, err
	}
	watchCtx, cancel := context.WithCancel(ctx)
	ch, err := em.NewWatchChannel(watchCtx)
	if err != nil {
		cancel()
		return nil, err
	}
	go func() {
		defer func() {
			e.watcherMux.Lock()
			delete(e.watchers, key)
			e.watcherMux.Unlock()
			cancel()
		}()
		current := map[string]endpoints.Endpoint{}
		for ups := range ch {
			for _, up := range ups {
				switch up.Op {
				case endpoints.Add:
					current[up.Key] = up.Endpoint
				case endpoints.Delete:
					delete(current, up.Key)
				}
			}
			latest := make([]discovery.Endpoint, 0, len(current))
			for _, ep := range current {
				latest = append(latest, discovery.Endpoint{Addr: ep.Addr})
			}
			if handler != nil {
				handler(latest)
			}
		}
	}()
	e.watcherMux.Lock()
	e.watchers[key] = cancel
	e.watcherMux.Unlock()
	return cancel, nil
}

// NewEtcdRegister constructs a Registerer + Resolver backed by etcd.
func NewEtcdRegister(uri string) discovery.Registerer {
	return &etcdRegister{
		uri: uri,
	}
}
