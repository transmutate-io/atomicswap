package transaction

import (
	"crypto/rand"
	"testing"
	"time"

	"transmutate.io/pkg/atomicswap/key"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"transmutate.io/pkg/atomicswap/cryptos"
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

var testCryptos = []string{
	"bitcoin",
	"litecoin",
	"dogecoin",
	"bitcoin-cash",
}

func testTxUTXO(t *testing.T, c *cryptos.Crypto, tx Tx) {
	// inputs and outputs
	txUTXO := tx.TxUTXO()
	txUTXO.AddInput(readRandom(32), 1, []byte{0x52, 0x52, 0x93, 0x54, 0x87})
	txUTXO.AddInput(readRandom(32), 3, []byte{0x53, 0x52, 0x93, 0x55, 0x87})
	txUTXO.AddOutput(100000000, []byte{0x54, 0x52, 0x93, 0x56, 0x87})
	k, err := key.NewPrivate(c)
	require.NoError(t, err, "can't create new private key")
	for i := 0; i < 2; i++ {
		_, err = txUTXO.InputSignature(i, 0, k)
		require.NoError(t, err, "can't sign input")
	}
	// set lock time
	txUTXO.SetLockTime(time.Now().UTC())
	// copy
	tx2 := tx.Copy()
	require.Equal(t, tx, tx2, "transactions mismatch")
	require.Equal(t, tx, tx2, "transactions mismatch")
	// serialize
	b, err := tx.Serialize()
	require.NoError(t, err, "can't serialize")
	b2, err := tx2.Serialize()
	require.NoError(t, err, "can't serialize")
	require.Equal(t, b, b2, "transactions mismatch")
	require.Equal(t, tx.SerializedSize(), tx2.SerializedSize(), "transactions mismatch")
	// marshal
	b, err = yaml.Marshal(tx)
	require.NoError(t, err, "can't marshal")
	// unmarshal
	err = yaml.Unmarshal(b, tx2)
	require.NoError(t, err, "can't unmarshal")
}

func TestTx(t *testing.T) {
	for _, name := range testCryptos {
		t.Run(name, func(t *testing.T) {
			crypto, err := cryptos.Parse(name)
			require.NoError(t, err, "can't parse crypto")
			// new tx
			tx, err := NewTx(crypto)
			require.NoError(t, err, "can't create a new tx")
			switch txType := tx.Type(); txType {
			case cryptotypes.UTXO:
				testTxUTXO(t, crypto, tx)
			default:
				t.Errorf("unknown crypto type: %v", txType)
				return
			}
		})
	}
}
