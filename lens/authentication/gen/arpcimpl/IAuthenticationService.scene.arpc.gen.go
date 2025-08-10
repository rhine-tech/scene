package arpcimpl

import (
	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/authentication"
	sarpc "github.com/rhine-tech/scene/scenes/arpc"
	"time"
)

// Method definition

const (
	ARpcNameAuthenticationIAuthenticationServiceAddUser             = "authentication.IAuthenticationService.AddUser"
	ARpcNameAuthenticationIAuthenticationServiceDeleteUser          = "authentication.IAuthenticationService.DeleteUser"
	ARpcNameAuthenticationIAuthenticationServiceUpdateUser          = "authentication.IAuthenticationService.UpdateUser"
	ARpcNameAuthenticationIAuthenticationServiceAuthenticate        = "authentication.IAuthenticationService.Authenticate"
	ARpcNameAuthenticationIAuthenticationServiceAuthenticateByToken = "authentication.IAuthenticationService.AuthenticateByToken"
	ARpcNameAuthenticationIAuthenticationServiceHasUser             = "authentication.IAuthenticationService.HasUser"
	ARpcNameAuthenticationIAuthenticationServiceUserById            = "authentication.IAuthenticationService.UserById"
	ARpcNameAuthenticationIAuthenticationServiceUserByName          = "authentication.IAuthenticationService.UserByName"
	ARpcNameAuthenticationIAuthenticationServiceUserByEmail         = "authentication.IAuthenticationService.UserByEmail"
)

type IAuthenticationServiceAddUserArgs struct {
	Val0 string
	Val1 string
}

type IAuthenticationServiceAddUserResult struct {
	Val0 authentication.User
	Val1 error
}
type IAuthenticationServiceDeleteUserArgs struct {
	Val0 string
}

type IAuthenticationServiceDeleteUserResult struct {
	Val0 error
}
type IAuthenticationServiceUpdateUserArgs struct {
	Val0 authentication.User
}

type IAuthenticationServiceUpdateUserResult struct {
	Val0 error
}
type IAuthenticationServiceAuthenticateArgs struct {
	Val0 string
	Val1 string
}

type IAuthenticationServiceAuthenticateResult struct {
	Val0 string
	Val1 error
}
type IAuthenticationServiceAuthenticateByTokenArgs struct {
	Val0 string
}

type IAuthenticationServiceAuthenticateByTokenResult struct {
	Val0 string
	Val1 error
}
type IAuthenticationServiceHasUserArgs struct {
	Val0 string
}

type IAuthenticationServiceHasUserResult struct {
	Val0 bool
	Val1 error
}
type IAuthenticationServiceUserByIdArgs struct {
	Val0 string
}

type IAuthenticationServiceUserByIdResult struct {
	Val0 authentication.User
	Val1 error
}
type IAuthenticationServiceUserByNameArgs struct {
	Val0 string
}

type IAuthenticationServiceUserByNameResult struct {
	Val0 authentication.User
	Val1 error
}
type IAuthenticationServiceUserByEmailArgs struct {
	Val0 string
}

type IAuthenticationServiceUserByEmailResult struct {
	Val0 authentication.User
	Val1 error
}

// Service (Client) Implementation

type arpcClientIAuthenticationService struct {
	client  sarpc.Client `aperture:""`
	timeout time.Duration
	log     logger.ILogger `aperture:""`
}

func NewARpcIAuthenticationService(client sarpc.Client) authentication.IAuthenticationService {
	return &arpcClientIAuthenticationService{
		client:  client,
		timeout: time.Second * 5,
	}
}

func NewARpcIAuthenticationServiceWithTimeout(client sarpc.Client, timeout time.Duration) authentication.IAuthenticationService {
	return &arpcClientIAuthenticationService{
		client:  client,
		timeout: timeout,
	}
}

func (r *arpcClientIAuthenticationService) SrvImplName() scene.ImplName {
	return authentication.Lens.ImplName("IAuthenticationService", "arpc")
}

// Deprecated: no longer used
func (r *arpcClientIAuthenticationService) WithSceneContext(ctx scene.Context) authentication.IAuthenticationService {
	return r
}

