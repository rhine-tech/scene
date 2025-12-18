package arpc

import (
	"context"
	"github.com/google/uuid"
	"github.com/lesismal/arpc"
	"time"
)

type handlerProxy struct {
	arpc.Handler
	methods []string
}

func (h *handlerProxy) Handle(method string, handler arpc.HandlerFunc, args ...interface{}) {
	h.methods = append(h.methods, method)
	h.Handler.Handle(method, handler, args...)
}

func (h *handlerProxy) HandleStream(method string, handler arpc.StreamHandlerFunc, args ...interface{}) {
	h.methods = append(h.methods, method)
	h.Handler.HandleStream(method, handler, args...)
}

// UseNexusGateway turns the current server into a nexus gateway for proxy-based discovery mode.
// If holder pointers are provided, the created gateway will be written back for later use.
func UseNexusGateway(timeout time.Duration, holders ...**NexusGateway) ServerOption {
	return func(server *arpc.Server) error {
		gw := EnableNexus(server, timeout, nil)
		for _, h := range holders {
			if h != nil {
				*h = gw
			}
		}
		return nil
	}
}

func WithNexusGateway(apps ...ARpcApp) ClientOption {
	return func(client Client) error {
		// proxy hack to get registered method dynamically. only human can do. llm would
		// never come up with this
		proxy := &handlerProxy{Handler: client.Client().Handler, methods: make([]string, 0)}
		for _, app := range apps {
			err := app.RegisterService(proxy)
			if err != nil {
				return err
			}
		}
		id := uuid.NewString()
		client.AddConnectedHandler(func(c *arpc.Client) {
			ctx := context.Background()
			client.Logger().Infof("Nexus Gateway connected, register methods to nexus")
			for _, method := range proxy.methods {
				_, err := RegisterMethodViaNexus(ctx, client, RegisterMethodRequest{
					Method:     method,
					InstanceID: id,
					Metadata:   map[string]string{},
				}, -1)
				if err != nil {
					client.Logger().ErrorW("fail to register method %s to nexus", "error", err)
					continue
				}
				client.Logger().Infof("successfully register method %s to nexus", method)
			}
		})
		return nil
	}
}
