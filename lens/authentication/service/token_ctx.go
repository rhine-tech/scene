package service

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/lens/permission"
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
	pctx, ok := permission.GetPermContext(a.ctx)
	if !ok {
		a.s.logger.WarnW("permission denied: no permission context found", "operation", "CreateToken")
		return authentication.AccessToken{}, permission.ErrPermissionDenied
	}

	if !pctx.HasPermission(authentication.PermTokenCreate) {
		a.s.logger.WarnW("permission denied: missing required permission", "operation", "CreateToken", "required_perm", authentication.PermTokenCreate)
		return authentication.AccessToken{}, permission.ErrPermissionDenied
	}

	// For non-admins, verify they are creating a token for themselves.
	if !pctx.HasPermission(authentication.PermAdmin) {
		actx, ok := authentication.GetAuthContext(a.ctx)
		if !ok {
			a.s.logger.WarnW("permission denied: user not logged in for self-service action", "operation", "CreateToken", "target_user", userId)
			return authentication.AccessToken{}, authentication.ErrNotLogin
		}
		if actx.UserID != userId {
			a.s.logger.WarnW("permission denied: user attempting to create token for another user", "operation", "CreateToken", "requesting_user", actx.UserID, "target_user", userId)
			return authentication.AccessToken{}, permission.ErrPermissionDenied
		}
	}

	a.s.logger.DebugW("permission granted", "operation", "CreateToken", "target_user", userId)
	return a.s.Create(userId, name, expireAt)
}

func (a *accessTokenServiceCtx) ListByUser(userId string, offset, limit int64) (result model.PaginationResult[authentication.AccessToken], err error) {
	pctx, ok := permission.GetPermContext(a.ctx)
	if !ok {
		a.s.logger.WarnW("permission denied: no permission context found", "operation", "ListTokensByUser")
		return result, permission.ErrPermissionDenied
	}

	if !pctx.HasPermission(authentication.PermTokenList) {
		a.s.logger.WarnW("permission denied: missing required permission", "operation", "ListTokensByUser", "required_perm", authentication.PermTokenList)
		return result, permission.ErrPermissionDenied
	}

	if !pctx.HasPermission(authentication.PermAdmin) {
		actx, ok := authentication.GetAuthContext(a.ctx)
		if !ok {
			a.s.logger.WarnW("permission denied: user not logged in for self-service action", "operation", "ListTokensByUser", "target_user", userId)
			return result, authentication.ErrNotLogin
		}
		if actx.UserID != userId {
			a.s.logger.WarnW("permission denied: user attempting to list tokens for another user", "operation", "ListTokensByUser", "requesting_user", actx.UserID, "target_user", userId)
			return result, permission.ErrPermissionDenied
		}
	}

	a.s.logger.DebugW("permission granted", "operation", "ListTokensByUser", "target_user", userId)
	return a.s.ListByUser(userId, offset, limit)
}

func (a *accessTokenServiceCtx) List(offset, limit int64) (result model.PaginationResult[authentication.AccessToken], err error) {
	pctx, ok := permission.GetPermContext(a.ctx)
	if !ok {
		return result, permission.ErrPermissionDenied
	}
	if !pctx.HasPermission(authentication.PermAdmin) {
		return result, permission.ErrPermissionDenied
	}
	return a.s.List(offset, limit)
}

func (a *accessTokenServiceCtx) Validate(token string) (userId string, valid bool, err error) {
	// no permission
	return a.s.Validate(token)
}

func (a *accessTokenServiceCtx) Delete(token string) error {
	pctx, ok := permission.GetPermContext(a.ctx)
	if !ok {
		a.s.logger.WarnW("permission denied: no permission context found", "operation", "DeleteToken")
		return permission.ErrPermissionDenied
	}

	if !pctx.HasPermission(authentication.PermTokenDelete) {
		a.s.logger.WarnW("permission denied: missing required permission", "operation", "DeleteToken", "required_perm", authentication.PermTokenDelete)
		return permission.ErrPermissionDenied
	}

	// For non-admins, verify they own the token they are trying to delete.
	if !pctx.HasPermission(authentication.PermAdmin) {
		actx, ok := authentication.GetAuthContext(a.ctx)
		if !ok {
			a.s.logger.WarnW("permission denied: user not logged in for self-service action", "operation", "DeleteToken")
			return authentication.ErrNotLogin
		}

		// CRITICAL FIX: Check the result of Validate before making decisions.
		ownerId, _, err := a.s.Validate(token)
		if err != nil {
			// Propagate validation errors like "token expired"
			a.s.logger.InfoW("token validation failed during delete check", "operation", "DeleteToken", "requesting_user", actx.UserID, "error", err)
			return err
		}
		// Finally, check ownership.
		if actx.UserID != ownerId || !actx.IsLogin() {
			a.s.logger.ErrorW("permission denied: user attempted to delete another user's token", "operation", "DeleteToken", "requesting_user", actx.UserID, "token_owner", ownerId)
			return permission.ErrPermissionDenied
		}
	}

	a.s.logger.DebugW("permission granted", "operation", "DeleteToken")
	return a.s.Delete(token)
}
