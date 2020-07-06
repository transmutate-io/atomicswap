package tx

import (
	"bytes"
	"time"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrec"
	"github.com/decred/dcrd/txscript/v3"
	"github.com/decred/dcrd/wire"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/key"
	"github.com/transmutate-io/atomicswap/script"
)

// tx represents a transaction
type txDCR wire.MsgTx

// NewDCR creates a new *txDCR
func NewDCR() (Tx, error) { return (*txDCR)(wire.NewMsgTx()), nil }

func (tx *txDCR) tx() *wire.MsgTx { return (*wire.MsgTx)(tx) }

// AddOutput adds an output to the transaction
func (tx *txDCR) AddOutput(value uint64, script []byte) {
	tx.tx().AddTxOut(wire.NewTxOut(int64(value), script))
}

// AddInput adds an input to the transaction
func (tx *txDCR) AddInput(txID []byte, idx uint32, script []byte, amount uint64) error {
	h, err := chainhash.NewHash(bytesReverse(txID))
	if err != nil {
		return err
	}
	tx.tx().AddTxIn(wire.NewTxIn(wire.NewOutPoint(h, idx, wire.TxTreeRegular), int64(amount), script))
	return nil
}

// InputSignature returns the signature for an existing input
func (tx *txDCR) InputSignature(idx int, hashType uint32, privKey key.Private) ([]byte, error) {
	return txscript.RawTxInSignature(
		tx.tx(),
		idx,
		tx.TxIn[idx].SignatureScript,
		txscript.SigHashType(hashType),
		privKey.Serialize(),
		dcrec.STEcdsaSecp256k1,
	)
}

// SetInputSequenceNumber sets the sequence number for a given input
func (tx *txDCR) SetInputSequenceNumber(idx int, seq uint32) {
	tx.TxIn[idx].Sequence = seq
}

// InputSequenceNumber returns the sequence number of a given input
func (tx *txDCR) InputSequenceNumber(idx int) uint32 { return tx.tx().TxIn[idx].Sequence }

// SetLockTimeUInt32 sets the locktime
func (tx *txDCR) SetLockTimeUInt32(lt uint32) { tx.LockTime = lt }

// SetLockTime sets the locktime
func (tx *txDCR) SetLockTime(lt time.Time) { tx.LockTime = uint32(lt.UTC().Unix()) }

// SetLockDuration sets the locktime as a duration (counting from time.Now().UTC())
func (tx *txDCR) SetLockDuration(d time.Duration) { tx.SetLockTime(time.Now().UTC().Add(d)) }

// InputSignatureScript returns the signatureScript field of an input
func (tx *txDCR) InputSignatureScript(idx int) []byte {
	return tx.TxIn[idx].SignatureScript
}

// SetInputSignatureScript sets the signatureScript field of an input
func (tx *txDCR) SetInputSignatureScript(idx int, ss []byte) {
	tx.TxIn[idx].SignatureScript = ss
}

// SignP2PKInput signs an p2pk input
func (tx *txDCR) SignP2PKInput(idx int, hashType uint32, privKey key.Private) error {
	sig, err := tx.InputSignature(idx, hashType, privKey)
	if err != nil {
		return err
	}
	tx.SetInputSignatureScript(idx, script.NewGeneratorDCR().Data(sig))
	return nil
}

// SignP2PKHInput signs a p2pkh input
func (tx *txDCR) SignP2PKHInput(idx int, hashType uint32, privKey key.Private) error {
	sig, err := tx.InputSignature(idx, hashType, privKey)
	if err != nil {
		return err
	}
	s := script.NewEngineDCR().
		Data(sig).
		Data(privKey.Public().SerializeCompressed()).
		Bytes()
	tx.SetInputSignatureScript(idx, s)
	return nil
}

// Serialize serializes the transaction
func (tx *txDCR) Serialize() ([]byte, error) {
	r := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := tx.tx().Serialize(r); err != nil {
		return nil, err
	}
	return r.Bytes(), nil
}

// SerializedSize returns the size of the serialized transaction
func (tx *txDCR) SerializedSize() uint64 { return uint64(tx.tx().SerializeSize()) }

// TxUTXO returns a TxUTXO transaction
func (tx *txDCR) TxUTXO() (TxUTXO, bool) { return tx, true }

// TxStateBased returns a TxStateBased transaction
func (tx *txDCR) TxStateBased() (TxStateBased, bool) { return nil, false }

func (tx *txDCR) Crypto() *cryptos.Crypto { return cryptos.Cryptos["bitcoin"] }

// Copy returns a copy of tx
func (tx *txDCR) Copy() Tx { return (*txDCR)(tx.tx().Copy()) }
