package {{ .Values.package | default "main" }}

var ({{ $tn := .Values.type_name -}}
	{{- range $short, $data := .Values.cryptos }}
	{{ title ( dashed_to_camel $data.name ) }} = &{{ $tn }}{
		Name:     "{{ $data.name }}",
		Short:    "{{ $short }}",
		Decimals: {{ $data.decimals }},
		Type:     {{ $data.type }},
	}
	{{- end }}

	{{ .Values.type_name }}s = map[string]*{{ .Values.type_name }}{
	{{- range $short, $data := .Values.cryptos }}
		"{{ $data.name }}": {{ title ( dashed_to_camel $data.name ) }},
	{{- end }}
	}
)
