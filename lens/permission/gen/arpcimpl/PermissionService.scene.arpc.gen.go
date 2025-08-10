package arpcimpl
import (
	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	sarpc "github.com/rhine-tech/scene/scenes/arpc"
	"time"
	"github.com/rhine-tech/scene/lens/permission"
)

// Method definition

const (
	ARpcNamePermissionPermissionServiceHasPermission = "permission.PermissionService.HasPermission"
	ARpcNamePermissionPermissionServiceHasPermissionStr = "permission.PermissionService.HasPermissionStr"
	ARpcNamePermissionPermissionServiceListPermissions = "permission.PermissionService.ListPermissions"
	ARpcNamePermissionPermissionServiceAddPermission = "permission.PermissionService.AddPermission"
	ARpcNamePermissionPermissionServiceRemovePermission = "permission.PermissionService.RemovePermission"
)

type PermissionServiceHasPermissionArgs struct {
	Val0 string
	Val1 *permission.Permission
}

type PermissionServiceHasPermissionResult struct {
	Val0 bool
}
type PermissionServiceHasPermissionStrArgs struct {
	Val0 string
	Val1 string
}

type PermissionServiceHasPermissionStrResult struct {
	Val0 bool
}
type PermissionServiceListPermissionsArgs struct {
	Val0 string
}

type PermissionServiceListPermissionsResult struct {
	Val0 []*permission.Permission
}
type PermissionServiceAddPermissionArgs struct {
	Val0 string
	Val1 string
}

type PermissionServiceAddPermissionResult struct {
	Val0 error
}
type PermissionServiceRemovePermissionArgs struct {
	Val0 string
	Val1 string
}

type PermissionServiceRemovePermissionResult struct {
	Val0 error
}

// Service (Client) Implementation

type arpcClientPermissionService struct {
	client sarpc.Client  `aperture:""`
	timeout time.Duration
	log 	logger.ILogger `aperture:""`
}


func NewARpcPermissionService(client sarpc.Client) permission.PermissionService {
	return &arpcClientPermissionService{
		client:  client,
		timeout: time.Second * 5,
	}
}

func NewARpcPermissionServiceWithTimeout(client sarpc.Client, timeout time.Duration) permission.PermissionService {
	return &arpcClientPermissionService{
		client:  client,
		timeout: timeout,
	}
}

func (r *arpcClientPermissionService) SrvImplName() scene.ImplName {
	return permission.Lens.ImplName("PermissionService", "arpc")
}

// Deprecated: no longer used
func (r *arpcClientPermissionService) WithSceneContext(ctx scene.Context) permission.PermissionService {
	return r
}

func (r *arpcClientPermissionService) HasPermission(owner string, perm *permission.Permission, ) (bool) {
	var resp PermissionServiceHasPermissionResult
	err := r.client.Call(ARpcNamePermissionPermissionServiceHasPermission, &PermissionServiceHasPermissionArgs{
		Val0: owner,
		Val1: perm,
	}, &resp,r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNamePermissionPermissionServiceHasPermission, "err", err)
		return *new(bool)
	}
	return resp.Val0
}
func (r *arpcClientPermissionService) HasPermissionStr(owner string, perm string, ) (bool) {
	var resp PermissionServiceHasPermissionStrResult
	err := r.client.Call(ARpcNamePermissionPermissionServiceHasPermissionStr, &PermissionServiceHasPermissionStrArgs{
		Val0: owner,
		Val1: perm,
	}, &resp,r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNamePermissionPermissionServiceHasPermissionStr, "err", err)
		return *new(bool)
	}
	return resp.Val0
}
func (r *arpcClientPermissionService) ListPermissions(owner string, ) ([]*permission.Permission) {
	var resp PermissionServiceListPermissionsResult
	err := r.client.Call(ARpcNamePermissionPermissionServiceListPermissions, &PermissionServiceListPermissionsArgs{
		Val0: owner,
	}, &resp,r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNamePermissionPermissionServiceListPermissions, "err", err)
		return *new([]*permission.Permission)
	}
	return resp.Val0
}
func (r *arpcClientPermissionService) AddPermission(owner string, perm string, ) (error) {
	var resp PermissionServiceAddPermissionResult
	err := r.client.Call(ARpcNamePermissionPermissionServiceAddPermission, &PermissionServiceAddPermissionArgs{
		Val0: owner,
		Val1: perm,
	}, &resp,r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNamePermissionPermissionServiceAddPermission, "err", err)
		return err
	}
	return resp.Val0
}
func (r *arpcClientPermissionService) RemovePermission(owner string, perm string, ) (error) {
	var resp PermissionServiceRemovePermissionResult
	err := r.client.Call(ARpcNamePermissionPermissionServiceRemovePermission, &PermissionServiceRemovePermissionArgs{
		Val0: owner,
		Val1: perm,
	}, &resp,r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNamePermissionPermissionServiceRemovePermission, "err", err)
		return err
	}
	return resp.Val0
}

