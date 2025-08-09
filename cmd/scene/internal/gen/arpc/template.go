package arpc

import "github.com/rhine-tech/scene/registry"

const apertureTag = "`" + registry.InjectTag + `:""` + "`"

const serviceTemplate = `
import (
	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	sarpc "github.com/rhine-tech/scene/scenes/arpc"
	"time"
{{- range $path := .RequiredPackage }}
	"{{ $path }}"
{{- end }}
)

// Method definition

const (
{{- range .Methods }}
	ARpcName{{UpperFirst $.PackageName}}{{ $.InterfaceName }}{{ .Name }} = "{{$.PackageName}}.{{ $.InterfaceName }}.{{ .Name }}"
{{- end }}
)
{{""}}
{{- range .Methods }}
type {{ $.InterfaceName }}{{ .Name }}Args struct {
	{{- range $index, $ret := .Args }}
	Val{{ $index }} {{ $ret.Type }}
	{{- end }}
}

type {{ $.InterfaceName }}{{ .Name }}Result struct {
	{{- range $index, $ret := .Returns }}
	Val{{ $index }} {{ $ret.Type }}
	{{- end }}
}
{{- end }}

// Service (Client) Implementation

type arpcClient{{ .InterfaceName }} struct {
	client sarpc.Client  ` + apertureTag + `
	timeout time.Duration
	log 	logger.ILogger ` + apertureTag + `
}


func NewARpc{{ .InterfaceName }}(client sarpc.Client) {{ $.PackageName }}.{{ .InterfaceName }} {
	return &arpcClient{{ .InterfaceName }}{
		client:  client,
		timeout: time.Second * 5,
	}
}

func NewARpc{{ .InterfaceName }}WithTimeout(client sarpc.Client, timeout time.Duration) {{ $.PackageName }}.{{ .InterfaceName }} {
	return &arpcClient{{ .InterfaceName }}{
		client:  client,
		timeout: timeout,
	}
}

func (r *arpcClient{{ .InterfaceName }}) SrvImplName() scene.ImplName {
	return {{ $.PackageName }}.Lens.ImplName("{{ .InterfaceName }}", "arpc")
}

// Deprecated: no longer used
func (r *arpcClient{{ .InterfaceName }}) WithSceneContext(ctx scene.Context) {{ $.PackageName }}.{{ .InterfaceName }} {
	return r
}
{{ "" }}

{{- range .Methods }}
func (r *arpcClient{{ $.InterfaceName }}) {{ .Name }}({{ range .Args }}{{ .Name }} {{ .Type }}, {{ end }}) ({{ range $index, $ret := .Returns }}{{ if $index }}, {{ end }}{{ $ret.Type }}{{ end }}) {
	var resp {{ $.InterfaceName }}{{ .Name }}Result
	err := r.client.Call(ARpcName{{UpperFirst $.PackageName}}{{ $.InterfaceName }}{{ .Name }}, &{{ $.InterfaceName }}{{ .Name }}Args{
		{{- range $index, $arg := .Args }}
		Val{{ $index }}: {{ $arg.Name }},
		{{- end }}
	}, &resp,r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcName{{UpperFirst $.PackageName}}{{ $.InterfaceName }}{{ .Name }}, "err", err)
		return   {{- range $index, $ret := .Returns }}{{ if $index }},{{ end }} {{ if eq $ret.Type "error" }}err{{ else }}*new({{ $ret.Type }}){{ end }}{{- end }}
	}
	return    {{- range $index, $ret := .Returns }}{{ if $index }},{{ end }} resp.Val{{ $index }}{{- end }}
}
{{- end }}

// Server Implementation

type ARpcServer{{ .InterfaceName }} struct {
	srv {{ $.PackageName }}.{{ .InterfaceName }} ` + apertureTag + `
}

func Handle{{ .InterfaceName }}(srv {{ $.PackageName }}.{{ .InterfaceName }}, handler arpc.Handler) {
	svr := NewARpcServer{{ .InterfaceName }}(srv)
	HandleARpcServer{{ .InterfaceName }}(svr,handler)
}

func HandleARpcServer{{ .InterfaceName }}(svr *ARpcServer{{ .InterfaceName }}, handler arpc.Handler) {
{{- range .Methods }}
	handler.Handle(ARpcName{{UpperFirst $.PackageName}}{{ $.InterfaceName }}{{ .Name }} , svr.{{ .Name }})
{{- end }}
} 

func NewARpcServer{{ .InterfaceName }}(srv {{ $.PackageName }}.{{ .InterfaceName }}) *ARpcServer{{ .InterfaceName }} {
	return &ARpcServer{{ .InterfaceName }}{
		srv: srv,
	}
}

{{- range .Methods }}
func (r *ARpcServer{{ $.InterfaceName }}) {{ .Name }}(c *arpc.Context) {
	var req {{ $.InterfaceName }}{{ .Name }}Args
	var resp {{ $.InterfaceName }}{{ .Name }}Result
	err := c.Bind(&req)
	if err != nil {
		return
	}
	{{range $index, $arg := .Returns }}{{ if $index }},{{ end }} a{{ $index }}{{end }} := r.srv.{{ .Name }}(
		{{- range $index, $ret := .Args }}
		req.Val{{ $index }},
		{{- end }}
	)
	{{- range $index, $ret := .Returns }}
	resp.Val{{ $index }} = a{{ $index }}
	{{- end }}
	_ = c.Write(&resp)
	return
}
{{- end }}

// Scene App Definition

type ARpcApp{{ .InterfaceName }} struct {
	srv {{ $.PackageName }}.{{ .InterfaceName }} ` + apertureTag + `
}

func (r *ARpcApp{{ .InterfaceName }}) Name() scene.ImplName {
	return {{ $.PackageName }}.Lens.ImplNameNoVer("ARpcApplication")
}

func (r *ARpcApp{{ .InterfaceName }}) RegisterService(server *arpc.Server) error {
	Handle{{ .InterfaceName }}(r.srv, server.Handler)
	return nil
}

`
