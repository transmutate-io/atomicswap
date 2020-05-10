package tx

import "transmutate.io/pkg/atomicswap/cryptos"

type txLTC struct{ *txBTC }

// NewLTC creates a new *txLTC
func NewLTC() (Tx, error) {
	b, _ := NewBTC()
	return &txLTC{txBTC: b.(*txBTC)}, nil
}

func (tx *txLTC) Crypto() *cryptos.Crypto { return cryptos.Cryptos["litecoin"] }

func (tx *txLTC) Copy() Tx { return &txLTC{txBTC: tx.txBTC.Copy().(*txBTC)} }

func (tx *txLTC) MarshalYAML() (interface{}, error) { return tx.txBTC, nil }

func (tx *txLTC) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(tx.txBTC)
}
