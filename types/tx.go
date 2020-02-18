package types

import (
	"bytes"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"transmutate.io/pkg/atomicswap/script"
)

type Tx struct {
	tx           *wire.MsgTx
	inputScripts [][]byte
}

func NewTx() *Tx {
	const defaultSize = 8
	return &Tx{
		tx:           wire.NewMsgTx(wire.TxVersion),
		inputScripts: make([][]byte, 0, defaultSize),
	}
}

func (tx *Tx) AddOutput(value int64, script []byte) {
	tx.tx.AddTxOut(wire.NewTxOut(value, script))
}

func (tx *Tx) AddInput(txHash []byte, idx uint32, script []byte) error {
	h, err := chainhash.NewHash(txHash)
	if err != nil {
		return err
	}
	tx.tx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(h, idx), nil, nil))
	tx.inputScripts = append(tx.inputScripts, script)
	return nil
}

func (tx *Tx) InputSignature(idx int, hashType txscript.SigHashType, privKey *btcec.PrivateKey) ([]byte, error) {
	return txscript.RawTxInSignature(tx.tx, idx, tx.inputScripts[idx], hashType, privKey)
}

func (tx *Tx) SignP2PKInput(idx int, hashType txscript.SigHashType, privKey *btcec.PrivateKey) error {
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

func (tx *Tx) SignP2PKHInput(idx int, hashType txscript.SigHashType, privKey *btcec.PrivateKey) error {
	sig, err := tx.InputSignature(idx, hashType, privKey)
	if err != nil {
		return err
	}
	s, err := script.Validate(bytesConcat(
		script.Data(sig),
		script.Data(privKey.PubKey().SerializeCompressed()),
	))
	if err != nil {
		return err
	}
	tx.tx.TxIn[idx].SignatureScript = s
	return nil
}

func (tx *Tx) SignP2SHInput(idx int, sig []byte) error {
	s, err := script.Validate(bytesConcat(
		script.Data(sig),
		script.Data(tx.inputScripts[idx]),
	))
	if err != nil {
		return err
	}
	tx.tx.TxIn[idx].SignatureScript = s
	return nil
}

func (tx *Tx) Serialize() ([]byte, error) {
	r := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := tx.tx.Serialize(r); err != nil {
		return nil, err
	}
	return r.Bytes(), nil
}
