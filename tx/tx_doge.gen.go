package tx

import "github.com/transmutate-io/atomicswap/cryptos"

type txDOGE struct{ *txBTC }

// NewDOGE creates a new transaction for dogecoin
func NewDOGE() (Tx, error) {
	b, _ := NewBTC()
	return &txDOGE{txBTC: b.(*txBTC)}, nil
}

// Crypto implement Tx
func (tx *txDOGE) Crypto() *cryptos.Crypto { return cryptos.Cryptos["dogecoin"] }

// Copy implement Tx
func (tx *txDOGE) Copy() Tx { return &txDOGE{txBTC: tx.txBTC.Copy().(*txBTC)} }

// MarshalYAML implement yaml.Marshaler
func (tx *txDOGE) MarshalYAML() (interface{}, error) { return tx.txBTC, nil }

// UnmarshalYAML implement yaml.Unmarshaler
func (tx *txDOGE) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(tx.txBTC)
}
