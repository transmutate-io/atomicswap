package tx

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/internal/testutil"
	"github.com/transmutate-io/atomicswap/key"
	"github.com/transmutate-io/cryptocore/types"
	"gopkg.in/yaml.v2"
)

func init() {
	for _, i := range testutil.Cryptos {
		if err := testutil.SetupMinerAddress(i); err != nil {
			panic(err)
		}
	}
}

func testTxUTXOMarshalUnmarshal(t *testing.T, tc *testutil.Crypto, tx Tx) {
	c := testutil.MustParseCrypto(t, tc.Name)
	// inputs and outputs
	txUTXO, ok := tx.TxUTXO()
	require.True(t, ok, "expecting an utxo tx")
	txUTXO.AddInput(testutil.MustReadRandom(t, 32), 1, []byte{0x52, 0x52, 0x93, 0x54, 0x87}, 0)
	txUTXO.AddInput(testutil.MustReadRandom(t, 32), 3, []byte{0x53, 0x52, 0x93, 0x55, 0x87}, 0)
	txUTXO.AddOutput(100000000, []byte{0x54, 0x52, 0x93, 0x56, 0x87})
	k := testutil.MustNewPrivateKey(t, c)
	for i := 0; i < 2; i++ {
		_, err := txUTXO.InputSignature(i, 0, k)
		require.NoError(t, err, "can't sign input")
	}
	// set lock time
	txUTXO.SetLockTime(time.Now().UTC())
	// copy
	tx2 := tx.Copy()
	require.Equal(t, tx, tx2, "transactions mismatch")
	// serialize
	b, err := tx.Serialize()
	require.NoError(t, err, "can't serialize")
	b2, err := tx2.Serialize()
	require.NoError(t, err, "can't serialize")
	require.Equal(t, b, b2, "transactions mismatch")
	// marshal
	b, err = yaml.Marshal(tx)
	require.NoError(t, err, "can't marshal")
	// unmarshal
	err = yaml.Unmarshal(b, tx2)
	require.NoError(t, err, "can't unmarshal")
	// serialize and compare
	b, err = tx.Serialize()
	require.NoError(t, err, "can't serialize")
	b2, err = tx2.Serialize()
	require.NoError(t, err, "can't serialize")
	require.Equal(t, b, b2, "transactions mismatch")
}

