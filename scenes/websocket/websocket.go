package websocket

import (
	"github.com/aynakeya/scene"
	"github.com/gorilla/websocket"
	"net/http"
)

//type WebsocketConn *websocket.Conn

type WebsocketHandler func(conn *websocket.Conn, clos func()) WebsocketMessageHandler
type WebsocketMessageHandler func(msgType int, msg []byte, err error) error

type WebsocketMux interface {
	http.Handler
	HandleFunc(pattern string, handler WebsocketHandler)
	UsePrefix(pattern string) WebsocketMux
}

type WebsocketApplication interface {
	scene.Application
	Prefix() string
	Create(mux WebsocketMux) error
}
