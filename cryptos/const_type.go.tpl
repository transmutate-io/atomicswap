package {{ .Values.package | default "main" }}

import "fmt"

type Invalid{{ .Values.type_name }}Error string

func (e Invalid{{ .Values.type_name }}Error) Error() string {
	return fmt.Sprintf("invalid {{ .Values.type_desc | default (lower .Values.type_name) }}: \"%s\"", string(e))
}

type {{ .Values.type_name }} int

func Parse{{ .Values.type_name }}(s string) ({{ .Values.type_name }}, error) {
	var r {{ .Values.type_name }}
	if err := (&r).Set(s); err != nil {
		return 0, err
	}
	return r, nil
}

func (v {{ .Values.type_name }}) String() string { return _{{ .Values.type_name }}[v] }

func (v *{{ .Values.type_name }}) Set(sv string) error {
	nv, ok := _{{ .Values.type_name }}Names[sv]
	if !ok {
		return Invalid{{ .Values.type_name }}Error(sv)
	}
	*v = nv
	return nil
}

func (v {{ .Values.type_name }}) MarshalYAML() (interface{}, error) { return v.String(), nil }

func (v *{{ .Values.type_name }}) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	return v.Set(r)
}

const (
{{- $tn := .Values.type_name -}}
{{- range $n, $i := .Values.consts }}
	{{ $i.name }}{{ if eq $n 0 }} {{ $tn }} {{ end }}{{ if eq $n 0 }}= iota{{ end }}
{{- end }}
)

var (
	_{{ .Values.type_name }} = map[{{ .Values.type_name }}]string{
{{- range $n, $i := .Values.consts }}
		{{ $i.name }}: "{{ $i.value }}",
{{- end }}
	}
	_{{ .Values.type_name }}Names map[string]{{ .Values.type_name }}
)

func init() {
	_{{ .Values.type_name }}Names = make(map[string]{{ .Values.type_name }}, len(_{{ .Values.type_name }}))
	for k, v := range _{{ .Values.type_name }} {
		_{{ .Values.type_name }}Names[v] = k
	}
}

