package arpcimpl

import (
	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/model"
	sarpc "github.com/rhine-tech/scene/scenes/arpc"
	"time"
)

// Method definition

const (
	ARpcNameAuthenticationIAccessTokenServiceCreate     = "authentication.IAccessTokenService.Create"
	ARpcNameAuthenticationIAccessTokenServiceListByUser = "authentication.IAccessTokenService.ListByUser"
	ARpcNameAuthenticationIAccessTokenServiceList       = "authentication.IAccessTokenService.List"
	ARpcNameAuthenticationIAccessTokenServiceValidate   = "authentication.IAccessTokenService.Validate"
	ARpcNameAuthenticationIAccessTokenServiceDelete     = "authentication.IAccessTokenService.Delete"
)

type IAccessTokenServiceCreateArgs struct {
	Val0 string
	Val1 string
	Val2 int64
}

type IAccessTokenServiceCreateResult struct {
	Val0 authentication.AccessToken
	Val1 error
}
type IAccessTokenServiceListByUserArgs struct {
	Val0 string
	Val1 int64
	Val2 int64
}

type IAccessTokenServiceListByUserResult struct {
	Val0 model.PaginationResult[authentication.AccessToken]
	Val1 error
}
type IAccessTokenServiceListArgs struct {
	Val0 int64
	Val1 int64
}

type IAccessTokenServiceListResult struct {
	Val0 model.PaginationResult[authentication.AccessToken]
	Val1 error
}
type IAccessTokenServiceValidateArgs struct {
	Val0 string
}

type IAccessTokenServiceValidateResult struct {
	Val0 string
	Val1 bool
	Val2 error
}
type IAccessTokenServiceDeleteArgs struct {
	Val0 string
}

type IAccessTokenServiceDeleteResult struct {
	Val0 error
}

// Service (Client) Implementation

type arpcClientIAccessTokenService struct {
	client  sarpc.Client `aperture:""`
	timeout time.Duration
	log     logger.ILogger `aperture:""`
}

func NewARpcIAccessTokenService(client sarpc.Client) authentication.IAccessTokenService {
	return &arpcClientIAccessTokenService{
		client:  client,
		timeout: time.Second * 5,
	}
}

func NewARpcIAccessTokenServiceWithTimeout(client sarpc.Client, timeout time.Duration) authentication.IAccessTokenService {
	return &arpcClientIAccessTokenService{
		client:  client,
		timeout: timeout,
	}
}

func (r *arpcClientIAccessTokenService) SrvImplName() scene.ImplName {
	return authentication.Lens.ImplName("IAccessTokenService", "arpc")
}

// Deprecated: no longer used
func (r *arpcClientIAccessTokenService) WithSceneContext(ctx scene.Context) authentication.IAccessTokenService {
	return r
}

func (r *arpcClientIAccessTokenService) Create(userId string, name string, expireAt int64) (authentication.AccessToken, error) {
	var resp IAccessTokenServiceCreateResult
	err := r.client.Call(ARpcNameAuthenticationIAccessTokenServiceCreate, &IAccessTokenServiceCreateArgs{
		Val0: userId,
		Val1: name,
		Val2: expireAt,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameAuthenticationIAccessTokenServiceCreate, "err", err)
		return *new(authentication.AccessToken), err
	}
	return resp.Val0, resp.Val1
}
func (r *arpcClientIAccessTokenService) ListByUser(userId string, offset int64, limit int64) (model.PaginationResult[authentication.AccessToken], error) {
	var resp IAccessTokenServiceListByUserResult
	err := r.client.Call(ARpcNameAuthenticationIAccessTokenServiceListByUser, &IAccessTokenServiceListByUserArgs{
		Val0: userId,
		Val1: offset,
		Val2: limit,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameAuthenticationIAccessTokenServiceListByUser, "err", err)
		return *new(model.PaginationResult[authentication.AccessToken]), err
	}
	return resp.Val0, resp.Val1
}
func (r *arpcClientIAccessTokenService) List(offset int64, limit int64) (model.PaginationResult[authentication.AccessToken], error) {
	var resp IAccessTokenServiceListResult
	err := r.client.Call(ARpcNameAuthenticationIAccessTokenServiceList, &IAccessTokenServiceListArgs{
		Val0: offset,
		Val1: limit,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameAuthenticationIAccessTokenServiceList, "err", err)
		return *new(model.PaginationResult[authentication.AccessToken]), err
	}
	return resp.Val0, resp.Val1
}
func (r *arpcClientIAccessTokenService) Validate(token string) (string, bool, error) {
	var resp IAccessTokenServiceValidateResult
	err := r.client.Call(ARpcNameAuthenticationIAccessTokenServiceValidate, &IAccessTokenServiceValidateArgs{
		Val0: token,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameAuthenticationIAccessTokenServiceValidate, "err", err)
		return *new(string), *new(bool), err
	}
	return resp.Val0, resp.Val1, resp.Val2
}
func (r *arpcClientIAccessTokenService) Delete(token string) error {
	var resp IAccessTokenServiceDeleteResult
	err := r.client.Call(ARpcNameAuthenticationIAccessTokenServiceDelete, &IAccessTokenServiceDeleteArgs{
		Val0: token,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameAuthenticationIAccessTokenServiceDelete, "err", err)
		return err
	}
	return resp.Val0
}