func testMarshalUnmarshal(t *testing.T, tc *testutil.Crypto) {
	// parse crypto
	crypto, err := cryptos.Parse(tc.Name)
	require.NoError(t, err, "can't parse crypto")
	// new tx
	tx, err := New(crypto)
	require.NoError(t, err, "can't create a new tx")
	switch cryptoType := crypto.Type; cryptoType {
	case cryptos.UTXO:
		testTxUTXOMarshalUnmarshal(t, tc, tx)
	default:
		t.Errorf("unknown crypto type: %v", cryptoType)
		return
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	for _, i := range testutil.Cryptos {
		t.Run(i.Name, func(t *testing.T) { testMarshalUnmarshal(t, i) })
	}
}

func testP2PKH(t *testing.T, tc *testutil.Crypto) {
	// parse crypto
	c := testutil.MustParseCrypto(t, tc.Name)
	// generate new key
	k1 := testutil.MustNewPrivateKey(t, c)
	t.Logf("new key: %s %s\n", hex.EncodeToString(k1.Public().SerializeCompressed()), k1.Public().KeyData().Hex())
	// deposit address
	addr1 := testutil.MustP2PKHAddress(t, tc, k1.Public().KeyData())
	t.Logf("new key address: %s\n", addr1)
	// send to address
	txids := make([][]byte, 0, 2)
	for i := 0; i < 2; i++ {
		amt := types.Amount("1")
		testutil.MustEnsureBalance(t, tc, amt)
		txid, err := tc.Client.SendToAddress(addr1, amt)
		require.NoError(t, err, "can't send to address")
		testutil.MustGenerateBlocks(t, tc, 1)
		txids = append(txids, txid)
	}
	if tc.ConfirmBlocks > 0 {
		testutil.MustGenerateBlocks(t, tc, 1)
	}
	t.Logf("sent to address %s\n", addr1)
	// find txs
	outputs := testutil.MustFindIdxs(t, tc, txids, addr1)
	t.Logf("outputs found:\n")
	for i, out := range outputs {
		t.Logf("    %s %d (%s %s)\n", hex.EncodeToString(txids[i]), out.N(), out.Value(), c.Short)
	}
	k2, err := key.NewPrivate(c)
	require.NoError(t, err, "can't generate new key")
	t.Logf("new key: %s %s\n", hex.EncodeToString(k2.Public().SerializeCompressed()), k2.Public().KeyData().Hex())
	addr2 := testutil.MustP2PKHAddress(t, tc, k2.Public().KeyData())
	t.Logf("new address generated %s\n", addr2)
	// new engine
	gen := testutil.MustNewGenerator(t, c)
	// new tx
	tx, err := New(c)
	require.NoError(t, err, "can't create new transaction")
	// output
	txUTXO, ok := tx.TxUTXO()
	require.True(t, ok, "expecting an utxo tx")
	txUTXO.AddOutput(199000000, gen.P2PKHHash(k2.Public().KeyData()))
	// inputs
	for i, out := range outputs {
		err = txUTXO.AddInput(txids[i], uint32(out.N()), out.LockScript().Bytes(), out.Value().UInt64(c.Decimals))
		require.NoError(t, err, "can't add input")
	}
	// sign inputs
	for i := range outputs {
		err = txUTXO.SignP2PKHInput(i, 1, k1)
		require.NoError(t, err, "can't sign input")
	}
	// serialize
	b, err := tx.Serialize()
	require.NoError(t, err, "can't serialize")
	t.Logf("tx: %s\n", hex.EncodeToString(b))
	// send
	txid, err := testutil.SendRawTransaction(t, tc, b, 1)
	require.NoError(t, err, "can't send raw transaction")
	t.Logf("txid: %s\n", hex.EncodeToString(txid))
}

func TestP2PKH(t *testing.T) {
	for _, i := range testutil.Cryptos {
		t.Run(i.Name, func(t *testing.T) { testP2PKH(t, i) })
	}
}

func testP2SH(t *testing.T, tc *testutil.Crypto) {
	// parse crypto
	c := testutil.MustParseCrypto(t, tc.Name)
	// generate new key
	k := testutil.MustNewPrivateKey(t, c)
	// new engine
	gen := testutil.MustNewGenerator(t, c)
	// inputs script
	s := gen.P2PKHHash(k.Public().KeyData())
	// deposit address
	depositAddr := testutil.MustP2SHAddress(t, tc, s)
	// send to address
	txids := make([][]byte, 0, 2)
	amt := types.Amount("1")
	for j := 0; j < 2; j++ {
		testutil.MustEnsureBalance(t, tc, amt)
		txid, err := tc.Client.SendToAddress(depositAddr, amt)
		require.NoError(t, err, "can't send to address")
		testutil.MustGenerateBlocks(t, tc, 1)
		txids = append(txids, txid)
	}
	if tc.ConfirmBlocks > 0 {
		testutil.MustGenerateBlocks(t, tc, 1)
	}
	t.Logf("sent to address %s\n", depositAddr)
	// find txs
	outputs := testutil.MustFindIdxs(t, tc, txids, depositAddr)
	for i, out := range outputs {
		t.Logf("    %s %d (%s %s)\n", hex.EncodeToString(txids[i]), out.N(), out.Value(), c.Short)
	}
	// new tx
	tx, err := New(c)
	require.NoError(t, err, "can't create new transaction")
	txUTXO, ok := tx.TxUTXO()
	require.True(t, ok, "expecting an utxo tx")
	// output
	txUTXO.AddOutput(199900000, gen.P2PKHHash(k.Public().KeyData()))
	// inputs
	for i, out := range outputs {
		err = txUTXO.AddInput(txids[i], uint32(out.N()), s, out.Value().UInt64(c.Decimals))
		require.NoError(t, err, "can't add input")
	}
	// sign inputs
	for i := range outputs {
		sig, err := txUTXO.InputSignature(i, 1, k)
		require.NoError(t, err, "can't sign input")
		txUTXO.SetInputSignatureScript(i, gen.P2SHRedeem(s, sig, k.Public().SerializeCompressed()))
	}
	// serialize
	b, err := tx.Serialize()
	require.NoError(t, err, "can't serialize")
	t.Logf("tx: %s\n", hex.EncodeToString(b))
	// send
	txid, err := testutil.SendRawTransaction(t, tc, b, 1)
	require.NoError(t, err, "can't send raw transaction")
	t.Logf("txid: %s\n", hex.EncodeToString(txid))
}

func TestP2SH(t *testing.T) {
	for _, i := range testutil.Cryptos {
		t.Run(i.Name, func(t *testing.T) { testP2SH(t, i) })
	}
}
