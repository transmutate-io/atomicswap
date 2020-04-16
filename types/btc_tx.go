package types

import (
	"bytes"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"transmutate.io/pkg/atomicswap/script"
	"transmutate.io/pkg/atomicswap/types/key"
)

// tx represents a transaction
type btcTx struct {
	tx           *wire.MsgTx
	inputScripts [][]byte
}

// NewTxBTC creates a new *btcTx
func NewTxBTC() Tx {
	return &btcTx{
		tx:           wire.NewMsgTx(wire.TxVersion),
		inputScripts: make([][]byte, 0, 8),
	}
}
func copyInputScripts(inputs [][]byte) [][]byte {
	r := make([][]byte, 0, 8)
	for _, i := range inputs {
		r = append(r, append(make([]byte, 0, len(i)), i...))
	}
	return r
}

func (tx *btcTx) Copy() Tx {
	return &btcTx{
		tx:           tx.tx.Copy(),
		inputScripts: copyInputScripts(tx.inputScripts),
	}
}

func (tx *btcTx) Tx() interface{} { return tx.tx }

// AddOutput adds an output to the transaction
func (tx *btcTx) AddOutput(value uint64, script []byte) {
	tx.tx.AddTxOut(wire.NewTxOut(int64(value), script))
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
func (tx *btcTx) AddInput(txID []byte, idx uint32, script []byte) error {
	txHash := bytesReverse(txID)
	h, err := chainhash.NewHash(txHash)
	if err != nil {
		return err
	}
	tx.tx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(h, idx), nil, nil))
	tx.inputScripts = append(tx.inputScripts, script)
	return nil
}

// InputSignature signature for an existing input
func (tx *btcTx) InputSignature(idx int, hashType uint32, privKey key.Private) ([]byte, error) {
	return txscript.RawTxInSignature(
		tx.tx,
		idx,
		tx.inputScripts[idx],
		txscript.SigHashType(hashType),
		privKey.Key().(*btcec.PrivateKey),
	)
}

// SignP2PKInput signs an p2pk input
func (tx *btcTx) SignP2PKInput(idx int, hashType uint32, privKey key.Private) error {
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

func bytesConcat(b ...[]byte) []byte { return bytes.Join(b, []byte{}) }

// SignP2PKHInput signs a p2pkh input
func (tx *btcTx) SignP2PKHInput(idx int, hashType uint32, privKey key.Private) error {
	sig, err := tx.InputSignature(idx, hashType, privKey)
	if err != nil {
		return err
	}
	s, err := script.Validate(bytesConcat(
		script.Data(sig),
		script.Data(privKey.Key().(*btcec.PublicKey).SerializeCompressed()),
	))
	if err != nil {
		return err
	}
	tx.tx.TxIn[idx].SignatureScript = s
	return nil
}

// SetP2SHInputPrefixes sets the prefix data for a p2sh input
func (tx *btcTx) SetP2SHInputPrefixes(idx int, pref ...[]byte) error {
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
func (tx *btcTx) AddP2SHInputPrefix(idx int, p []byte) {
	var ss []byte
	if ss = tx.tx.TxIn[idx].SignatureScript; ss == nil {
		ss = []byte{}
	}
	tx.tx.TxIn[idx].SignatureScript = append(script.Data(p), ss...)
}

// SetP2SHInputSignatureScript sets the signatureScript field of a p2sh input
func (tx *btcTx) SetP2SHInputSignatureScript(idx int, ss []byte) {
	tx.tx.TxIn[idx].SignatureScript = ss
}

// Serialize serializes the transaction
func (tx *btcTx) Serialize() ([]byte, error) {
	r := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := tx.tx.Serialize(r); err != nil {
		return nil, err
	}
	return r.Bytes(), nil
}

func (tx *btcTx) SerializedSize() uint64 { return uint64(tx.tx.SerializeSize()) }
