package gin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	ginbinding "github.com/gin-gonic/gin/binding"
	ginjson "github.com/gin-gonic/gin/codec/json"
)

// Binding binds one request source into target.
//
// The explicit BindURI, BindQuery, BindJSON, and form bindings defer
// validation. Handle validates the action once after all declared bindings
// have run, so multiple request sources can populate the same action first.
type Binding func(ctx *gin.Context, target any) error

// BindingProvider supplies request bindings in execution order.
type BindingProvider interface {
	Bindings() []Binding
}

// BindAuto uses Gin's content-type and method based binding selection.
// It is intended for an action that uses one automatically selected source.
func BindAuto(ctx *gin.Context, target any) error {
	if ctx.Request == nil {
		return errors.New("invalid request")
	}
	selected := ginbinding.Default(ctx.Request.Method, ctx.ContentType())
	switch selected.Name() {
	case ginbinding.JSON.Name():
		return BindJSON(ctx, target)
	case ginbinding.Form.Name():
		return BindForm(ctx, target)
	case ginbinding.FormPost.Name():
		return BindFormURLEncoded(ctx, target)
	default:
		return selected.Bind(ctx.Request, target)
	}
}

// BindJSON decodes a JSON request body without validating target.
func BindJSON(ctx *gin.Context, target any) error {
	if ctx.Request == nil || ctx.Request.Body == nil {
		return errors.New("invalid request")
	}
	decoder := ginjson.API.NewDecoder(ctx.Request.Body)
	if ginbinding.EnableDecoderUseNumber {
		decoder.UseNumber()
	}
	if ginbinding.EnableDecoderDisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	return decoder.Decode(target)
}

// BindForm binds query and form values without validating target.
func BindForm(ctx *gin.Context, target any) error {
	if ctx.Request == nil {
		return errors.New("invalid request")
	}
	if err := ctx.Request.ParseForm(); err != nil {
		return err
	}
	if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil &&
		!errors.Is(err, http.ErrNotMultipart) {
		return err
	}
	return ginbinding.MapFormWithTag(target, ctx.Request.Form, "form")
}

// BindFormURLEncoded binds URL-encoded form values without validating target.
func BindFormURLEncoded(ctx *gin.Context, target any) error {
	if ctx.Request == nil {
		return errors.New("invalid request")
	}
	if err := ctx.Request.ParseForm(); err != nil {
		return err
	}
	return ginbinding.MapFormWithTag(target, ctx.Request.PostForm, "form")
}

// BindQuery binds URL query values without validating target.
func BindQuery(ctx *gin.Context, target any) error {
	if ctx.Request == nil {
		return errors.New("invalid request")
	}
	return ginbinding.MapFormWithTag(target, ctx.Request.URL.Query(), "form")
}

// BindURI binds Gin route parameters without validating target.
func BindURI(ctx *gin.Context, target any) error {
	params := make(map[string][]string, len(ctx.Params))
	for _, param := range ctx.Params {
		params[param.Key] = []string{param.Value}
	}
	return ginbinding.MapFormWithTag(target, params, "uri")
}

// RequestAuto selects a binding from the request method and content type.
type RequestAuto struct{}

func (*RequestAuto) Bindings() []Binding {
	return []Binding{BindAuto}
}

// RequestJson binds a JSON request body.
type RequestJson struct{}

func (*RequestJson) Bindings() []Binding {
	return []Binding{BindJSON}
}

// RequestForm binds query and form values.
type RequestForm struct{}

func (*RequestForm) Bindings() []Binding {
	return []Binding{BindForm}
}

// RequestFormUrlEncoded binds a URL-encoded form body.
type RequestFormUrlEncoded struct{}

func (*RequestFormUrlEncoded) Bindings() []Binding {
	return []Binding{BindFormURLEncoded}
}

// RequestQuery binds URL query values.
type RequestQuery struct{}

func (*RequestQuery) Bindings() []Binding {
	return []Binding{BindQuery}
}

// RequestURI binds Gin route parameters.
type RequestURI struct{}

func (*RequestURI) Bindings() []Binding {
	return []Binding{BindURI}
}
