package tx

import "github.com/transmutate-io/atomicswap/cryptos"

type txLTC struct{ *txBTC }

// NewLTC creates a new transaction for litecoin
func NewLTC() (Tx, error) {
	b, _ := NewBTC()
	return &txLTC{txBTC: b.(*txBTC)}, nil
}

// Crypto implement Tx
func (tx *txLTC) Crypto() *cryptos.Crypto { return cryptos.Cryptos["litecoin"] }

// Copy implement Tx
func (tx *txLTC) Copy() Tx { return &txLTC{txBTC: tx.txBTC.Copy().(*txBTC)} }

// MarshalYAML implement yaml.Marshaler
func (tx *txLTC) MarshalYAML() (interface{}, error) { return tx.txBTC, nil }

// UnmarshalYAML implement yaml.Unmarshaler
func (tx *txLTC) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(tx.txBTC)
}
