package {{ .All.Package | default "main" }}

var (
	{{ .All.InterfaceType }}s = map[string]*{{ .All.InterfaceType }}{
	{{- range $short, $data := .All.MainTemplate.Values }}
		"{{ $data.name }}": &Crypto{
			Name:     "{{ $data.name }}",
			Short:    "{{ $short }}",
			Decimals: {{ $data.decimals }},
			Type:     {{ $data.type }},
		},
	{{- end }}
	}
)