func (r *arpcClientIAuthenticationService) AddUser(username string, password string) (authentication.User, error) {
	var resp IAuthenticationServiceAddUserResult
	err := r.client.Call(ARpcNameAuthenticationIAuthenticationServiceAddUser, &IAuthenticationServiceAddUserArgs{
		Val0: username,
		Val1: password,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameAuthenticationIAuthenticationServiceAddUser, "err", err)
		return *new(authentication.User), err
	}
	return resp.Val0, resp.Val1
}
func (r *arpcClientIAuthenticationService) DeleteUser(userId string) error {
	var resp IAuthenticationServiceDeleteUserResult
	err := r.client.Call(ARpcNameAuthenticationIAuthenticationServiceDeleteUser, &IAuthenticationServiceDeleteUserArgs{
		Val0: userId,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameAuthenticationIAuthenticationServiceDeleteUser, "err", err)
		return err
	}
	return resp.Val0
}
func (r *arpcClientIAuthenticationService) UpdateUser(user authentication.User) error {
	var resp IAuthenticationServiceUpdateUserResult
	err := r.client.Call(ARpcNameAuthenticationIAuthenticationServiceUpdateUser, &IAuthenticationServiceUpdateUserArgs{
		Val0: user,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameAuthenticationIAuthenticationServiceUpdateUser, "err", err)
		return err
	}
	return resp.Val0
}
func (r *arpcClientIAuthenticationService) Authenticate(username string, password string) (string, error) {
	var resp IAuthenticationServiceAuthenticateResult
	err := r.client.Call(ARpcNameAuthenticationIAuthenticationServiceAuthenticate, &IAuthenticationServiceAuthenticateArgs{
		Val0: username,
		Val1: password,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameAuthenticationIAuthenticationServiceAuthenticate, "err", err)
		return *new(string), err
	}
	return resp.Val0, resp.Val1
}
func (r *arpcClientIAuthenticationService) AuthenticateByToken(token string) (string, error) {
	var resp IAuthenticationServiceAuthenticateByTokenResult
	err := r.client.Call(ARpcNameAuthenticationIAuthenticationServiceAuthenticateByToken, &IAuthenticationServiceAuthenticateByTokenArgs{
		Val0: token,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameAuthenticationIAuthenticationServiceAuthenticateByToken, "err", err)
		return *new(string), err
	}
	return resp.Val0, resp.Val1
}
func (r *arpcClientIAuthenticationService) HasUser(userId string) (bool, error) {
	var resp IAuthenticationServiceHasUserResult
	err := r.client.Call(ARpcNameAuthenticationIAuthenticationServiceHasUser, &IAuthenticationServiceHasUserArgs{
		Val0: userId,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameAuthenticationIAuthenticationServiceHasUser, "err", err)
		return *new(bool), err
	}
	return resp.Val0, resp.Val1
}
func (r *arpcClientIAuthenticationService) UserById(userId string) (authentication.User, error) {
	var resp IAuthenticationServiceUserByIdResult
	err := r.client.Call(ARpcNameAuthenticationIAuthenticationServiceUserById, &IAuthenticationServiceUserByIdArgs{
		Val0: userId,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameAuthenticationIAuthenticationServiceUserById, "err", err)
		return *new(authentication.User), err
	}
	return resp.Val0, resp.Val1
}
func (r *arpcClientIAuthenticationService) UserByName(username string) (authentication.User, error) {
	var resp IAuthenticationServiceUserByNameResult
	err := r.client.Call(ARpcNameAuthenticationIAuthenticationServiceUserByName, &IAuthenticationServiceUserByNameArgs{
		Val0: username,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameAuthenticationIAuthenticationServiceUserByName, "err", err)
		return *new(authentication.User), err
	}
	return resp.Val0, resp.Val1
}
func (r *arpcClientIAuthenticationService) UserByEmail(email string) (authentication.User, error) {
	var resp IAuthenticationServiceUserByEmailResult
	err := r.client.Call(ARpcNameAuthenticationIAuthenticationServiceUserByEmail, &IAuthenticationServiceUserByEmailArgs{
		Val0: email,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameAuthenticationIAuthenticationServiceUserByEmail, "err", err)
		return *new(authentication.User), err
	}
	return resp.Val0, resp.Val1
}

// Server Implementation

type ARpcServerIAuthenticationService struct {
	srv authentication.IAuthenticationService `aperture:""`
}

func HandleIAuthenticationService(srv authentication.IAuthenticationService, handler arpc.Handler) {
	svr := NewARpcServerIAuthenticationService(srv)
	HandleARpcServerIAuthenticationService(svr, handler)
}

