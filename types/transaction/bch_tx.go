package transaction

import (
	"bytes"

	"github.com/gcash/bchd/bchec"
	"github.com/gcash/bchd/chaincfg/chainhash"
	"github.com/gcash/bchd/txscript"
	"github.com/gcash/bchd/wire"
	"transmutate.io/pkg/atomicswap/script"
	"transmutate.io/pkg/atomicswap/types/key"
)

type bchTx struct {
	tx           *wire.MsgTx
	inputScripts [][]byte
}

// NewBCH creates a new *bchTx
func NewBCH() Tx {
	return &bchTx{
		tx:           wire.NewMsgTx(wire.TxVersion),
		inputScripts: make([][]byte, 0, 8),
	}
}

func (tx *bchTx) Tx() interface{} { return tx.tx }

func (tx *bchTx) Copy() Tx {
	return &bchTx{
		tx:           tx.tx.Copy(),
		inputScripts: copyInputScripts(tx.inputScripts),
	}
}

// AddOutput adds an output to the transaction
func (tx *bchTx) AddOutput(value uint64, script []byte) {
	tx.tx.AddTxOut(wire.NewTxOut(int64(value), script))
}

// AddInput adds an input to the transaction
func (tx *bchTx) AddInput(txID []byte, idx uint32, script []byte) error {
	txHash := bytesReverse(txID)
	h, err := chainhash.NewHash(txHash)
	if err != nil {
		return err
	}
	tx.tx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(h, idx), nil))
	tx.inputScripts = append(tx.inputScripts, script)
	return nil
}

// InputSignature signature for an existing input
func (tx *bchTx) InputSignature(idx int, hashType uint32, privKey key.Private) ([]byte, error) {
	key := privKey.Key().(*bchec.PrivateKey)
	return txscript.LegacyTxInSignature(
		tx.tx,
		idx,
		tx.inputScripts[idx],
		txscript.SigHashType(hashType)|txscript.SigHashForkID,
		key,
	)
}

// SignP2PKInput signs an p2pk input
func (tx *bchTx) SignP2PKInput(idx int, hashType uint32, privKey key.Private) error {
	sig, err := tx.InputSignature(idx, hashType, privKey)
	if err != nil {
		return err
	}
	s, err := script.Validate(script.Data(sig))
	if err != nil {
		return err
	}
	tx.tx.TxIn[idx].SignatureScript = s
	return nil
}

// SignP2PKHInput signs a p2pkh input
func (tx *bchTx) SignP2PKHInput(idx int, hashType uint32, privKey key.Private) error {
	sig, err := tx.InputSignature(idx, hashType, privKey)
	if err != nil {
		return err
	}
	s, err := script.Validate(bytesConcat(
		script.Data(sig),
		script.Data(privKey.Public().SerializeCompressed()),
	))
	if err != nil {
		return err
	}
	tx.tx.TxIn[idx].SignatureScript = s
	return nil
}

// SetP2SHInputPrefixes sets the prefix data for a p2sh input
func (tx *bchTx) SetP2SHInputPrefixes(idx int, pref ...[]byte) error {
	b := make([]byte, 0, 1024)
	for _, i := range pref {
		b = append(b, script.Data(i)...)
	}
	b = append(b, script.Data(tx.inputScripts[idx])...)
	b, err := script.Validate(b)
	if err != nil {
		return err
	}
	tx.tx.TxIn[idx].SignatureScript = b
	return nil
}

// AddP2SHInputPrefix add a prefix to a p2sh input
func (tx *bchTx) AddP2SHInputPrefix(idx int, p []byte) {
	var ss []byte
	if ss = tx.tx.TxIn[idx].SignatureScript; ss == nil {
		ss = []byte{}
	}
	tx.tx.TxIn[idx].SignatureScript = append(script.Data(p), ss...)
}

// SetP2SHInputSignatureScript sets the signatureScript field of a p2sh input
func (tx *bchTx) SetP2SHInputSignatureScript(idx int, ss []byte) {
	tx.tx.TxIn[idx].SignatureScript = ss
}

// Serialize serializes the transaction
func (tx *bchTx) Serialize() ([]byte, error) {
	r := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := tx.tx.Serialize(r); err != nil {
		return nil, err
	}
	return r.Bytes(), nil
}

func (tx *bchTx) SerializedSize() uint64 { return uint64(tx.tx.SerializeSize()) }

func (tx *bchTx) Type() TransactionType { return UTXO }

func (tx *bchTx) SetLockTime(lt uint32) { tx.tx.LockTime = lt }

func (tx *bchTx) SetInputSequence(idx int, seq uint32) { tx.tx.TxIn[idx].Sequence = seq }

func (tx *bchTx) TxUTXO() TxUTXO { return tx }
