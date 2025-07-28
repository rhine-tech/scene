package factory

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene/lens/authentication"
	authMw "github.com/rhine-tech/scene/lens/authentication/delivery/middleware"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

func GinWithAuthContext(verifiers ...HttpVerifier) sgin.GinOption {
	return func(engine *gin.Engine) error {
		HttpVerifiers := make([]authentication.HTTPLoginStatusVerifier, len(verifiers))
		for i, verifier := range verifiers {
			HttpVerifiers[i] = verifier.Provide()
		}
		engine.Use(authMw.GinAuthContext(HttpVerifiers...))
		return nil
	}
}
