package transaction

import "github.com/btcsuite/btcd/wire"

type ltcTx struct{ btcTx }

// NewLTC creates a new *ltcTx
func NewLTC() Tx {
	return &ltcTx{btcTx{
		tx:           wire.NewMsgTx(wire.TxVersion),
		inputScripts: make([][]byte, 0, 8),
	}}
}
