package gin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene/errcode"
	"github.com/rhine-tech/scene/model"
	"github.com/stretchr/testify/require"
)

type multiBindingTestApp struct {
	id     string
	filter string
	name   string
}

type multiBindingTestAction struct {
	ID     string `uri:"id" binding:"required"`
	Filter string `form:"filter" binding:"required"`
	Name   string `json:"name" binding:"required"`
}

func (a *multiBindingTestAction) GetRoute() HttpRouteInfo {
	return HttpRouteInfo{Methods: HttpMethodPut, Path: "/items/:id"}
}

func (a *multiBindingTestAction) Bindings() []Binding {
	return []Binding{BindURI, BindQuery, BindJSON}
}

func (a *multiBindingTestAction) Process(ctx *Context[*multiBindingTestApp]) (any, error) {
	ctx.App.id = a.ID
	ctx.App.filter = a.Filter
	ctx.App.name = a.Name
	return nil, nil
}

type wrappedErrorTestAction struct{}

func (a *wrappedErrorTestAction) GetRoute() HttpRouteInfo {
	return HttpRouteInfo{Methods: HttpMethodGet, Path: "/wrapped-error"}
}

func (a *wrappedErrorTestAction) Process(*Context[struct{}]) (any, error) {
	return nil, fmt.Errorf(
		"service failed: %w",
		errcode.ParameterError.WithDetailStr("invalid value"),
	)
}

func TestHandlePreservesWrappedErrorCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/wrapped-error", Handle(struct{}{}, new(wrappedErrorTestAction)))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/wrapped-error", nil)
	engine.ServeHTTP(recorder, request)

	var response model.AppResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, errcode.ParameterError.Code, response.Code)
	require.Contains(t, response.Msg, "invalid value")
}

func TestHandleAppliesMultipleBindingsBeforeValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	app := &multiBindingTestApp{}
	engine.PUT("/items/:id", Handle(app, new(multiBindingTestAction)))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPut,
		"/items/item-1?filter=active",
		bytes.NewBufferString(`{"name":"scene"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "item-1", app.id)
	require.Equal(t, "active", app.filter)
	require.Equal(t, "scene", app.name)
}
