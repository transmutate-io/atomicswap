package script

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ltcsuite/ltcd/btcec"
	"github.com/stretchr/testify/require"
	"transmutate.io/pkg/atomicswap/hash"
)

func TestInt64(t *testing.T) {
	b := make([]byte, 0, 1024)
	for i := int64(0); i < 10; i++ {
		b = append(b, Int64(i)...)
	}
	s, err := DisassembleString(b)
	require.NoError(t, err, "unexpected error")
	require.Equal(t, "0 1 2 3 4 5 6 7 8 9", s, "mismatching script code")
}

func TestBytes(t *testing.T) {
	b := make([]byte, 0, 1024)
	for _, i := range [][]byte{{1}, {2, 3}, {4, 5, 6}, {7, 8, 9, 10}, {11, 12, 13, 14, 15}} {
		b = append(b, Data(i)...)
	}
	s, err := DisassembleString(b)
	require.NoError(t, err, "unexpected error")
	require.Equal(t, "1 0203 040506 0708090a 0b0c0d0e0f", s, "mismatching script code")
}

func TestIf(t *testing.T) {
	zero := Int64(0)
	one := Int64(1)
	s, err := DisassembleString(If(one, zero))
	require.NoError(t, err, "unexpected error")
	require.Equal(t, "OP_IF 1 OP_ELSE 0 OP_ENDIF", s, "mismatching script code")
}

func TestP2PKH(t *testing.T) {
	key, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err, "unexpected error")
	s, err := DisassembleString(P2PKHPublicBytes(key.PubKey().SerializeCompressed()))
	require.NoError(t, err, "unexpected error")
	require.Equal(t, "OP_DUP OP_HASH160 "+hex.EncodeToString(hash.Hash160(key.PubKey().SerializeCompressed()))+" OP_EQUALVERIFY OP_CHECKSIG", s, "mismatching script code")
}

func TestP2PK(t *testing.T) {
	key, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err, "unexpected error")
	s, err := DisassembleString(P2PKPublicBytes(key.PubKey().SerializeCompressed()))
	require.NoError(t, err, "unexpected error")
	require.Equal(t, hex.EncodeToString(key.PubKey().SerializeCompressed())+" OP_CHECKSIG", s, "mismatching script code")
}

func TestP2MS(t *testing.T) {
	const (
		NKEYS = 3
		RKEYS = 2
	)
	keys := make([][]byte, 0, NKEYS)
	keysHex := make([]string, 0, NKEYS)
	for i := 0; i < NKEYS; i++ {
		key, err := btcec.NewPrivateKey(btcec.S256())
		require.NoError(t, err, "unexpected error")
		k := key.PubKey().SerializeCompressed()
		keys = append(keys, k)
		keysHex = append(keysHex, hex.EncodeToString(k))
	}
	s, err := DisassembleString(P2MS(false, RKEYS, keys...))
	require.NoError(t, err, "unexpected error")
	exp := fmt.Sprintf("%d %s %d OP_CHECKMULTISIG", RKEYS, strings.Join(keysHex, " "), NKEYS)
	require.Equal(t, exp, s, "mismatching script code")
}

func TestP2SH(t *testing.T) {
	sc := Int64(1234)
	s, err := DisassembleString(P2SHScript(sc))
	require.NoError(t, err, "unexpected error")
	require.Equal(t, "OP_HASH160 "+hex.EncodeToString(hash.Hash160(sc))+" OP_EQUAL", s, "mismatching script code")
}

func TestLockTime(t *testing.T) {
	n := time.Now()
	s, err := DisassembleString(LockTimeTime(n))
	require.NoError(t, err, "unexpected error")
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(n.Unix()))
	require.Equal(t, hex.EncodeToString(b)+" OP_CHECKLOCKTIMEVERIFY OP_DROP", s, "mismatching script code")
}
