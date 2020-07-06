package script

import (
	"github.com/decred/dcrd/txscript/v3"
	"github.com/transmutate-io/atomicswap/hash"
)

type generatorDCR struct{ generatorBTC }

func NewGeneratorDCR() Generator { return &generatorDCR{generatorBTC: generatorBTC{}} }

func (gen *generatorDCR) P2SHScript(s []byte) []byte {
	return gen.P2SHHash(hash.NewDCR().Hash160(s))
}

func (gen *generatorDCR) P2PKHPublic(pub []byte) []byte {
	return gen.P2PKHHash(hash.NewDCR().Hash160(pub))
}

func (gen *generatorDCR) HashLock(h []byte, verify bool) []byte {
	var checkOp []byte
	if verify {
		checkOp = []byte{txscript.OP_EQUALVERIFY}
	} else {
		checkOp = []byte{txscript.OP_EQUAL}
	}
	return bytesJoin([]byte{txscript.OP_SHA256, txscript.OP_RIPEMD160}, gen.Data(h), checkOp)
}

func (gen *generatorDCR) HTLC(lockScript, tokenHash, timeLockedScript, hashLockedScript []byte) []byte {
	return gen.If(
		bytesJoin(lockScript, timeLockedScript),
		bytesJoin(gen.HashLock(tokenHash, true), hashLockedScript),
	)
}

type disassemblerDCR struct{}

func NewDisassemblerDCR() Disassembler { return disassemblerDCR{} }

// Disassemble a script into a string
func (dis disassemblerDCR) DisassembleString(s []byte) (string, error) {
	return txscript.DisasmString(s)
}

type intParserDCR struct{ intParserBTC }

func NewIntParserDCR() IntParser { return &intParserDCR{intParserBTC: intParserBTC{}} }

func NewEngineDCR() *Engine { return newEngine(NewGeneratorDCR()) }
