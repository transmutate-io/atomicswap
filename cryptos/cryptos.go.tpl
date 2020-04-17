package {{ .All.Package | default "main" }}

var (
	_{{ .All.TypePrefix }}s = map[string]new{{ .All.InterfaceType }}Func{
		{{ $cn := .All.InterfaceType }}
		{{- range .All.Templates -}}
		"{{ .Values.name }}": new{{ $cn }}{{ .Values.short | upper }},
		{{ end -}}
	}
)
