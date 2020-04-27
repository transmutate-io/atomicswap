package {{ .Values.package }}

import "transmutate.io/pkg/atomicswap/cryptos"

type tx{{ .Values.short }} struct{ *txBTC }

// NewTx{{ .Values.short }} creates a new *tx{{ .Values.short }}
func NewTx{{ .Values.short }}() (Tx, error) {
	b, _ := NewTxBTC()
	return &tx{{ .Values.short }}{txBTC: b.(*txBTC)}, nil
}

func (tx *tx{{ .Values.short }}) Crypto() *cryptos.Crypto { return cryptos.Cryptos["{{ .Values.name }}"] }

func (tx *tx{{ .Values.short }}) Copy() Tx { return &tx{{ .Values.short }}{txBTC: tx.txBTC.Copy().(*txBTC)} }

func (tx *tx{{ .Values.short }}) MarshalYAML() (interface{}, error) { return tx.txBTC, nil }

func (tx *tx{{ .Values.short }}) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(tx.txBTC)
}
