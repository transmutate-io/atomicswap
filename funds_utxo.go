package atomicswap

import (
	"transmutate.io/pkg/atomicswap/cryptotypes"
	"transmutate.io/pkg/cryptocore/types"
)

// Output represents an output
type Output struct {
	TxID   types.Bytes `yaml:"txid"`
	N      uint32      `yaml:"n"`
	Amount uint64      `yaml:"amount"`
}

type fundsUTXO struct {
	Outputs    []*Output   `yaml:"outputs"`
	LockScript types.Bytes `yaml:"lock_script"`
}

func (f *fundsUTXO) CryptoType() cryptotypes.CryptoType { return cryptotypes.UTXO }

func newFundsUTXO() *fundsUTXO { return &fundsUTXO{Outputs: make([]*Output, 0, 4)} }
