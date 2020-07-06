package {{ .Values.package }}

import "github.com/transmutate-io/cryptocore/types"

var (
	newFundsDataFuncs = map[string ]func()FundsData{
		{{- range $short, $data := .Values.cryptos }}
		"{{ $data.name }}": newFundsData{{ $short }},
		{{- end }}
	}
	newFundsLockFuncs = map[string]func(types.Bytes)Lock{
		{{- range $short, $data := .Values.cryptos }}
		"{{ $data.name }}": newFundsLock{{ $short }},
		{{- end }}
	}
)