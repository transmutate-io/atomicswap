package tx

import (
	"bytes"
	"time"

	"github.com/gcash/bchd/bchec"
	"github.com/gcash/bchd/chaincfg/chainhash"
	"github.com/gcash/bchd/txscript"
	"github.com/gcash/bchd/wire"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/key"
	"github.com/transmutate-io/atomicswap/script"
)

// tx represents a transaction
type txBCH struct {
	*wire.MsgTx
	InputsAmounts []uint64
}

// NewBCH creates a new transaction for bitcoin-cash
func NewBCH() (Tx, error) {
	return &txBCH{
		MsgTx:         wire.NewMsgTx(wire.TxVersion),
		InputsAmounts: make([]uint64, 0, 16),
	}, nil
}

// AddOutput implement TxUTXO
func (tx *txBCH) AddOutput(value uint64, script []byte) {
	tx.MsgTx.AddTxOut(wire.NewTxOut(int64(value), script))
}

// AddInput implement TxUTXO
func (tx *txBCH) AddInput(txID []byte, idx uint32, script []byte, amount uint64) error {
	h, err := chainhash.NewHash(bytesReverse(txID))
	if err != nil {
		return err
	}
	tx.MsgTx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(h, idx), script))
	tx.InputsAmounts = append(tx.InputsAmounts, amount)
	return nil
}

// InputSignature implement TxUTXO
func (tx *txBCH) InputSignature(idx int, hashType uint32, privKey key.Private) ([]byte, error) {
	return txscript.RawTxInECDSASignature(
		tx.MsgTx,
		idx,
		tx.TxIn[idx].SignatureScript,
		txscript.SigHashType(hashType)|txscript.SigHashForkID,
		privKey.Key().(*bchec.PrivateKey),
		int64(tx.InputsAmounts[idx]),
	)
}

// SetInputSequenceNumber implement TxUTXO
func (tx *txBCH) SetInputSequenceNumber(idx int, seq uint32) {
	tx.TxIn[idx].Sequence = seq
}

// InputSequenceNumber implement TxUTXO
func (tx *txBCH) InputSequenceNumber(idx int) uint32 { return tx.MsgTx.TxIn[idx].Sequence }

// SetLockTimeUInt32 implement TxUTXO
func (tx *txBCH) SetLockTimeUInt32(lt uint32) { tx.LockTime = lt }

// SetLockTime implement TxUTXO
func (tx *txBCH) SetLockTime(lt time.Time) { tx.LockTime = uint32(lt.UTC().Unix()) }

// SetLockDuration implement TxUTXO
func (tx *txBCH) SetLockDuration(d time.Duration) { tx.SetLockTime(time.Now().UTC().Add(d)) }

// InputSignatureScript implement TxUTXO
func (tx *txBCH) InputSignatureScript(idx int) []byte {
	return tx.TxIn[idx].SignatureScript
}

// SetInputSignatureScript implement TxUTXO
func (tx *txBCH) SetInputSignatureScript(idx int, ss []byte) {
	tx.TxIn[idx].SignatureScript = ss
}

// SignP2PKInput implement TxUTXO
func (tx *txBCH) SignP2PKInput(idx int, hashType uint32, privKey key.Private) error {
	sig, err := tx.InputSignature(idx, hashType, privKey)
	if err != nil {
		return err
	}
	tx.SetInputSignatureScript(idx, script.NewGeneratorBCH().Data(sig))
	return nil
}

// SignP2PKHInput implement TxUTXO
func (tx *txBCH) SignP2PKHInput(idx int, hashType uint32, privKey key.Private) error {
	sig, err := tx.InputSignature(idx, hashType, privKey)
	if err != nil {
		return err
	}
	s := script.NewEngineBCH().
		Data(sig).
		Data(privKey.Public().SerializeCompressed()).
		Bytes()
	tx.SetInputSignatureScript(idx, s)
	return nil
}

// Serialize implement Serializer
func (tx *txBCH) Serialize() ([]byte, error) {
	r := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := tx.MsgTx.Serialize(r); err != nil {
		return nil, err
	}
	return r.Bytes(), nil
}

// SerializedSize implement Serializer
func (tx *txBCH) SerializedSize() uint64 { return uint64(tx.MsgTx.SerializeSize()) }

// TxUTXO implement Tx
func (tx *txBCH) TxUTXO() (TxUTXO, bool) { return tx, true }

// TxStateBased implement Tx
func (tx *txBCH) TxStateBased() (TxStateBased, bool) { return nil, false }

func (tx *txBCH) Crypto() *cryptos.Crypto { return cryptos.Cryptos["bitcoin"] }

// Copy implement Tx
func (tx *txBCH) Copy() Tx {
	r := &txBCH{
		MsgTx:         tx.MsgTx.Copy(),
		InputsAmounts: make([]uint64, 0, len(tx.InputsAmounts)),
	}
	for _, i := range tx.InputsAmounts {
		r.InputsAmounts = append(r.InputsAmounts, i)
	}
	return r
}
