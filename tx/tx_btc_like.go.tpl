package {{ .Values.package }}

import "github.com/transmutate-io/atomicswap/cryptos"

type tx{{ .Values.short }} struct{ *txBTC }

// New{{ .Values.short }} creates a new transaction for {{ .Values.name }}
func New{{ .Values.short }}() (Tx, error) {
	b, _ := NewBTC()
	return &tx{{ .Values.short }}{txBTC: b.(*txBTC)}, nil
}

// Crypto implement Tx
func (tx *tx{{ .Values.short }}) Crypto() *cryptos.Crypto { return cryptos.Cryptos["{{ .Values.name }}"] }

// Copy implement Tx
func (tx *tx{{ .Values.short }}) Copy() Tx { return &tx{{ .Values.short }}{txBTC: tx.txBTC.Copy().(*txBTC)} }

// MarshalYAML implement yaml.Marshaler
func (tx *tx{{ .Values.short }}) MarshalYAML() (interface{}, error) { return tx.txBTC, nil }

// UnmarshalYAML implement yaml.Unmarshaler
func (tx *tx{{ .Values.short }}) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(tx.txBTC)
}
