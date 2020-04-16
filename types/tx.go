package types

import (
	"transmutate.io/pkg/atomicswap/types/key"
)

type NewTxFunc = func() Tx

type Serializer interface {
	Serialize() ([]byte, error)
	SerializedSize() uint64
}

type Tx interface {
	Copy() Tx
	AddOutput(value uint64, script []byte)
	AddInput(txID []byte, idx uint32, script []byte) error
	InputSignature(idx int, hashType uint32, privKey key.Private) ([]byte, error)
	SignP2PKInput(idx int, hashType uint32, privKey key.Private) error
	SignP2PKHInput(idx int, hashType uint32, privKey key.Private) error
	SetP2SHInputPrefixes(idx int, pref ...[]byte) error
	AddP2SHInputPrefix(idx int, p []byte)
	SetP2SHInputSignatureScript(idx int, ss []byte)
	Tx() interface{}
	Serializer
}
