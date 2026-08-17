package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/lens/authentication/service/loginstatus"
	"github.com/rhine-tech/scene/registry"
)

type HttpVerifier scene.IModuleDependencyProvider[authentication.HTTPLoginStatusVerifier]

type scopedHttpVerifier interface {
	provideIn(*registry.Scope) authentication.HTTPLoginStatusVerifier
}

func provideHttpVerifier(scope *registry.Scope, verifier HttpVerifier) authentication.HTTPLoginStatusVerifier {
	if scoped, ok := verifier.(scopedHttpVerifier); ok {
		return scoped.provideIn(scope)
	}
	return verifier.Provide()
}

type JWTVerifier struct {
	Key    string
	Secret []byte
}

func (J JWTVerifier) Default() JWTVerifier {
	cfg := registry.Use[config.IConfig](nil)
	return JWTVerifier{
		Key:    "scene_token",
		Secret: []byte(cfg.GetString("authentication.jwt.secret")),
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
	return loginstatus.NewTokenAuth(nil, t.HeaderKey, t.QueryKey)
}

func (t TokenVerifier) provideIn(scope *registry.Scope) authentication.HTTPLoginStatusVerifier {
	return loginstatus.NewTokenAuth(
		registry.ProvideIn[authentication.IAccessTokenService](scope),
		t.HeaderKey,
		t.QueryKey,
	)
}
