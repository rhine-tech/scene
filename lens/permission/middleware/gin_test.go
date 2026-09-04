package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/model"
	"github.com/stretchr/testify/require"
)

type failingPermissionService struct {
	err error
}

func (f *failingPermissionService) ImplName() scene.ImplName {
	return permission.Lens.ImplName("PermissionService", "failing")
}

func (f *failingPermissionService) HasPermission(context.Context, string, *permission.Permission) (bool, error) {
	return false, f.err
}

func (f *failingPermissionService) HasPermissionStr(context.Context, string, string) (bool, error) {
	return false, f.err
}

func (f *failingPermissionService) ListPermissions(context.Context, string) ([]*permission.Permission, error) {
	return nil, f.err
}

func (f *failingPermissionService) AddPermission(context.Context, string, string) error {
	return f.err
}

func (f *failingPermissionService) RemovePermission(context.Context, string, string) error {
	return f.err
}

func (f *failingPermissionService) ReplacePermissions(context.Context, string, *permission.Permission, []*permission.Permission) error {
	return f.err
}

func (f *failingPermissionService) ListOwnersWithPermission(context.Context, *permission.Permission, int64, int64) (model.PaginationResult[permission.OwnerPermissions], error) {
	return model.PaginationResult[permission.OwnerPermissions]{}, f.err
}

func (f *failingPermissionService) ListOwnersWithGrantsByPrefix(context.Context, *permission.Permission, int64, int64) (model.PaginationResult[permission.OwnerPermissions], error) {
	return model.PaginationResult[permission.OwnerPermissions]{}, f.err
}

func (f *failingPermissionService) RemovePermissionsByPrefix(context.Context, *permission.Permission) error {
	return f.err
}

func TestGinRequirePermissionReturnsInternalErrorOnServiceFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &failingPermissionService{err: errors.New("permission repository unavailable")}
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		ctx := permission.SetPermContext(c.Request.Context(), "owner", service)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	engine.GET("/", GinRequirePermission(permission.MustParsePermission("project:read")), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	engine.ServeHTTP(response, request)

	require.Equal(t, http.StatusInternalServerError, response.Code)
}
