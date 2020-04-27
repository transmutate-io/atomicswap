package {{ .Values.package | default "main" }}

var (
	{{ $tn := .Values.type_name -}}
	{{ .Values.type_name }}s = map[string]*{{ .Values.type_name }}{
	{{- range $short, $data := .Values.coins }}
		"{{ $data.name }}": &{{ $tn }}{
			Name:     "{{ $data.name }}",
			Short:    "{{ $short }}",
			Decimals: {{ $data.decimals }},
			Type:     {{ $data.type }},
		},
	{{- end }}
	}
)
