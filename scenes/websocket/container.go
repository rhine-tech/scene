package websocket

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	scommon "github.com/rhine-tech/scene/scenes/common"
)

type websocketContainer struct {
	*scommon.HttpAppContainer[WebsocketApplication]
}

func (g *websocketContainer) Name() scene.ImplName {
	return scene.NewSceneImplNameNoVer("websocket")
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
