package {{ .Values.package }}

var (
	generators = map[string]Generator{
		{{- range $short, $data := .Values.cryptos }}
			"{{ $data.name }}": NewGenerator{{ $short }}(),
		{{- end }}
	}

	disassemblers = map[string]Disassembler{
		{{- range $short, $data := .Values.cryptos }}
			"{{ $data.name }}": NewDisassembler{{ $short }}(),
		{{- end }}
	}

	intParsers = map[string]IntParser{
		{{- range $short, $data := .Values.cryptos }}
			"{{ $data.name }}": NewIntParser{{ $short }}(),
		{{- end }}
	}
)
