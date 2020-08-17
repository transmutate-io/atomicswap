package tx

import (
	"bytes"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/key"
	"github.com/transmutate-io/atomicswap/script"
)

// tx represents a transaction
type txBTC wire.MsgTx

// NewBTC creates a new transaction for bitcoin
func NewBTC() (Tx, error) { return (*txBTC)(wire.NewMsgTx(wire.TxVersion)), nil }

func (tx *txBTC) tx() *wire.MsgTx { return (*wire.MsgTx)(tx) }

// AddOutput implement TxUTXO
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

// AddInput implement TxUTXO
func (tx *txBTC) AddInput(txID []byte, idx uint32, script []byte, _ uint64) error {
	h, err := chainhash.NewHash(bytesReverse(txID))
	if err != nil {
		return err
	}
	tx.tx().AddTxIn(wire.NewTxIn(wire.NewOutPoint(h, idx), script, nil))
	return nil
}

// InputSignature implement TxUTXO
func (tx *txBTC) InputSignature(idx int, hashType uint32, privKey key.Private) ([]byte, error) {
	return txscript.RawTxInSignature(
		tx.tx(),
		idx,
		tx.TxIn[idx].SignatureScript,
		txscript.SigHashType(hashType),
		privKey.Key().(*btcec.PrivateKey),
	)
}

// SetInputSequenceNumber implement TxUTXO
func (tx *txBTC) SetInputSequenceNumber(idx int, seq uint32) {
	tx.TxIn[idx].Sequence = seq
}

// InputSequenceNumber implement TxUTXO
func (tx *txBTC) InputSequenceNumber(idx int) uint32 { return tx.tx().TxIn[idx].Sequence }

// SetLockTimeUInt32 implement TxUTXO
func (tx *txBTC) SetLockTimeUInt32(lt uint32) { tx.LockTime = lt }

// SetLockTime implement TxUTXO
func (tx *txBTC) SetLockTime(lt time.Time) { tx.LockTime = uint32(lt.UTC().Unix()) }

// SetLockDuration implement TxUTXO
func (tx *txBTC) SetLockDuration(d time.Duration) { tx.SetLockTime(time.Now().UTC().Add(d)) }

// InputSignatureScript implement TxUTXO
func (tx *txBTC) InputSignatureScript(idx int) []byte {
	return tx.TxIn[idx].SignatureScript
}

// SetInputSignatureScript implement TxUTXO
func (tx *txBTC) SetInputSignatureScript(idx int, ss []byte) {
	tx.TxIn[idx].SignatureScript = ss
}

// SignP2PKInput implement TxUTXO
func (tx *txBTC) SignP2PKInput(idx int, hashType uint32, privKey key.Private) error {
	sig, err := tx.InputSignature(idx, hashType, privKey)
	if err != nil {
		return err
	}
	tx.SetInputSignatureScript(idx, script.NewGeneratorBTC().Data(sig))
	return nil
}

// SignP2PKHInput implement TxUTXO
func (tx *txBTC) SignP2PKHInput(idx int, hashType uint32, privKey key.Private) error {
	sig, err := tx.InputSignature(idx, hashType, privKey)
	if err != nil {
		return err
	}
	s := script.NewEngineBTC().
		Data(sig).
		Data(privKey.Public().SerializeCompressed()).
		Bytes()
	tx.SetInputSignatureScript(idx, s)
	return nil
}

// Serialize implement Serializer
func (tx *txBTC) Serialize() ([]byte, error) {
	r := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := tx.tx().Serialize(r); err != nil {
		return nil, err
	}
	return r.Bytes(), nil
}

// SerializedSize implement Serializer
func (tx *txBTC) SerializedSize() uint64 { return uint64(tx.tx().SerializeSize()) }

// TxUTXO implement Tx
func (tx *txBTC) TxUTXO() (TxUTXO, bool) { return tx, true }

// TxStateBased implement Tx
func (tx *txBTC) TxStateBased() (TxStateBased, bool) { return nil, false }

// Crypto implement Tx
func (tx *txBTC) Crypto() *cryptos.Crypto { return cryptos.Cryptos["bitcoin"] }

// Copy implement Tx
func (tx *txBTC) Copy() Tx { return (*txBTC)(tx.tx().Copy()) }
