package transaction

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"transmutate.io/pkg/atomicswap/cryptotypes"
)

func readRandom(sz int) []byte {
	r := make([]byte, sz)
	if n, err := rand.Read(r); err != nil {
		panic(err)
	} else if n != sz {
		panic("can't read enough random bytes")
	}
	return r
}

type testTx struct{ newTx NewTxFunc }

var testTxs = map[string]*testTx{
	"bitcoin":      &testTx{newTx: NewBTC},
	"litecoin":     &testTx{newTx: NewLTC},
	"dogecoin":     &testTx{newTx: NewDOGE},
	"bitcoin-cash": &testTx{newTx: NewBCH},
}

func testTxUTXO(t *testing.T, tx TxUTXO) {
	// random txid
	txId := readRandom(32)
	// inputs and outputs
	tx.AddInput(txId, 1, []byte{0x52, 0x52, 0x93, 0x54, 0x87})
	tx.AddInput(txId, 3, []byte{0x53, 0x52, 0x93, 0x55, 0x87})
	tx.AddOutput(100000000, []byte{0x54, 0x52, 0x93, 0x56, 0x87})
	// copy
	tx2 := tx.(Tx).Copy()
	require.Equal(t, tx, tx2, "transactions mismatch")
	require.Equal(t, tx.(Tx), tx2, "transactions mismatch")
	// serialize
	b, err := tx.(Tx).Serialize()
	require.NoError(t, err, "can't serialize")
	b2, err := tx2.Serialize()
	require.NoError(t, err, "can't serialize")
	require.Equal(t, b, b2, "transactions mismatch")
	require.Equal(t, tx.(Tx).SerializedSize(), tx2.SerializedSize(), "transactions mismatch")
}

func TestTx(t *testing.T) {
	for n, i := range testTxs {
		t.Run(n, func(t *testing.T) {
			// new tx
			tx := i.newTx()
			switch txType := tx.Type(); txType {
			case cryptotypes.UTXO:
				testTxUTXO(t, tx.TxUTXO())
			default:
				t.Errorf("unknown crypto type: %v", txType)
				return
			}
		})
	}
}
