package {{ .Values.package }}

var testCryptos = []string{
	{{- range $i, $v := .Values.cryptos }}
	"{{ $v.name }}",
	{{- end }}
}