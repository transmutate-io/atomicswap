package transaction

import (
	"errors"
	"time"

	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/script"
)

type (
	NewTxFunc = func() (Tx, error)

	Serializer interface {
		// Serialize serializes the transaction
		Serialize() ([]byte, error)
		// SerializedSize returns the size of the serialized transaction
		SerializedSize() uint64
	}

	TxUTXO interface {
		// NewScript returns a new script engine
		NewScript() script.Engine
		// AddOutput adds an output to the transaction
		AddOutput(value uint64, script []byte)
		// AddInput adds an input to the transaction
		AddInput(txID []byte, idx uint32, script []byte) error
		// InputSignature returns the signature for an existing input
		InputSignature(idx int, hashType uint32, privKey key.Private) ([]byte, error)
		// SetInputSequenceNumber sets the sequence number for a given input
		SetInputSequenceNumber(idx int, seq uint32)
		// InputSequenceNumber returns the sequence number of a given input
		InputSequenceNumber(idx int) uint32
		// SetLockTimeUInt32 sets the locktime
		SetLockTimeUInt32(lt uint32)
		// SetLockTime sets the locktime
		SetLockTime(lt time.Time)
		// SetLockDuration sets the locktime as a duration (counting from time.Now().UTC())
		SetLockDuration(d time.Duration)
		// InputSignatureScript returns the signatureScript field of an input
		InputSignatureScript(idx int) []byte
		// SetInputSignatureScript sets the signatureScript field of an input
		SetInputSignatureScript(idx int, ss []byte)
		// SignP2PKInput signs an p2pk input
		SignP2PKInput(idx int, hashType uint32, privKey key.Private) error
		// SignP2PKHInput signs a p2pkh input
		SignP2PKHInput(idx int, hashType uint32, privKey key.Private) error
		// // AddInputPrefixes add prefixes to a p2sh input
		// AddInputPrefixes(idx int, p ...[]byte)
	}

	TxStateBased interface{}

	Tx interface {
		Serializer
		Copy() Tx

		Crypto() *cryptos.Crypto

		// Copy returns a copy of tx
		// TxUTXO returns a TxUTXO transaction
		TxUTXO() TxUTXO
		// TxStateBased returns a TxStateBased transaction
		TxStateBased() TxStateBased
		// Type returns the crypto/message type
	}
)

var (
	ErrNotStateBased = errors.New("not state based")
	ErrNotUTXO       = errors.New("not UTXO")
)

func NewTx(c *cryptos.Crypto) (Tx, error) {
	nf, ok := txFuncs[c.Name]
	if !ok {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return nf()
}
