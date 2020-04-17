package transaction

import "github.com/btcsuite/btcd/wire"

type ltcTx = btcTx

// NewLTC creates a new *ltcTx
func NewLTC() Tx {
	return &ltcTx{
		tx:           wire.NewMsgTx(wire.TxVersion),
		inputScripts: make([][]byte, 0, 8),
	}
}
