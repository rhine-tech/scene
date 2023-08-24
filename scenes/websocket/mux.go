package websocket

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
)

type webSocketMux struct {
	router   *mux.Router
	upgrader websocket.Upgrader
}

func (w *webSocketMux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	w.router.ServeHTTP(writer, request)
}

func (w *webSocketMux) UsePrefix(pattern string) WebsocketMux {
	return &webSocketMux{
		router:   w.router.PathPrefix(pattern).Subrouter(),
		upgrader: w.upgrader,
	}
}

func NewWebSocketMux() WebsocketMux {
	return &webSocketMux{
		router:   mux.NewRouter(),
		upgrader: websocket.Upgrader{},
	}
}

func (w *webSocketMux) HandleFunc(pattern string, handler WebsocketHandler) {
	w.router.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		conn, err := w.upgrader.Upgrade(writer, request, nil)
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
	})
}
