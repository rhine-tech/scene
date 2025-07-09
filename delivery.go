package scene

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
)

// HttpMethod convert http.Method to uint16 version
func HttpMethod(method string) uint16 {
	switch method {
	case "GET":
		return HttpMethodGet
	case "HEAD":
		return HttpMethodHead
	case "POST":
		return HttpMethodPost
	case "PUT":
		return HttpMethodPut
	case "PATCH":
		return HttpMethodPatch
	case "DELETE":
		return HttpMethodDelete
	case "CONNECT":
		return HttpMethodConnect
	case "OPTIONS":
		return HttpMethodOptions
	case "TRACE":
		return HttpMethodTrace
	default:
		return 0
	}
}

type HttpRouteInfo struct {
	Method  string // Method specify a single method, might be deprecated in the future
	Methods uint16 // Methods aim to handle same route with multiple method
	Path    string
}

type HttpRoute interface {
	GetRoute() HttpRouteInfo
}
