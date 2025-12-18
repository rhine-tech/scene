package arpc

import "github.com/lesismal/arpc"

type recordableHandler struct {
	arpc.Handler
	record func(method string)
}

func (h *recordableHandler) Handle(method string, handler arpc.HandlerFunc, args ...interface{}) {
	if h.record != nil {
		h.record(method)
	}
	h.Handler.Handle(method, handler, args...)
}

func (h *recordableHandler) HandleStream(method string, handler arpc.StreamHandlerFunc, args ...interface{}) {
	if h.record != nil {
		h.record(method)
	}
	h.Handler.HandleStream(method, handler, args...)
}
