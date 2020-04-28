package {{ .Values.package }}

var txFuncs = map[string]NewTxFunc{
    {{- range $short,$d := .Values.cryptos }}
	"{{ $d.name }}": NewTx{{ $short }},
    {{- end }}
}