// Server Implementation

type ARpcServerPermissionService struct {
	srv permission.PermissionService `aperture:""`
}

func HandlePermissionService(srv permission.PermissionService, handler arpc.Handler) {
	svr := NewARpcServerPermissionService(srv)
	HandleARpcServerPermissionService(svr,handler)
}

func HandleARpcServerPermissionService(svr *ARpcServerPermissionService, handler arpc.Handler) {
	handler.Handle(ARpcNamePermissionPermissionServiceHasPermission , svr.HasPermission)
	handler.Handle(ARpcNamePermissionPermissionServiceHasPermissionStr , svr.HasPermissionStr)
	handler.Handle(ARpcNamePermissionPermissionServiceListPermissions , svr.ListPermissions)
	handler.Handle(ARpcNamePermissionPermissionServiceAddPermission , svr.AddPermission)
	handler.Handle(ARpcNamePermissionPermissionServiceRemovePermission , svr.RemovePermission)
} 

func NewARpcServerPermissionService(srv permission.PermissionService) *ARpcServerPermissionService {
	return &ARpcServerPermissionService{
		srv: srv,
	}
}
func (r *ARpcServerPermissionService) HasPermission(c *arpc.Context) {
	var req PermissionServiceHasPermissionArgs
	var resp PermissionServiceHasPermissionResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	 a0 := r.srv.HasPermission(
		req.Val0,
		req.Val1,
	)
	resp.Val0 = a0
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerPermissionService) HasPermissionStr(c *arpc.Context) {
	var req PermissionServiceHasPermissionStrArgs
	var resp PermissionServiceHasPermissionStrResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	 a0 := r.srv.HasPermissionStr(
		req.Val0,
		req.Val1,
	)
	resp.Val0 = a0
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerPermissionService) ListPermissions(c *arpc.Context) {
	var req PermissionServiceListPermissionsArgs
	var resp PermissionServiceListPermissionsResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	 a0 := r.srv.ListPermissions(
		req.Val0,
	)
	resp.Val0 = a0
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerPermissionService) AddPermission(c *arpc.Context) {
	var req PermissionServiceAddPermissionArgs
	var resp PermissionServiceAddPermissionResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	 a0 := r.srv.AddPermission(
		req.Val0,
		req.Val1,
	)
	resp.Val0 = a0
	_ = c.Write(&resp)
	return
}
func (r *ARpcServerPermissionService) RemovePermission(c *arpc.Context) {
	var req PermissionServiceRemovePermissionArgs
	var resp PermissionServiceRemovePermissionResult
	err := c.Bind(&req)
	if err != nil {
		return
	}
	 a0 := r.srv.RemovePermission(
		req.Val0,
		req.Val1,
	)
	resp.Val0 = a0
	_ = c.Write(&resp)
	return
}

// Scene App Definition

type ARpcAppPermissionService struct {
	srv permission.PermissionService `aperture:""`
}

func (r *ARpcAppPermissionService) Name() scene.ImplName {
	return permission.Lens.ImplNameNoVer("ARpcApplication")
}

func (r *ARpcAppPermissionService) RegisterService(server *arpc.Server) error {
	HandlePermissionService(r.srv, server.Handler)
	return nil
}

