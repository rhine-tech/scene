package gin

import "github.com/gin-gonic/gin"

func MiddlewareChain(middlewares ...gin.HandlerFunc) func(handler gin.HandlerFunc) []gin.HandlerFunc {
	return func(handler gin.HandlerFunc) []gin.HandlerFunc {
		handlers := make([]gin.HandlerFunc, 0, len(middlewares)+1)
		handlers = append(handlers, middlewares...)
		handlers = append(handlers, handler)
		return handlers
	}
}
