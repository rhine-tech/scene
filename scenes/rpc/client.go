package rpc

import (
	"errors"
	"net/rpc"
)

// Client is a rpc client manage tcp connection
// it will establish connection when needed and
// reconnect if socket breaks.
type Client interface {
	Dial() error
	Call(serviceMethod string, args any, reply any) error
}

type Dialer interface {
	Dial() (*rpc.Client, error)
}

type tcpDialer struct {
	protocol string
	address  string
}

func (t *tcpDialer) Dial() (*rpc.Client, error) {
	return rpc.Dial(t.protocol, t.address)
}

type httpDialer struct {
	protocol string
	address  string
	path     string
}

func (t *httpDialer) Dial() (*rpc.Client, error) {
	return rpc.DialHTTPPath(t.protocol, t.address, t.path)
}

type defaultClient struct {
	dialer              Dialer
	client              *rpc.Client
	reconnectOnShutdown bool
}

func NewClient(network string, addr string) Client {
	return &defaultClient{
		dialer: &tcpDialer{
			protocol: network,
			address:  addr,
		},
		reconnectOnShutdown: true,
	}
}

func NewHttpClient(network string, addr string, path string) Client {
	return &defaultClient{
		dialer: &httpDialer{
			protocol: network,
			address:  addr,
			path:     path,
		},
		reconnectOnShutdown: true,
	}
}

func (c *defaultClient) Dial() error {
	if c.client != nil {
		_ = c.client.Close()
	}
	var err error
	c.client, err = c.dialer.Dial()
	return err
}

func (c *defaultClient) Call(serviceMethod string, args any, reply any) error {
	var err error
	if c.client == nil {
		err = c.Dial()
		if err != nil {
			return err
		}
	}
	err = c.client.Call(serviceMethod, args, reply)
	// try reconnect
	if errors.Is(err, rpc.ErrShutdown) && c.reconnectOnShutdown {
		err = c.Dial()
		// if reconnect errors, return error directly
		if err != nil {
			return err
		}
		// else call once more
		err = c.client.Call(serviceMethod, args, reply)
	}
	return err
}