func HandleARpcServerIAuthenticationService(svr *ARpcServerIAuthenticationService, handler arpc.Handler) {
	handler.Handle(ARpcNameAuthenticationIAuthenticationServiceAddUser, svr.AddUser)
	handler.Handle(ARpcNameAuthenticationIAuthenticationServiceDeleteUser, svr.DeleteUser)
	handler.Handle(ARpcNameAuthenticationIAuthenticationServiceUpdateUser, svr.UpdateUser)
	handler.Handle(ARpcNameAuthenticationIAuthenticationServiceAuthenticate, svr.Authenticate)
	handler.Handle(ARpcNameAuthenticationIAuthenticationServiceAuthenticateByToken, svr.AuthenticateByToken)
	handler.Handle(ARpcNameAuthenticationIAuthenticationServiceHasUser, svr.HasUser)
	handler.Handle(ARpcNameAuthenticationIAuthenticationServiceUserById, svr.UserById)
	handler.Handle(ARpcNameAuthenticationIAuthenticationServiceUserByName, svr.UserByName)
	handler.Handle(ARpcNameAuthenticationIAuthenticationServiceUserByEmail, svr.UserByEmail)
}

func NewARpcServerIAuthenticationService(srv authentication.IAuthenticationService) *ARpcServerIAuthenticationService {
	return &ARpcServerIAuthenticationService{
		srv: srv,
	}
}
func (r *ARpcServerIAuthenticationService) AddUser(c *arpc.Context) {
	var req IAuthenticationServiceAddUserArgs
	var resp IAuthenticationServiceAddUserResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	a0, a1 := r.srv.AddUser(
		req.Val0,
		req.Val1,
	)
	resp.Val0 = a0
	resp.Val1 = a1
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerIAuthenticationService) DeleteUser(c *arpc.Context) {
	var req IAuthenticationServiceDeleteUserArgs
	var resp IAuthenticationServiceDeleteUserResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	a0 := r.srv.DeleteUser(
		req.Val0,
	)
	resp.Val0 = a0
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerIAuthenticationService) UpdateUser(c *arpc.Context) {
	var req IAuthenticationServiceUpdateUserArgs
	var resp IAuthenticationServiceUpdateUserResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	a0 := r.srv.UpdateUser(
		req.Val0,
	)
	resp.Val0 = a0
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerIAuthenticationService) Authenticate(c *arpc.Context) {
	var req IAuthenticationServiceAuthenticateArgs
	var resp IAuthenticationServiceAuthenticateResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	a0, a1 := r.srv.Authenticate(
		req.Val0,
		req.Val1,
	)
	resp.Val0 = a0
	resp.Val1 = a1
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerIAuthenticationService) AuthenticateByToken(c *arpc.Context) {
	var req IAuthenticationServiceAuthenticateByTokenArgs
	var resp IAuthenticationServiceAuthenticateByTokenResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	a0, a1 := r.srv.AuthenticateByToken(
		req.Val0,
	)
	resp.Val0 = a0
	resp.Val1 = a1
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerIAuthenticationService) HasUser(c *arpc.Context) {
	var req IAuthenticationServiceHasUserArgs
	var resp IAuthenticationServiceHasUserResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	a0, a1 := r.srv.HasUser(
		req.Val0,
	)
	resp.Val0 = a0
	resp.Val1 = a1
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerIAuthenticationService) UserById(c *arpc.Context) {
	var req IAuthenticationServiceUserByIdArgs
	var resp IAuthenticationServiceUserByIdResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	a0, a1 := r.srv.UserById(
		req.Val0,
	)
	resp.Val0 = a0
	resp.Val1 = a1
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerIAuthenticationService) UserByName(c *arpc.Context) {
	var req IAuthenticationServiceUserByNameArgs
	var resp IAuthenticationServiceUserByNameResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	a0, a1 := r.srv.UserByName(
		req.Val0,
	)
	resp.Val0 = a0
	resp.Val1 = a1
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerIAuthenticationService) UserByEmail(c *arpc.Context) {
	var req IAuthenticationServiceUserByEmailArgs
	var resp IAuthenticationServiceUserByEmailResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	a0, a1 := r.srv.UserByEmail(
		req.Val0,
	)
	resp.Val0 = a0
	resp.Val1 = a1
	_ = c.Write(&resp)
	return
}

// Scene App Definition

type ARpcAppIAuthenticationService struct {
	srv authentication.IAuthenticationService `aperture:""`
}

func (r *ARpcAppIAuthenticationService) Name() scene.ImplName {
	return authentication.Lens.ImplNameNoVer("ARpcApplication")
}

func (r *ARpcAppIAuthenticationService) RegisterService(server *arpc.Server) error {
	HandleIAuthenticationService(r.srv, server.Handler)
	return nil
}
