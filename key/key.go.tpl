package {{ .Values.package }}

var cryptoFuncs = map[string]newFuncs{
	{{- range $i, $v := .Values.coins }}
	"{{ $v.name }}": newFuncs{
		parsePriv: ParsePrivate{{ $v.short }},
		priv:      NewPrivate{{ $v.short }},
		pub:       NewPublic{{ $v.short }},
	},
	{{- end }}
}
