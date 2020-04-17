package transaction

import (
	"errors"

	"transmutate.io/pkg/atomicswap/cryptotypes"
	"transmutate.io/pkg/atomicswap/key"
)

type (
	NewTxFunc = func() Tx

	Serializer interface {
		Serialize() ([]byte, error)
		SerializedSize() uint64
	}

	TxUTXO interface {
		AddOutput(value uint64, script []byte)
		AddInput(txID []byte, idx uint32, script []byte) error
		InputSignature(idx int, hashType uint32, privKey key.Private) ([]byte, error)
		SignP2PKInput(idx int, hashType uint32, privKey key.Private) error
		SignP2PKHInput(idx int, hashType uint32, privKey key.Private) error
		SetP2SHInputPrefixes(idx int, pref ...[]byte) error
		AddP2SHInputPrefix(idx int, p []byte)
		SetP2SHInputSignatureScript(idx int, ss []byte)
		SetLockTime(lt uint32)
		SetInputSequence(idx int, seq uint32)
	}

	TxStateBased interface{}

	Tx interface {
		Copy() Tx
		Type() cryptotypes.CryptoType
		TxUTXO() TxUTXO
		TxStateBased() TxStateBased
		Serializer
	}
)

var (
	ErrNotStateBased = errors.New("not state based")
	ErrNotUTXO       = errors.New("not UTXO")
)