// Server Implementation

type ARpcServerIAccessTokenService struct {
	srv authentication.IAccessTokenService `aperture:""`
}

func HandleIAccessTokenService(srv authentication.IAccessTokenService, handler arpc.Handler) {
	svr := NewARpcServerIAccessTokenService(srv)
	HandleARpcServerIAccessTokenService(svr, handler)
}

func HandleARpcServerIAccessTokenService(svr *ARpcServerIAccessTokenService, handler arpc.Handler) {
	handler.Handle(ARpcNameAuthenticationIAccessTokenServiceCreate, svr.Create)
	handler.Handle(ARpcNameAuthenticationIAccessTokenServiceListByUser, svr.ListByUser)
	handler.Handle(ARpcNameAuthenticationIAccessTokenServiceList, svr.List)
	handler.Handle(ARpcNameAuthenticationIAccessTokenServiceValidate, svr.Validate)
	handler.Handle(ARpcNameAuthenticationIAccessTokenServiceDelete, svr.Delete)
}

func NewARpcServerIAccessTokenService(srv authentication.IAccessTokenService) *ARpcServerIAccessTokenService {
	return &ARpcServerIAccessTokenService{
		srv: srv,
	}
}
func (r *ARpcServerIAccessTokenService) Create(c *arpc.Context) {
	var req IAccessTokenServiceCreateArgs
	var resp IAccessTokenServiceCreateResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	a0, a1 := r.srv.Create(
		req.Val0,
		req.Val1,
		req.Val2,
	)
	resp.Val0 = a0
	resp.Val1 = a1
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerIAccessTokenService) ListByUser(c *arpc.Context) {
	var req IAccessTokenServiceListByUserArgs
	var resp IAccessTokenServiceListByUserResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	a0, a1 := r.srv.ListByUser(
		req.Val0,
		req.Val1,
		req.Val2,
	)
	resp.Val0 = a0
	resp.Val1 = a1
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerIAccessTokenService) List(c *arpc.Context) {
	var req IAccessTokenServiceListArgs
	var resp IAccessTokenServiceListResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	a0, a1 := r.srv.List(
		req.Val0,
		req.Val1,
	)
	resp.Val0 = a0
	resp.Val1 = a1
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerIAccessTokenService) Validate(c *arpc.Context) {
	var req IAccessTokenServiceValidateArgs
	var resp IAccessTokenServiceValidateResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	a0, a1, a2 := r.srv.Validate(
		req.Val0,
	)
	resp.Val0 = a0
	resp.Val1 = a1
	resp.Val2 = a2
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerIAccessTokenService) Delete(c *arpc.Context) {
	var req IAccessTokenServiceDeleteArgs
	var resp IAccessTokenServiceDeleteResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	a0 := r.srv.Delete(
		req.Val0,
	)
	resp.Val0 = a0
	_ = c.Write(&resp)
	return
}

// Scene App Definition

type ARpcAppIAccessTokenService struct {
	srv authentication.IAccessTokenService `aperture:""`
}

func (r *ARpcAppIAccessTokenService) Name() scene.ImplName {
	return authentication.Lens.ImplNameNoVer("ARpcApplication")
}

func (r *ARpcAppIAccessTokenService) RegisterService(server *arpc.Server) error {
	HandleIAccessTokenService(r.srv, server.Handler)
	return nil
}
