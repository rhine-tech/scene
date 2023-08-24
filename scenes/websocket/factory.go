package websocket

import (
	"github.com/aynakeya/scene"
)

type websocketFactory struct {
	mux WebsocketMux
}

func NewFactory(mux WebsocketMux) scene.ApplicationFactory[WebsocketApplication] {
	return &websocketFactory{
		mux: mux,
	}
}

func (w *websocketFactory) Name() string {
	return "scene.factory.websocket.gorilla"
}

func (w *websocketFactory) Create(app WebsocketApplication) error {
	return app.Create(w.mux.UsePrefix("/" + app.Prefix()))
}

func (w *websocketFactory) Destroy(app WebsocketApplication) error {
	return nil
}
