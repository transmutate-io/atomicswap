package transaction

import (
	"bytes"
	"time"

	"github.com/gcash/bchd/bchec"
	"github.com/gcash/bchd/chaincfg/chainhash"
	"github.com/gcash/bchd/txscript"
	"github.com/gcash/bchd/wire"
	"transmutate.io/pkg/atomicswap/cryptotypes"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/script"
)

// tx represents a transaction
type txBCH wire.MsgTx

// NewBCH creates a new *txBCH
func NewBCH() Tx { return (*txBCH)(wire.NewMsgTx(wire.TxVersion)) }

func (tx *txBCH) tx() *wire.MsgTx { return (*wire.MsgTx)(tx) }

// AddOutput adds an output to the transaction
func (tx *txBCH) AddOutput(value uint64, script []byte) {
	tx.tx().AddTxOut(wire.NewTxOut(int64(value), script))
}

// AddInput adds an input to the transaction
func (tx *txBCH) AddInput(txID []byte, idx uint32, script []byte) error {
	h, err := chainhash.NewHash(bytesReverse(txID))
	if err != nil {
		return err
	}
	tx.tx().AddTxIn(wire.NewTxIn(wire.NewOutPoint(h, idx), script))
	return nil
}

// InputSignature signature for an existing input
func (tx *txBCH) InputSignature(idx int, hashType uint32, privKey key.Private) ([]byte, error) {
	key := privKey.Key().(*bchec.PrivateKey)
	return txscript.LegacyTxInSignature(
		(*wire.MsgTx)(tx),
		idx,
		(*wire.MsgTx)(tx).TxIn[idx].SignatureScript,
		txscript.SigHashType(hashType)|txscript.SigHashForkID,
		key,
	)
}

// SetInputSequenceNumber sets the sequence number for a given input
func (tx *txBCH) SetInputSequenceNumber(idx int, seq uint32) {
	tx.TxIn[idx].Sequence = seq
}

// InputSequenceNumber returns the sequence number of a given input
func (tx *txBCH) InputSequenceNumber(idx int) uint32 { return tx.tx().TxIn[idx].Sequence }

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
	s, err := tx.NewScript().
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
	s, err := tx.NewScript().
		Data(sig).
		Data(privKey.Serialize()).
		Validate()
	if err != nil {
		return err
	}
	tx.SetInputSignatureScript(idx, s)
	return nil
}

// // AddInputPrefixes add prefixes to a p2sh input
// func (tx *txBCH) AddInputPrefixes(idx int, p ...[]byte) {
// 	var ss []byte
// 	if ss = tx.InputSignatureScript(idx); ss == nil {
// 		ss = []byte{}
// 	}
// 	b := make([][]byte, 0, len(p)+1)
// 	for _, i := range p {
// 		b = append(b, script.Data(i))
// 	}
// 	b = append(b, ss)
// 	tx.SetInputSignatureScript(idx, bytesConcat(b...))
// }

// Serialize serializes the transaction
func (tx *txBCH) Serialize() ([]byte, error) {
	r := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := tx.tx().Serialize(r); err != nil {
		return nil, err
	}
	return r.Bytes(), nil
}

// SerializedSize returns the size of the serialized transaction
func (tx *txBCH) SerializedSize() uint64 { return uint64(tx.tx().SerializeSize()) }

// TxUTXO returns a TxUTXO transaction
func (tx *txBCH) TxUTXO() TxUTXO { return tx }

// TxStateBased returns a TxStateBased transaction
func (tx *txBCH) TxStateBased() TxStateBased { panic(ErrNotStateBased) }

// Type returns the crypto/message type
func (tx *txBCH) Type() cryptotypes.CryptoType { return cryptotypes.UTXO }

// Copy returns a copy of tx
func (tx *txBCH) Copy() Tx { return (*txBCH)(tx.tx().Copy()) }

// NewScript returns a new script engine
func (tx *txBCH) NewScript() script.Engine { return script.NewEngineBTC() }
