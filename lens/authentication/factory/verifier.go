package factory

import (
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/lens/authentication/service/loginstatus"
)

type JWTVerifier struct {
	Key    string
	Secret []byte
}

func (J JWTVerifier) Provide() authentication.HTTPLoginStatusVerifier {
	return loginstatus.NewJWT(J.Secret, J.Key)
}
