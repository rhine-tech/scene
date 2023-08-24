package websocket

import (
	"github.com/aynakeya/scene"
	"github.com/aynakeya/scene/lens/infrastructure/logger"
	"github.com/aynakeya/scene/registry"
	scommon "github.com/aynakeya/scene/scenes/common"
)

type websocketContainer struct {
	*scommon.HttpAppContainer[WebsocketApplication]
}

func (g *websocketContainer) Name() string {
	return "scene.app-container.websocket"
}

func NewContainer(addr string, apps ...WebsocketApplication) scene.ApplicationContainer {
	mux := NewWebSocketMux()
	return &websocketContainer{scommon.NewHttpAppContainer(
		scommon.NewAppManager(apps...),
		NewFactory(mux),
		registry.AcquireSingleton(logger.ILogger(nil)).WithPrefix("scene.app-container.websocket"),
		addr,
		mux,
	)}
}
