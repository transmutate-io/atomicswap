package {{ .Values.package }}

type hasher{{ .Values.short }} struct{ hasherBTC }

func New{{ .Values.short }}() Hasher { return hasher{{ .Values.short }}{} }
