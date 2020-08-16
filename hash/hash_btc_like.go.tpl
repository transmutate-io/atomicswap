package {{ .Values.package }}

type hasher{{ .Values.short }} struct{ hasherBTC }

// New{{ .Values.short }} returns an hasher for {{ .Values.name }}
func New{{ .Values.short }}() Hasher { return hasher{{ .Values.short }}{} }
