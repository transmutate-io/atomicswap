package tx

import "transmutate.io/pkg/atomicswap/cryptos"

type txDOGE struct{ *txBTC }

// NewDOGE creates a new *txDOGE
func NewDOGE() (Tx, error) {
	b, _ := NewBTC()
	return &txDOGE{txBTC: b.(*txBTC)}, nil
}

func (tx *txDOGE) Crypto() *cryptos.Crypto { return cryptos.Cryptos["dogecoin"] }

func (tx *txDOGE) Copy() Tx { return &txDOGE{txBTC: tx.txBTC.Copy().(*txBTC)} }

func (tx *txDOGE) MarshalYAML() (interface{}, error) { return tx.txBTC, nil }

func (tx *txDOGE) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(tx.txBTC)
}
