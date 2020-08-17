package tx

import (
	"errors"
	"time"

	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/key"
)

type (
	// NewTxFunc represents a new transaction function
	NewTxFunc = func() (Tx, error)

	// Serializer represents a serializable object
	Serializer interface {
		// Serialize serializes the transaction
		Serialize() ([]byte, error)
		// SerializedSize returns the size of the serialized transaction
		SerializedSize() uint64
	}

	// TxUTXO represents a utxo transaction
	TxUTXO interface {
		// AddOutput adds an output to the transaction
		AddOutput(value uint64, script []byte)
		// AddInput adds an input to the transaction
		AddInput(txID []byte, idx uint32, script []byte, amount uint64) error
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
	}

	// TxStateBased represents a state based transaction
	TxStateBased interface{}

	Tx interface {
		Serializer

		// Copy returns a copy of the transaction
		Copy() Tx

		// Crypto returns the transaction crypto
		Crypto() *cryptos.Crypto

		// TxUTXO returns a TxUTXO transaction
		TxUTXO() (TxUTXO, bool)
		// TxStateBased returns a TxStateBased transaction
		TxStateBased() (TxStateBased, bool)
		// Type returns the crypto/message type
	}
)

var (
	// ErrNotStateBased is returned when the transaction is not state based
	ErrNotStateBased = errors.New("not state based")

	// ErrNotUTXO is returned when the transaction is not utxo
	ErrNotUTXO = errors.New("not UTXO")
)

// New returns a new transaction for the given crypto
func New(c *cryptos.Crypto) (Tx, error) {
	nf, ok := txFuncs[c.Name]
	if !ok {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return nf()
}
