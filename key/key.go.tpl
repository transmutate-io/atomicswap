package {{ .Values.package }}

var cryptoFuncs = map[string]newFuncs{
	{{- range $i, $v := .Values.cryptos }}
	"{{ $v.name }}": newFuncs{
		parsePriv: ParsePrivate{{ $i }},
		parsePub:       ParsePublic{{ $i }},
		newPriv:      NewPrivate{{ $i }},
	},
	{{- end }}
}
