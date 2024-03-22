package websocket

import (
	"github.com/gorilla/websocket"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"net/http"
)

type WsOption func(mux *WebSocketMux) error

type _wsUpgraderWithLogger struct {
	logger logger.ILogger
}

func (v *_wsUpgraderWithLogger) upgraderHandler(upgrader *websocket.Upgrader, handler WebsocketHandler) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		v.logger.Infof("%s %s %s",
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
		msgHandler := handler(conn, func() {
			running = false
		})
		for running {
			if err = msgHandler(conn.ReadMessage()); err != nil {
				break
			}
		}
		err = conn.Close()
		v.logger.Infof("connection closed",
			"remoteAddr", request.RemoteAddr,
			"method", request.Method,
			"url", request.URL.Path,
			"err", err)
	}
}

func WithLogger(logger logger.ILogger) WsOption {
	logger = registry.Use(logger)
	value := &_wsUpgraderWithLogger{
		logger: logger.WithPrefix(scene.NewSceneImplNameNoVer("websocket", "router").Identifier()),
	}
	return func(mux *WebSocketMux) error {
		mux.UpgraderHandler = value.upgraderHandler
		return nil
	}
}

func WithCors() WsOption {
	return func(mux *WebSocketMux) error {
		mux.upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
		return nil
	}
}
