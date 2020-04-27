package {{ .Values.package }}

var txFuncs = map[string]NewTxFunc{
    {{- range .Values.coins }}
	"{{ .name }}": NewTx{{ .short }},
    {{- end }}
}
