package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
)

type WsOption func(scope *registry.Scope, mux *WebSocketMux) error

type _wsUpgraderWithLogger struct {
	logger logger.ILogger
}

func (v *_wsUpgraderWithLogger) upgraderHandler(upgrader *websocket.Upgrader, handler WebsocketHandler) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		v.logger.Infof("http -> Websocket %s %s %s",
			request.RemoteAddr, request.Method, request.URL.Path)
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			v.logger.WarnW("fail to upgrade http request to websocket",
				"err", err,
				"remoteAddr", request.RemoteAddr,
				"method", request.Method,
				"url", request.URL.Path)
			return
		}
		running := true
		msgHandler := handler(request, conn, func() {
			running = false
		})
		for running {
			if err = msgHandler(conn.ReadMessage()); err != nil {
				break
			}
		}
		err = conn.Close()
		v.logger.InfoW("connection closed",
			"remoteAddr", request.RemoteAddr,
			"url", request.URL.Path,
			"err", err)
	}
}

func WithLogger(log logger.ILogger) WsOption {
	return func(scope *registry.Scope, mux *WebSocketMux) error {
		value := &_wsUpgraderWithLogger{
			logger: registry.UseIn(scope, log).WithPrefix(scene.NewSceneImplNameNoVer("websocket", "router").Identifier()),
		}
		mux.UpgraderHandler = value.upgraderHandler
		return nil
	}
}

func WithCors() WsOption {
	return func(_ *registry.Scope, mux *WebSocketMux) error {
		mux.upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
		return nil
	}
}
