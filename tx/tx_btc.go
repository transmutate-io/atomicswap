package tx

import (
	"bytes"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/script"
)

// tx represents a transaction
type txBTC wire.MsgTx

// NewBTC creates a new *txBTC
func NewBTC() (Tx, error) { return (*txBTC)(wire.NewMsgTx(wire.TxVersion)), nil }

func (tx *txBTC) tx() *wire.MsgTx { return (*wire.MsgTx)(tx) }

// AddOutput adds an output to the transaction
func (tx *txBTC) AddOutput(value uint64, script []byte) {
	tx.tx().AddTxOut(wire.NewTxOut(int64(value), script))
}

func bytesReverse(b []byte) []byte {
	sz := len(b)
	r := make([]byte, sz)
	for i := 0; i < sz; i++ {
		r[i] = b[sz-1-i]
	}
	return r
}

// AddInput adds an input to the transaction
func (tx *txBTC) AddInput(txID []byte, idx uint32, script []byte) error {
	h, err := chainhash.NewHash(bytesReverse(txID))
	if err != nil {
		return err
	}
	tx.tx().AddTxIn(wire.NewTxIn(wire.NewOutPoint(h, idx), script, nil))
	return nil
}

// InputSignature returns the signature for an existing input
func (tx *txBTC) InputSignature(idx int, hashType uint32, privKey key.Private) ([]byte, error) {
	return txscript.RawTxInSignature(
		tx.tx(),
		idx,
		tx.TxIn[idx].SignatureScript,
		txscript.SigHashType(hashType),
		privKey.Key().(*btcec.PrivateKey),
	)
}

// SetInputSequenceNumber sets the sequence number for a given input
func (tx *txBTC) SetInputSequenceNumber(idx int, seq uint32) {
	tx.TxIn[idx].Sequence = seq
}

// InputSequenceNumber returns the sequence number of a given input
func (tx *txBTC) InputSequenceNumber(idx int) uint32 { return tx.tx().TxIn[idx].Sequence }

// SetLockTimeUInt32 sets the locktime
func (tx *txBTC) SetLockTimeUInt32(lt uint32) { tx.LockTime = lt }

// SetLockTime sets the locktime
func (tx *txBTC) SetLockTime(lt time.Time) { tx.LockTime = uint32(lt.UTC().Unix()) }

// SetLockDuration sets the locktime as a duration (counting from time.Now().UTC())
func (tx *txBTC) SetLockDuration(d time.Duration) { tx.SetLockTime(time.Now().UTC().Add(d)) }

// InputSignatureScript returns the signatureScript field of an input
func (tx *txBTC) InputSignatureScript(idx int) []byte {
	return tx.TxIn[idx].SignatureScript
}

// SetInputSignatureScript sets the signatureScript field of an input
func (tx *txBTC) SetInputSignatureScript(idx int, ss []byte) {
	tx.TxIn[idx].SignatureScript = ss
}

// SignP2PKInput signs an p2pk input
func (tx *txBTC) SignP2PKInput(idx int, hashType uint32, privKey key.Private) error {
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
func (tx *txBTC) SignP2PKHInput(idx int, hashType uint32, privKey key.Private) error {
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
// func (tx *txBTC) AddInputPrefixes(idx int, p ...[]byte) {
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
func (tx *txBTC) Serialize() ([]byte, error) {
	r := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := tx.tx().Serialize(r); err != nil {
		return nil, err
	}
	return r.Bytes(), nil
}

// SerializedSize returns the size of the serialized transaction
func (tx *txBTC) SerializedSize() uint64 { return uint64(tx.tx().SerializeSize()) }

// TxUTXO returns a TxUTXO transaction
func (tx *txBTC) TxUTXO() TxUTXO { return tx }

// TxStateBased returns a TxStateBased transaction
func (tx *txBTC) TxStateBased() TxStateBased { panic(ErrNotStateBased) }

func (tx *txBTC) Crypto() *cryptos.Crypto { return cryptos.Cryptos["bitcoin"] }

// Copy returns a copy of tx
func (tx *txBTC) Copy() Tx { return (*txBTC)(tx.tx().Copy()) }

// NewScript returns a new script engine
func (tx *txBTC) NewScript() script.Engine { return script.NewEngineBTC() }
