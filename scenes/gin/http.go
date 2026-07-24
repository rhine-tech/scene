package gin

import "net/http"

// HTTP method bitmap flags used by HttpRouteInfo.
const (
	HttpMethodGet     uint16 = 0b1
	HttpMethodHead    uint16 = 0b10
	HttpMethodPost    uint16 = 0b100
	HttpMethodPut     uint16 = 0b1000
	HttpMethodPatch   uint16 = 0b10000 // RFC 5789
	HttpMethodDelete  uint16 = 0b100000
	HttpMethodConnect uint16 = 0b1000000
	HttpMethodOptions uint16 = 0b10000000
	HttpMethodTrace   uint16 = 0b100000000
	httpMethodAll     uint16 = HttpMethodTrace<<1 - 1
)

var httpMethods = [...]struct {
	bitmap uint16
	name   string
}{
	{HttpMethodGet, http.MethodGet},
	{HttpMethodHead, http.MethodHead},
	{HttpMethodPost, http.MethodPost},
	{HttpMethodPut, http.MethodPut},
	{HttpMethodPatch, http.MethodPatch},
	{HttpMethodDelete, http.MethodDelete},
	{HttpMethodConnect, http.MethodConnect},
	{HttpMethodOptions, http.MethodOptions},
	{HttpMethodTrace, http.MethodTrace},
}

// HttpRouteInfo declares the methods and relative path of an action.
type HttpRouteInfo struct {
	Methods uint16
	Path    string
}

// HttpRoute provides route metadata for an Action.
type HttpRoute interface {
	GetRoute() HttpRouteInfo
}
