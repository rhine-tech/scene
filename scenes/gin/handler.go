package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/rhine-tech/scene/errcode"
	"github.com/rhine-tech/scene/model"
	"net/http"
	"reflect"
)

type Request[T any] interface {
	Process(ctx *Context[T]) (data any, err error)
	Binding() binding.Binding
}

type ParameterBinder[T any] interface {
	Bind(ctx *Context[T]) error
}

type RequestNoParam struct{}

func (r *RequestNoParam) Binding() binding.Binding { return nil }

type RequestJson struct{}

func (r *RequestJson) Binding() binding.Binding {
	return binding.JSON
}

type RequestQuery struct{}

func (r *RequestQuery) Binding() binding.Binding { return binding.Query }

type uriBindingPlaceHolder struct{}

func (u uriBindingPlaceHolder) Name() string {
	return "uri_placeholder"
}

func (u uriBindingPlaceHolder) Bind(request *http.Request, obj any) error {
	return nil
}

type RequestURI struct{}

func (r *RequestURI) Binding() binding.Binding { return uriBindingPlaceHolder{} }

func RequestWrapper[T any](app T) func(Request[T]) gin.HandlerFunc {
	return func(request Request[T]) gin.HandlerFunc {
		return Handle(app, request)
	}
}

func WrapReq[A any, T Request[A]](app A, request T) gin.HandlerFunc {
	return Handle(app, request)
}

func Handle[A any, T Request[A]](app A, request T) gin.HandlerFunc {
	// T is a request object, should be a pointer to a struct, and implement Request interface
	// if not, Request object will not be able to receive bound parameters
	if reflect.TypeOf(request).Kind() != reflect.Ptr {
		panic("scene-gin: request object should be a pointer")
	}
	reqType := reflect.ValueOf(request).Elem().Type()

	// for choice of status code, please refer to
	// https://www.aynakeya.com/articles/coding/my-approach-for-using-status-code-in-restful-api/
	return func(nativeCtx *gin.Context) {
		ctx := &Context[A]{nativeCtx, app}
		// create a new request object
		var r Request[A] = reflect.New(reqType).Interface().(Request[A])
		// ======== here is the delivery layer, so https status code will be standard http code ========
		// bind parameter with request
		// process request, explicitly bind uri parameter
		if r.Binding() != nil {
			if r.Binding().Name() == "uri_placeholder" {
				if err := ctx.ShouldBindUri(r); err != nil {
					ec := errcode.ParameterError.WithDetail(err)
					_ = ctx.Error(ec)
					ctx.JSON(http.StatusBadRequest, model.NewErrorCodeResponse(ec))
					return
				}
			} else {
				// otherwise, bind request with query or json
				if err := ctx.ShouldBindWith(r, r.Binding()); err != nil {
					ec := errcode.ParameterError.WithDetail(err)
					_ = ctx.Error(ec)
					ctx.JSON(http.StatusBadRequest, model.NewErrorCodeResponse(ec))
					return
				}
			}
		}
		// bind extra parameter if exists
		extraBinder, ok := r.(ParameterBinder[A])
		if ok {
			if err := extraBinder.Bind(ctx); err != nil {
				ec := errcode.ParameterError.WithDetail(err)
				_ = ctx.Error(ec)
				ctx.JSON(http.StatusBadRequest, model.NewErrorCodeResponse(ec))
				return
			}
		}
		// === here starts the business logic layer ===
		resp, err := r.Process(ctx)
		if err != nil {
			ec, ok := err.(*errcode.Error)
			if !ok {
				ec = errcode.UnknownError.WithDetail(err)
			}
			_ = ctx.Error(ec)
			ctx.JSON(http.StatusOK, model.NewErrorCodeResponse(ec))
			return
		}
		ctx.JSON(http.StatusOK, model.NewDataResponse(resp))
	}
}
