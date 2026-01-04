package arpc

import (
	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"net"
	"time"
)

type Client interface {
	scene.Named
	Logger() logger.ILogger
	Client() *arpc.Client
	Call(method string, req interface{}, rsp interface{}, timeout time.Duration, args ...interface{}) error
	AddConnectedHandler(func(c *arpc.Client))
}

type defaultClient struct {
	client              *arpc.Client
	dialer              arpc.DialerFunc
	options             []ClientOption
	onConnectedHandlers []func(c *arpc.Client)
	log                 logger.ILogger `aperture:""`
}

func (c *defaultClient) ImplName() scene.ImplName {
	return scene.NewSceneImplNameNoVer("arpc", "Client")
}

func (c *defaultClient) Setup() error {
	err := c.setupClient()
	if err != nil {
		c.log.Error("Setup client Error: %v", err)
	} else {
		c.log.Info("Setup arpc client success")
	}
	return nil
}

func (c *defaultClient) Logger() logger.ILogger {
	return c.log
}

func (c *defaultClient) Client() *arpc.Client {
	return c.client
}

func (c *defaultClient) setupClient() error {
	if c.client != nil {
		c.client.Stop()
		c.client = nil
	}
	client, err := arpc.NewClient(c.dialer)
	if err != nil {
		return err
	}
	client.Set("logger.ILogger", c.log)
	c.client = client
	for _, option := range c.options {
		if err = option(c); err != nil {
			return err
		}
	}
	client.Handler.SetLogTag("[Client]")
	client.Handler.HandleConnected(c.onConnectedHandler)
	return nil
}

func (c *defaultClient) Call(method string, req interface{}, rsp interface{}, timeout time.Duration, args ...interface{}) error {
	var err error
	if c.client == nil {
		err = c.setupClient()
		if err != nil {
			return err
		}
	}
	err = c.client.Call(method, req, rsp, timeout, args...)
	return err
}

func (c *defaultClient) AddConnectedHandler(f func(c *arpc.Client)) {
	c.onConnectedHandlers = append(c.onConnectedHandlers, f)
}

func (c *defaultClient) onConnectedHandler(client *arpc.Client) {
	for _, handler := range c.onConnectedHandlers {
		handler(client)
	}
}

func NewClient(network string, addr string, options ...ClientOption) Client {
	return &defaultClient{
		client: nil,
		dialer: func() (net.Conn, error) {
			return net.DialTimeout(network, addr, time.Second*3)
		},
		options: options,
	}
}
