package {{ .Values.package }}

var cryptoFuncs = map[string]newFuncs{
	{{- range $i, $v := .Values.cryptos }}
	"{{ $v.name }}": newFuncs{
		parsePriv: ParsePrivate{{ $i }},
		priv:      NewPrivate{{ $i }},
		pub:       NewPublic{{ $i }},
	},
	{{- end }}
}
