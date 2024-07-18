package rpc

const serviceTemplate = `
import (
	"github.com/rhine-tech/scene"
	srpc "github.com/rhine-tech/scene/scenes/rpc"
{{- range $path := .RequiredPackage }}
	"{{ $path }}"
{{- end }}
)

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

type rpcClient{{ .InterfaceName }} struct {
	client srpc.Client
}

func NewRpc{{ .InterfaceName }}(client srpc.Client) {{ $.PackageName }}.{{ .InterfaceName }} {
	return &rpcClient{{ .InterfaceName }}{
		client: client,
	}
}

func (r *rpcClient{{ .InterfaceName }}) SrvImplName() scene.ImplName {
	return {{ $.PackageName }}.Lens.ImplName("{{ .InterfaceName }}", "rpc")
}
{{ "" }}

{{- range .Methods }}
func (r *rpcClient{{ $.InterfaceName }}) {{ .Name }}({{ range .Args }}{{ .Name }} {{ .Type }}, {{ end }}) ({{ range $index, $ret := .Returns }}{{ if $index }}, {{ end }}{{ $ret.Type }}{{ end }}) {
	var resp {{ $.InterfaceName }}{{ .Name }}Result
	err := r.client.Call("{{UpperFirst $.PackageName}}.{{ $.InterfaceName }}.{{ .Name }}", &{{ $.InterfaceName }}{{ .Name }}Args{
		{{- range $index, $arg := .Args }}
		Val{{ $index }}: {{ $arg.Name }},
		{{- end }}
	}, &resp)
	if err != nil {
		return   {{- range $index, $ret := .Returns }}{{ if $index }},{{ end }} *new({{ $ret.Type }}){{- end }}
	}
	return    {{- range $index, $ret := .Returns }}{{ if $index }},{{ end }} resp.Val{{ $index }}{{- end }}
}
{{- end }}

type RpcServer{{ .InterfaceName }} struct {
	srv {{ $.PackageName }}.{{ .InterfaceName }}
}

func NewRpcServer{{ .InterfaceName }}(srv {{ $.PackageName }}.{{ .InterfaceName }}) *RpcServer{{ .InterfaceName }} {
	return &RpcServer{{ .InterfaceName }}{
		srv: srv,
	}
}

{{- range .Methods }}
func (r *RpcServer{{ $.InterfaceName }}) {{ .Name }}(req *{{ $.InterfaceName }}{{ .Name }}Args, resp *{{ $.InterfaceName }}{{ .Name }}Result) error {
	{{range $index, $arg := .Returns }}{{ if $index }},{{ end }} a{{ $index }}{{end }} := r.srv.{{ .Name }}(
		{{- range $index, $ret := .Args }}
		req.Val{{ $index }},
		{{- end }}
	)
	{{- range $index, $ret := .Returns }}
	resp.Val{{ $index }} = a{{ $index }}
	{{- end }}
	return nil
}
{{- end }}
`
