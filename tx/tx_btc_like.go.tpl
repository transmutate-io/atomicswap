package {{ .Values.package }}

import "github.com/transmutate-io/atomicswap/cryptos"

type tx{{ .Values.short }} struct{ *txBTC }

// New{{ .Values.short }} creates a new *tx{{ .Values.short }}
func New{{ .Values.short }}() (Tx, error) {
	b, _ := NewBTC()
	return &tx{{ .Values.short }}{txBTC: b.(*txBTC)}, nil
}

func (tx *tx{{ .Values.short }}) Crypto() *cryptos.Crypto { return cryptos.Cryptos["{{ .Values.name }}"] }

func (tx *tx{{ .Values.short }}) Copy() Tx { return &tx{{ .Values.short }}{txBTC: tx.txBTC.Copy().(*txBTC)} }

func (tx *tx{{ .Values.short }}) MarshalYAML() (interface{}, error) { return tx.txBTC, nil }

func (tx *tx{{ .Values.short }}) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(tx.txBTC)
}
