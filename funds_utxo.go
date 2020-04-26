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

func newFundsUTXO() *fundsUTXO { return &fundsUTXO{Outputs: make([]*Output, 0, 4)} }

func (f *fundsUTXO) CryptoType() cryptotypes.CryptoType { return cryptotypes.UTXO }

func (f *fundsUTXO) Funds() interface{} { return f.Outputs }

func (f *fundsUTXO) AddFunds(funds interface{}) {
	f.Outputs = append(f.Outputs, funds.(*Output))
}

func (f fundsUTXO) Lock() Lock { return fundsUTXOLock(f.LockScript) }

func (f *fundsUTXO) SetLock(lock Lock) { f.LockScript = lock.(fundsUTXOLock).Bytes() }

type fundsUTXOLock types.Bytes

func (fl fundsUTXOLock) Bytes() types.Bytes { return types.Bytes(fl) }

func (fl fundsUTXOLock) Data() interface{} { return fl.Bytes() }
