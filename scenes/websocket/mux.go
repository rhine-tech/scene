package websocket

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
)

// WebsocketUpgraderHandler its the default WebsocketUpgraderHandler function for WebsocketMux
func defaultWebsocketUpgraderHandler(upgrader *websocket.Upgrader, handler WebsocketHandler) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
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
		_ = conn.Close()
	}
}

type WebSocketMux struct {
	router          *mux.Router
	upgrader        websocket.Upgrader
	UpgraderHandler WebsocketUpgraderHandler
}

func NewWebSocketMux() *WebSocketMux {
	return &WebSocketMux{
		router:          mux.NewRouter(),
		upgrader:        websocket.Upgrader{},
		UpgraderHandler: defaultWebsocketUpgraderHandler,
	}
}

func (w *WebSocketMux) createWsUpgrader(handler WebsocketHandler) func(writer http.ResponseWriter, request *http.Request) {
	return w.UpgraderHandler(&w.upgrader, handler)
}

func (w *WebSocketMux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	w.router.ServeHTTP(writer, request)
}

func (w *WebSocketMux) UsePrefix(pattern string) IWebsocketMux {
	return &webSocketSubRouter{
		root:   w,
		router: w.router.PathPrefix(pattern).Subrouter(),
	}
}

func (w *WebSocketMux) HandleFunc(pattern string, handler WebsocketHandler) {
	w.router.HandleFunc(pattern, w.createWsUpgrader(handler))
}

type webSocketSubRouter struct {
	root   *WebSocketMux
	router *mux.Router
}

func (w *webSocketSubRouter) UsePrefix(pattern string) IWebsocketMux {
	return &webSocketSubRouter{
		root:   w.root,
		router: w.router.PathPrefix(pattern).Subrouter(),
	}
}

func (w *webSocketSubRouter) HandleFunc(pattern string, handler WebsocketHandler) {
	w.router.HandleFunc(pattern, w.root.createWsUpgrader(handler))
}
