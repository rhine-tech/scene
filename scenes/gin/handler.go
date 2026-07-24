package gin

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	ginbinding "github.com/gin-gonic/gin/binding"
	"github.com/rhine-tech/scene/errcode"
	"github.com/rhine-tech/scene/model"
)

// Handle adapts an Action to a Gin handler.
//
// action is a pointer prototype. Handle creates a fresh zero-valued action for
// every request so bound request data is never shared between requests.
func Handle[T any](app T, action Action[T]) gin.HandlerFunc {
	actionType := reflect.TypeOf(action)
	if actionType == nil || actionType.Kind() != reflect.Ptr ||
		actionType.Elem().Kind() != reflect.Struct {
		panic("scene-gin: action should be a pointer to a struct")
	}

	// for choice of status code, please refer to
	// https://www.aynakeya.com/articles/coding/my-approach-for-using-status-code-in-restful-api/
	return func(nativeCtx *gin.Context) {
		ctx := &Context[T]{nativeCtx, app}
		// Create a fresh zero-valued action for each request so fields populated
		// by bindings are never shared between concurrent requests.
		r := reflect.New(actionType.Elem()).Interface().(Action[T])

		if provider, ok := r.(BindingProvider); ok {
			for _, bind := range provider.Bindings() {
				if err := bind(nativeCtx, r); err != nil {
					renderParameterError(ctx, err)
					return
				}
			}
			if ginbinding.Validator != nil {
				if err := ginbinding.Validator.ValidateStruct(r); err != nil {
					renderParameterError(ctx, err)
					return
				}
			}
		}

		resp, err := r.Process(ctx)
		if err != nil {
			if errors.Is(err, ErrAlreadyDone) {
				return
			}
			var ec *errcode.Error
			if !errors.As(err, &ec) {
				ec = errcode.UnknownError.WithDetail(err)
			}
			_ = ctx.Error(ec)
			ctx.JSON(http.StatusOK, model.NewErrorCodeResponse(ec))
			return
		}
		ctx.JSON(http.StatusOK, model.NewDataResponse(resp))
	}
}

func renderParameterError[T any](ctx *Context[T], err error) {
	ec := errcode.ParameterError.WithDetail(err)
	_ = ctx.Error(ec)
	ctx.JSON(http.StatusBadRequest, model.NewErrorCodeResponse(ec))
}
