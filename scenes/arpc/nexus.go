package arpc

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
)

const (
	nexusRegisterMethod = "scene.arpc.nexus.register"
	nexusDiscoverMethod = "scene.arpc.nexus.discover"
)

type NexusGateway struct {
	mu          sync.RWMutex
	routes      map[string]*routeBucket
	byClient    map[*arpc.Client]*nexusPeer
	log         logger.ILogger
	defaultCall time.Duration
}

type routeBucket struct {
	peers map[string]*nexusPeer
	order []string
	seq   uint64
}

type nexusPeer struct {
	id      string
	method  string
	meta    map[string]string
	session *arpc.Client
}

type RegisterMethodRequest struct {
	Method     string            `json:"method"`
	InstanceID string            `json:"instance_id"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type RegisterServiceResponse struct {
	InstanceID string `json:"instance_id"`
}

type ProxyRequest struct {
	Method   string          `json:"method"`
	Payload  json.RawMessage `json:"payload"`
	TimeoutM int64           `json:"timeout_ms,omitempty"`
}

type ProxyResponse struct {
	Payload json.RawMessage `json:"payload,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type DiscoverResponse struct {
	Methods map[string][]ServiceInstance `json:"methods"`
}

type ServiceInstance struct {
	InstanceID string            `json:"instance_id"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// EnableNexus wires a server to behave as a nexus gateway.
// It registers internal handlers for registration/proxy/discovery.
func EnableNexus(server *arpc.Server, defaultTimeout time.Duration, log logger.ILogger) *NexusGateway {
	if log == nil {
		log = registry.Logger
	}
	n := &NexusGateway{
		routes:      map[string]*routeBucket{},
		byClient:    map[*arpc.Client]*nexusPeer{},
		log:         log.WithPrefix("arpc.nexus"),
		defaultCall: defaultTimeout,
	}
	if n.defaultCall <= 0 {
		n.defaultCall = 10 * time.Second
	}

	server.Handler.Handle(nexusRegisterMethod, n.handleRegister)
	server.Handler.Handle(nexusDiscoverMethod, n.handleDiscover)
	server.Handler.HandleNotFound(n.handleProxy)
	server.Handler.HandleDisconnected(n.onDisconnected)
	return n
}

func (n *NexusGateway) handleRegister(ctx *arpc.Context) {
	var req RegisterMethodRequest
	var rsp RegisterServiceResponse
	if err := ctx.Bind(&req); err != nil {
		ctx.Error(err)
		return
	}
	key := req.Method
	if key == "" {
		ctx.Error(errors.New("method is required"))
		return
	}
	if req.InstanceID == "" {
		req.InstanceID = uuid.NewString()
	}
	peer := &nexusPeer{
		id:      req.InstanceID,
		method:  key,
		meta:    req.Metadata,
		session: ctx.Client,
	}
	n.addPeer(peer)
	n.log.Infof("service registered via nexus: %s(%s)", key, req.InstanceID)
	rsp.InstanceID = req.InstanceID
	ctx.Write(&rsp)
}

func (n *NexusGateway) handleProxy(ctx *arpc.Context) {
	var req ProxyRequest
	var rsp ProxyResponse
	if err := ctx.Bind(&req); err != nil {
		ctx.Error(err)
		return
	}
	key := req.Method
	if req.Method == "" {
		rsp.Error = "routing method and method are required"
		ctx.Write(&rsp)
		return
	}
	target := n.pick(key)
	if target == nil {
		rsp.Error = "target service unavailable"
		ctx.Write(&rsp)
		return
	}

	timeout := n.defaultCall
	if req.TimeoutM > 0 {
		timeout = time.Duration(req.TimeoutM) * time.Millisecond
	}

	var raw json.RawMessage
	err := target.session.Call(req.Method, req.Payload, &raw, timeout)
	if err != nil {
		rsp.Error = err.Error()
		ctx.Write(&rsp)
		return
	}
	rsp.Payload = raw
	ctx.Write(&rsp)
}

func (n *NexusGateway) handleDiscover(ctx *arpc.Context) {
	resp := DiscoverResponse{
		Methods: map[string][]ServiceInstance{},
	}
	n.mu.RLock()
	for name, bucket := range n.routes {
		instances := make([]ServiceInstance, 0, len(bucket.peers))
		for _, peer := range bucket.peers {
			instances = append(instances, ServiceInstance{
				InstanceID: peer.id,
				Metadata:   peer.meta,
			})
		}
		resp.Methods[name] = instances
	}
	n.mu.RUnlock()
	ctx.Write(&resp)
}

func (n *NexusGateway) addPeer(peer *nexusPeer) {
	n.mu.Lock()
	defer n.mu.Unlock()
	bucket, ok := n.routes[peer.method]
	if !ok {
		bucket = &routeBucket{peers: map[string]*nexusPeer{}}
		n.routes[peer.method] = bucket
	}
	if _, exists := bucket.peers[peer.id]; !exists {
		bucket.order = append(bucket.order, peer.id)
	}
	bucket.peers[peer.id] = peer
	n.byClient[peer.session] = peer
}

func (n *NexusGateway) removePeer(peer *nexusPeer) {
	if peer == nil {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.byClient, peer.session)
	if bucket, ok := n.routes[peer.method]; ok {
		delete(bucket.peers, peer.id)
		filtered := make([]string, 0)
		for _, id := range bucket.order {
			if id != peer.id {
				filtered = append(filtered, id)
			}
		}
		bucket.order = filtered
		if len(bucket.peers) == 0 {
			delete(n.routes, peer.method)
		}
	}
}

func (n *NexusGateway) pick(service string) *nexusPeer {
	n.mu.RLock()
	defer n.mu.RUnlock()
	bucket := n.routes[service]
	if bucket == nil || len(bucket.order) == 0 {
		return nil
	}
	idx := atomic.AddUint64(&bucket.seq, 1)
	id := bucket.order[int(idx%uint64(len(bucket.order)))]
	return bucket.peers[id]
}

func (n *NexusGateway) onDisconnected(c *arpc.Client) {
	n.removePeer(n.byClient[c])
}

// RegisterMethodViaNexus registers the current service instance to a nexus gateway.
// If ctx is nil, context.Background will be used.
func RegisterMethodViaNexus(ctx context.Context, nexus Client, req RegisterMethodRequest, timeout time.Duration) (RegisterServiceResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if req.Method == "" {
		return RegisterServiceResponse{}, errors.New("method is required")
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	var rsp RegisterServiceResponse
	err := nexus.Call(nexusRegisterMethod, &req, &rsp, timeout)
	return rsp, err
}

// DiscoverViaNexus queries services currently attached to the gateway.
func DiscoverViaNexus(ctx context.Context, nexus Client, timeout time.Duration) (DiscoverResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	var rsp DiscoverResponse
	err := nexus.Call(nexusDiscoverMethod, struct{}{}, &rsp, timeout)
	return rsp, err
}
