package scene

type HttpRouteInfo struct {
	Method string
	Path   string
}

type HttpRoute interface {
	GetRoute() HttpRouteInfo
}
