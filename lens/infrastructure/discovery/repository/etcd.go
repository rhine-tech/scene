package repository

import (
	"context"
	"fmt"
	"github.com/rhine-tech/scene/lens/infrastructure/discovery"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
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
}

func (e *etcdRegister) Dispose() error {
	for _, node := range e.nodes {
		err := e.Deregister(node)
		if err != nil {
			e.log.Warnf("deregister node %s error: %s", node, err)
		}
	}
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
