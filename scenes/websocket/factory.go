package websocket

import (
	"github.com/rhine-tech/scene"
	"strings"
)

type websocketFactory struct {
	mux IWebsocketMux
}

func NewFactory(mux IWebsocketMux) scene.ApplicationFactory[WebsocketApplication] {
	return &websocketFactory{
		mux: mux,
	}
}

func (w *websocketFactory) Name() string {
	return "scene.factory.websocket.gorilla"
}

func (w *websocketFactory) Create(app WebsocketApplication) error {
	prefix := app.Prefix()
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	return app.Create(w.mux.UsePrefix(prefix))
}

func (w *websocketFactory) Destroy(app WebsocketApplication) error {
	return nil
}
