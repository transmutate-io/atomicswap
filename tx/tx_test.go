package tx

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/hash"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/networks"
	"transmutate.io/pkg/atomicswap/script"
	"transmutate.io/pkg/cryptocore"
	"transmutate.io/pkg/cryptocore/types"
)

var testCryptos = []*struct {
	c         string
	cl        cryptocore.Client
	minerAddr string
}{
	{
		"bitcoin",
		cryptocore.NewClientBTC(
			"bitcoin-core-regtest.docker:4444",
			"admin",
			"pass",
			false,
		),
		"2N9ujTyhnCk8NC7jPJbu65o2rtUPe9eDgoD",
	},
	{
		"litecoin",
		cryptocore.NewClientBTC(
			"litecoin-regtest.docker:4444",
			"admin",
			"pass",
			false,
		),
		"QSizW8nuutY2JdzeBaCNA1tvyx2HzVehnm",
	},
	{
		"dogecoin",
		cryptocore.NewClientBTC(
			"dogecoin-regtest.docker:4444",
			"admin",
			"pass",
			false,
		),
		"n4c8B563PDoutGhmY8jEmkTUeLVN7hi2EG",
	},
	{
		"bitcoin-cash",
		cryptocore.NewClientBTCCash(
			"bitcoin-cash-regtest.docker:4444",
			"admin",
			"pass",
			false,
		),
		"qqj2f6xml8e27sgqz0gxhlwnfzh3fw7cw57a7x4gfz",
	},
}

func readRandom(sz int) []byte {
	r := make([]byte, sz)
	if n, err := rand.Read(r); err != nil {
		panic(err)
	} else if n != sz {
		panic("can't read enough random bytes")
	}
	return r
}

