package transaction

import "github.com/btcsuite/btcd/wire"

type dogeTx = btcTx

// NewDOGE creates a new *dogeTx
func NewDOGE() Tx {
	return &dogeTx{
		tx:           wire.NewMsgTx(wire.TxVersion),
		inputScripts: make([][]byte, 0, 8),
	}
}
