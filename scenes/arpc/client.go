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
	Client() *arpc.Client
	Call(method string, req interface{}, rsp interface{}, timeout time.Duration, args ...interface{}) error
}

type defaultClient struct {
	client  *arpc.Client
	dialer  arpc.DialerFunc
	options []ClientOption
	log     logger.ILogger `aperture:""`
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
	for _, option := range c.options {
		if err = option(client); err != nil {
			return err
		}
	}
	client.Handler.SetLogTag("[Client]")
	c.client = client
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

func NewClient(network string, addr string, options ...ClientOption) Client {
	return &defaultClient{
		client: nil,
		dialer: func() (net.Conn, error) {
			return net.DialTimeout(network, addr, time.Second*3)
		},
		options: options,
	}
}
