package {{ .All.Package | default "main" }}

import "transmutate.io/pkg/atomicswap/cryptotypes"

var (
	{{ .All.InterfaceType }}s = map[string]*{{ .All.InterfaceType }}{
	{{- range $short, $data := .All.MainTemplate.Values }}
		"{{ $data.name }}": &Crypto{
			Name:     "{{ $data.name }}",
			Short:    "{{ $short }}",
			Decimals: {{ $data.decimals }},
			Type:     cryptotypes.{{ $data.type }},
		},
	{{- end }}
	}
)
