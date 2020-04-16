package types

import "github.com/btcsuite/btcd/wire"

type dogeTx = btcTx

// NewTxDOGE creates a new *dogeTx
func NewTxDOGE() Tx {
	return &dogeTx{
		tx:           wire.NewMsgTx(wire.TxVersion),
		inputScripts: make([][]byte, 0, 8),
	}
}
