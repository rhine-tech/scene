package factory

import (
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/lens/authentication/service/loginstatus"
	"github.com/rhine-tech/scene/registry"
)

type JWTVerifier struct {
	Key    string
	Secret []byte
}

func (J JWTVerifier) Default() JWTVerifier {
	return JWTVerifier{
		Key:    "scene_token",
		Secret: []byte(registry.Config.GetString("authentication.jwt.secret")),
	}
}

func (J JWTVerifier) Provide() authentication.HTTPLoginStatusVerifier {
	return loginstatus.NewJWT(J.Secret, J.Key)
}

type TokenVerifier struct {
	HeaderKey string
	QueryKey  string
}

func (t TokenVerifier) Default() TokenVerifier {
	return TokenVerifier{
		HeaderKey: "scene_token",
		QueryKey:  "scene_token",
	}
}

func (t TokenVerifier) Provide() authentication.HTTPLoginStatusVerifier {
	return registry.Load(loginstatus.NewTokenAuth(nil, t.HeaderKey, t.QueryKey))
}
