package websocket

import (
	"github.com/gorilla/websocket"
	"github.com/rhine-tech/scene"
	"net/http"
)

// type WebsocketConn *websocket.Conn

type WebsocketUpgraderHandler func(upgrader *websocket.Upgrader, handler WebsocketHandler) func(writer http.ResponseWriter, request *http.Request)

// WebsocketHandler is a function that handles websocket connections
// clos is a function that can be called to close the connection
type WebsocketHandler func(conn *websocket.Conn, clos func()) WebsocketMessageHandler

// WebsocketMessageHandler is a function that handles websocket messages
// pass err from ReadMessage to WebsocketMessageHandler
// so that the handler can decide what to do with the error
type WebsocketMessageHandler func(msgType int, msg []byte, err error) error

// IWebsocketMux is mux for websocket application
type IWebsocketMux interface {
	HandleFunc(pattern string, handler WebsocketHandler)
	UsePrefix(pattern string) IWebsocketMux
}

type WebsocketApplication interface {
	scene.Application
	Prefix() string
	Create(mux IWebsocketMux) error
}