func testTxUTXOMarshalUnmarshal(t *testing.T, c *cryptos.Crypto, tx Tx) {
	// inputs and outputs
	txUTXO := tx.TxUTXO()
	txUTXO.AddInput(readRandom(32), 1, []byte{0x52, 0x52, 0x93, 0x54, 0x87}, 0)
	txUTXO.AddInput(readRandom(32), 3, []byte{0x53, 0x52, 0x93, 0x55, 0x87}, 0)
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

func TestMarshalUnmarshal(t *testing.T) {
	for _, i := range testCryptos {
		t.Run(i.c, func(t *testing.T) {
			crypto, err := cryptos.Parse(i.c)
			require.NoError(t, err, "can't parse crypto")
			// new tx
			tx, err := New(crypto)
			require.NoError(t, err, "can't create a new tx")
			switch cryptoType := crypto.Type; cryptoType {
			case cryptos.UTXO:
				testTxUTXOMarshalUnmarshal(t, crypto, tx)
			default:
				t.Errorf("unknown crypto type: %v", cryptoType)
				return
			}
		})
	}
}

func TestP2PKH(t *testing.T) {
	for _, i := range testCryptos {
		t.Run(i.c, func(t *testing.T) {
			// parse crypto
			c, err := cryptos.Parse(i.c)
			require.NoError(t, err, "can't parse crypto")
			// generate new key
			k, err := key.NewPrivate(c)
			require.NoError(t, err, "can't generate new key")
			t.Logf("new key: %s %s\n", hex.EncodeToString(k.Public().SerializeCompressed()), k.Public().KeyData().Hex())
			// deposit address
			depositAddr, err := networks.RegressionByName[i.c].P2PKH(k.Public().KeyData())
			require.NoError(t, err, "can't generate address")
			t.Logf("new key address: %s\n", depositAddr)
			// send to address
			txids := make([]types.Bytes, 0, 2)
			for j := 0; j < 2; j++ {
				txid, err := i.cl.SendToAddress(depositAddr, types.Amount("1"))
				require.NoError(t, err, "can't send to address")
				txids = append(txids, txid)
			}
			// generate blocks
			_, err = i.cl.GenerateToAddress(101, i.minerAddr)
			require.NoError(t, err, "can't generate blocks")
			t.Logf("sent to address %s\n", depositAddr)
			// find txs
			outputs, err := findIdxs(i.cl, txids, depositAddr)
			require.NoError(t, err, "can't find outputs")
			t.Logf("outputs found:\n")
			for i, out := range outputs {
				t.Logf("    %s %d (%s %s)\n", txids[i].Hex(), out.N, out.Value, c.Short)
			}
			// new tx
			tx, err := New(c)
			require.NoError(t, err, "can't create new transaction")
			txUTXO := tx.TxUTXO()
			k2, err := key.NewPrivate(c)
			require.NoError(t, err, "can't generate new key")
			t.Logf("new key: %s %s\n", hex.EncodeToString(k2.Public().SerializeCompressed()), k2.Public().KeyData().Hex())
			addr, err := networks.RegressionByName[c.Name].P2PKH(k2.Public().KeyData())
			require.NoError(t, err, "can't generate address")
			t.Logf("new address generated %s\n", addr)
			// new engine
			eng, err := script.NewEngine(c)
			require.NoError(t, err, "can't create scripting engine")
			// output
			txUTXO.AddOutput(199000000, eng.P2PKHHashBytes(k2.Public().KeyData()))
			// inputs
			amount := uint64(0)
			for i, out := range outputs {
				err = txUTXO.AddInput(txids[i], uint32(out.N), out.UnlockScript.Hex, out.Value.UInt64(8))
				require.NoError(t, err, "can't add input")
				amount += out.Value.UInt64(8)
			}
			// sign inputs
			for i := range outputs {
				err = txUTXO.SignP2PKHInput(i, 1, k)
				require.NoError(t, err, "can't sign input")
			}
			// serialize
			b, err := tx.Serialize()
			require.NoError(t, err, "can't serialize")
			t.Logf("tx: %s\n", hex.EncodeToString(b))
			// send
			txid, err := i.cl.SendRawTransaction(b)
			require.NoError(t, err, "can't send raw transaction")
			// generate blocks
			_, err = i.cl.GenerateToAddress(101, i.minerAddr)
			require.NoError(t, err, "can't generate blocks")
			t.Logf("txid: %s\n", txid.Hex())
		})
	}
}

func TestP2SH(t *testing.T) {
	for _, i := range testCryptos {
		t.Run(i.c, func(t *testing.T) {
			// parse crypto
			c, err := cryptos.Parse(i.c)
			require.NoError(t, err, "can't parse crypto")
			// generate new key
			k, err := key.NewPrivate(c)
			require.NoError(t, err, "can't generate new key")
			// new engine
			eng, err := script.NewEngine(c)
			require.NoError(t, err, "can't create scripting engine")
			// inputs script
			s := eng.P2PKHHashBytes(k.Public().KeyData())
			// deposit address
			depositAddr, err := networks.RegressionByName[i.c].P2SH(hash.Hash160(s))
			require.NoError(t, err, "can't generate address")
			// send to address
			txids := make([]types.Bytes, 0, 2)
			for j := 0; j < 2; j++ {
				txid, err := i.cl.SendToAddress(depositAddr, types.Amount("1"))
				require.NoError(t, err, "can't send to address")
				txids = append(txids, txid)
			}
			// generate blocks
			_, err = i.cl.GenerateToAddress(101, i.minerAddr)
			require.NoError(t, err, "can't generate blocks")
			t.Logf("sent to address %s\n", depositAddr)
			// find txs
			outputs, err := findIdxs(i.cl, txids, depositAddr)
			require.NoError(t, err, "can't find outputs")
			for i, out := range outputs {
				t.Logf("    %s %d (%s %s)\n", txids[i].Hex(), out.N, out.Value, c.Short)
			}
			// new tx
			tx, err := New(c)
			require.NoError(t, err, "can't create new transaction")
			txUTXO := tx.TxUTXO()
			// output
			txUTXO.AddOutput(199000000, eng.P2PKHHashBytes(k.Public().KeyData()))
			// inputs
			amount := uint64(0)
			for i, out := range outputs {
				err = txUTXO.AddInput(txids[i], uint32(out.N), s, out.Value.UInt64(8))
				require.NoError(t, err, "can't add input")
				amount += out.Value.UInt64(8)
			}
			// sign inputs
			for i := range outputs {
				sig, err := txUTXO.InputSignature(i, 1, k)
				require.NoError(t, err, "can't sign input")
				txUTXO.SetInputSignatureScript(i, eng.Reset().Data(sig).Data(k.Public().SerializeCompressed()).Data(s).Bytes())
			}
			// serialize
			b, err := tx.Serialize()
			require.NoError(t, err, "can't serialize")
			t.Logf("tx: %s\n", hex.EncodeToString(b))
			// send
			txid, err := i.cl.SendRawTransaction(b)
			require.NoError(t, err, "can't send raw transaction")
			// generate blocks
			_, err = i.cl.GenerateToAddress(101, i.minerAddr)
			require.NoError(t, err, "can't generate blocks")
			t.Logf("txid: %s\n", txid.Hex())
		})
	}
}

func findIdxs(cl cryptocore.Client, txids []types.Bytes, addr string) ([]*types.Output, error) {
	r := make([]*types.Output, 0, len(txids))
	for _, i := range txids {
		var found bool
		tx, err := cl.Transaction(i)
		if err != nil {
			return nil, err
		}
		for _, out := range tx.Outputs {
			for _, j := range out.UnlockScript.Addresses {
				if j == addr {
					r = append(r, out)
					found = true
					break
				}
			}
		}
		if !found {
			return nil, errors.New("not found")
		}
	}
	return r, nil
}
