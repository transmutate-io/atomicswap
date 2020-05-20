package tx

import (
	"bytes"
	"time"

	"github.com/gcash/bchd/bchec"
	"github.com/gcash/bchd/chaincfg/chainhash"
	"github.com/gcash/bchd/txscript"
	"github.com/gcash/bchd/wire"
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/script"
)

// tx represents a transaction
type txBCH struct {
	*wire.MsgTx
	inputsAmounts []uint64
}

// NewBCH creates a new *txBCH
func NewBCH() (Tx, error) {
	return &txBCH{
		MsgTx:         wire.NewMsgTx(wire.TxVersion),
		inputsAmounts: make([]uint64, 0, 16),
	}, nil
}

// AddOutput adds an output to the transaction
func (tx *txBCH) AddOutput(value uint64, script []byte) {
	tx.MsgTx.AddTxOut(wire.NewTxOut(int64(value), script))
}

// AddInput adds an input to the transaction
func (tx *txBCH) AddInput(txID []byte, idx uint32, script []byte, amount uint64) error {
	h, err := chainhash.NewHash(bytesReverse(txID))
	if err != nil {
		return err
	}
	tx.MsgTx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(h, idx), script))
	tx.inputsAmounts = append(tx.inputsAmounts, amount)
	return nil
}

// InputSignature returns the signature for an existing input
func (tx *txBCH) InputSignature(idx int, hashType uint32, privKey key.Private) ([]byte, error) {
	return txscript.RawTxInECDSASignature(
		tx.MsgTx,
		idx,
		tx.TxIn[idx].SignatureScript,
		txscript.SigHashType(hashType)|txscript.SigHashForkID,
		privKey.Key().(*bchec.PrivateKey),
		int64(tx.inputsAmounts[idx]),
	)
}

// SetInputSequenceNumber sets the sequence number for a given input
func (tx *txBCH) SetInputSequenceNumber(idx int, seq uint32) {
	tx.TxIn[idx].Sequence = seq
}

// InputSequenceNumber returns the sequence number of a given input
func (tx *txBCH) InputSequenceNumber(idx int) uint32 { return tx.MsgTx.TxIn[idx].Sequence }

// SetLockTimeUInt32 sets the locktime
func (tx *txBCH) SetLockTimeUInt32(lt uint32) { tx.LockTime = lt }

// SetLockTime sets the locktime
func (tx *txBCH) SetLockTime(lt time.Time) { tx.LockTime = uint32(lt.UTC().Unix()) }

// SetLockDuration sets the locktime as a duration (counting from time.Now().UTC())
func (tx *txBCH) SetLockDuration(d time.Duration) { tx.SetLockTime(time.Now().UTC().Add(d)) }

// InputSignatureScript returns the signatureScript field of an input
func (tx *txBCH) InputSignatureScript(idx int) []byte {
	return tx.TxIn[idx].SignatureScript
}

// SetInputSignatureScript sets the signatureScript field of an input
func (tx *txBCH) SetInputSignatureScript(idx int, ss []byte) {
	tx.TxIn[idx].SignatureScript = ss
}

// SignP2PKInput signs an p2pk input
func (tx *txBCH) SignP2PKInput(idx int, hashType uint32, privKey key.Private) error {
	sig, err := tx.InputSignature(idx, hashType, privKey)
	if err != nil {
		return err
	}
	s, err := script.NewEngineBTC().
		Data(sig).
		Validate()
	if err != nil {
		return err
	}
	tx.SetInputSignatureScript(idx, s)
	return nil
}

// SignP2PKHInput signs a p2pkh input
func (tx *txBCH) SignP2PKHInput(idx int, hashType uint32, privKey key.Private) error {
	sig, err := tx.InputSignature(idx, hashType, privKey)
	if err != nil {
		return err
	}
	s, err := script.NewEngineBTC().
		Data(sig).
		Data(privKey.Public().SerializeCompressed()).
		Validate()
	if err != nil {
		return err
	}
	tx.SetInputSignatureScript(idx, s)
	return nil
}

// Serialize serializes the transaction
func (tx *txBCH) Serialize() ([]byte, error) {
	r := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := tx.MsgTx.Serialize(r); err != nil {
		return nil, err
	}
	return r.Bytes(), nil
}

// SerializedSize returns the size of the serialized transaction
func (tx *txBCH) SerializedSize() uint64 { return uint64(tx.MsgTx.SerializeSize()) }

// TxUTXO returns a TxUTXO transaction
func (tx *txBCH) TxUTXO() TxUTXO { return tx }

// TxStateBased returns a TxStateBased transaction
func (tx *txBCH) TxStateBased() TxStateBased { panic(ErrNotStateBased) }

func (tx *txBCH) Crypto() *cryptos.Crypto { return cryptos.Cryptos["bitcoin"] }

// Copy returns a copy of tx
func (tx *txBCH) Copy() Tx {
	r := &txBCH{
		MsgTx:         tx.MsgTx.Copy(),
		inputsAmounts: make([]uint64, 0, len(tx.inputsAmounts)),
	}
	for _, i := range tx.inputsAmounts {
		r.inputsAmounts = append(r.inputsAmounts, i)
	}
	return r
}
