package {{ .Values.package }}

var newHasherFuncs = map[string]func() Hasher{
    {{- range $short,$d := .Values.cryptos }}
	"{{ $d.name }}": New{{ $short }},
    {{- end }}
}
