package types

import "github.com/btcsuite/btcd/wire"

type ltcTx = btcTx

// NewTxLTC creates a new *ltcTx
func NewTxLTC() Tx {
	return &ltcTx{
		tx:           wire.NewMsgTx(wire.TxVersion),
		inputScripts: make([][]byte, 0, 8),
	}
}
