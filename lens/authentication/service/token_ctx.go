package service

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/model"
)

type accessTokenServiceCtx struct {
	s   *accessTokenService
	ctx scene.Context
}

func (a *accessTokenServiceCtx) SrvImplName() scene.ImplName {
	return a.s.SrvImplName()
}

func (a *accessTokenServiceCtx) WithSceneContext(ctx scene.Context) authentication.IAccessTokenService {
	return a.s.WithSceneContext(ctx)
}

func (a *accessTokenServiceCtx) Create(userId, name string, expireAt int64) (authentication.AccessToken, error) {
	return a.s.Create(userId, name, expireAt)
}

func (a *accessTokenServiceCtx) ListByUser(userId string, offset, limit int64) (model.PaginationResult[authentication.AccessToken], error) {
	return a.s.ListByUser(userId, offset, limit)
}

func (a *accessTokenServiceCtx) List(offset, limit int64) (model.PaginationResult[authentication.AccessToken], error) {
	return a.s.List(offset, limit)
}

func (a *accessTokenServiceCtx) Validate(token string) (userId string, valid bool, err error) {
	return a.s.Validate(token)
}
