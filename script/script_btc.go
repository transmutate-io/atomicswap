package script

import (
	"bytes"
	"time"

	"github.com/btcsuite/btcd/txscript"
	"github.com/transmutate-io/atomicswap/hash"
)

type generatorBTC struct{}

// NewGeneratorBTC returns a new bitcoin generator
func NewGeneratorBTC() Generator { return generatorBTC{} }

// If implement Generator
func (gen generatorBTC) If(i []byte, e []byte) []byte {
	r := append(make([]byte, 0, len(i)+len(e)+3), txscript.OP_IF)
	r = append(r, i...)
	if e != nil {
		r = append(r, txscript.OP_ELSE)
		r = append(r, e...)
	}
	return append(r, txscript.OP_ENDIF)
}

// Data implement Generator
func (gen generatorBTC) Data(b []byte) []byte {
	r, _ := txscript.NewScriptBuilder().AddData(b).Script()
	return r
}

// Int64 implement Generator
func (gen generatorBTC) Int64(n int64) []byte {
	b, _ := txscript.NewScriptBuilder().AddInt64(n).Script()
	return b
}

func bytesJoin(b ...[]byte) []byte { return bytes.Join(b, []byte{}) }

// P2PKHHash implement Generator
func (gen generatorBTC) P2PKHHash(hash []byte) []byte {
	return bytesJoin(
		[]byte{txscript.OP_DUP, txscript.OP_HASH160},
		gen.Data(hash),
		[]byte{txscript.OP_EQUALVERIFY, txscript.OP_CHECKSIG},
	)
}

// P2PKHPublic implement Generator
func (gen generatorBTC) P2PKHPublic(pub []byte) []byte {
	return gen.P2PKHHash(hash.NewBTC().Hash160(pub))
}

// P2PKPublic implement Generator
func (gen generatorBTC) P2PKPublic(pub []byte) []byte {
	return bytesJoin(
		gen.Data(pub),
		[]byte{txscript.OP_CHECKSIG},
	)
}

// P2SHHash implement Generator
func (gen generatorBTC) P2SHHash(h []byte) []byte {
	return bytesJoin(
		[]byte{txscript.OP_HASH160},
		gen.Data(h),
		[]byte{txscript.OP_EQUAL},
	)
}

// P2SHScript implement Generator
func (gen generatorBTC) P2SHScript(s []byte) []byte {
	return gen.P2SHHash(hash.NewBTC().Hash160(s))
}

// P2SHRedeem implement Generator
func (gen generatorBTC) P2SHRedeem(s []byte, pref ...[]byte) []byte {
	r := make([][]byte, 0, len(pref)+1)
	for _, i := range pref {
		r = append(r, gen.Data(i))
	}
	return bytesJoin(append(r, gen.Data(s))...)
}

// P2MS implement Generator
func (gen generatorBTC) P2MS(verify bool, nRequired int64, pubKeys ...[]byte) []byte {
	r := append(make([]byte, 0, 1024), gen.Int64(nRequired)...)
	for _, i := range pubKeys {
		r = append(r, gen.Data(i)...)
	}
	var checkOp []byte
	if verify {
		checkOp = []byte{txscript.OP_CHECKMULTISIGVERIFY}
	} else {
		checkOp = []byte{txscript.OP_CHECKMULTISIG}
	}
	return bytesJoin(
		r,
		gen.Int64(int64(len(pubKeys))),
		checkOp,
	)
}

// LockTime implement Generator
func (gen generatorBTC) LockTime(lock int64) []byte {
	return bytesJoin(
		gen.Int64(lock),
		[]byte{txscript.OP_CHECKLOCKTIMEVERIFY, txscript.OP_DROP},
	)
}

// LockTimeTime implement Generator
func (gen generatorBTC) LockTimeTime(t time.Time) []byte {
	return gen.LockTime(t.Unix())
}

// Sequence implement Generator
func (gen generatorBTC) Sequence(lock int64) []byte {
	return bytesJoin(
		gen.Int64(lock),
		[]byte{txscript.OP_CHECKSEQUENCEVERIFY},
	)
}

// HashLock implement Generator
func (gen generatorBTC) HashLock(h []byte, verify bool) []byte {
	var checkOp []byte
	if verify {
		checkOp = []byte{txscript.OP_EQUALVERIFY}
	} else {
		checkOp = []byte{txscript.OP_EQUAL}
	}
	return bytesJoin([]byte{txscript.OP_SHA256, txscript.OP_RIPEMD160}, gen.Data(h), checkOp)
}

// HTLC implement Generator
func (gen generatorBTC) HTLC(lockScript, tokenHash, timeLockedScript, hashLockedScript []byte) []byte {
	return gen.If(
		bytesJoin(lockScript, timeLockedScript),
		bytesJoin(gen.HashLock(tokenHash, true), hashLockedScript),
	)
}

// HTLCRedeem implement Generator
func (gen generatorBTC) HTLCRedeem(sig, key, token, locksScript []byte) []byte {
	return bytesJoin(
		gen.Data(sig),
		gen.Data(key),
		gen.Data(token),
		gen.Int64(0),
		gen.Data(locksScript),
	)
}

// HTLCRecover implement Generator
func (gen generatorBTC) HTLCRecover(sig, key, locksScript []byte) []byte {
	return bytesJoin(
		gen.Data(sig),
		gen.Data(key),
		gen.Int64(1),
		gen.Data(locksScript),
	)
}

// MSTLC implement Generator
func (gen generatorBTC) MSTLC(lockScript, timeLockedScript []byte, nRequired int64, pubKeys ...[]byte) []byte {
	return gen.If(
		bytesJoin(lockScript, timeLockedScript),
		gen.P2MS(false, nRequired, pubKeys...),
	)
}

type disassemblerBTC struct{}

// NewDisassemblerBTC returns a new Disassembler for bitcoin
func NewDisassemblerBTC() Disassembler { return disassemblerBTC{} }

// Disassemble a script into a string
func (dis disassemblerBTC) DisassembleString(s []byte) (string, error) {
	return txscript.DisasmString(s)
}

type intParserBTC struct{}

// NewIntParserBTC returns a new int64 parser for bitcoin
func NewIntParserBTC() IntParser { return intParserBTC{} }

// ParseInt64 implement IntParser
func (p intParserBTC) ParseInt64(v []byte) (int64, error) {
	if len(v) == 0 {
		return 0, nil
	}
	var r int64
	for i, val := range v {
		r |= int64(val) << uint8(8*i)
	}
	if v[len(v)-1]&0x80 != 0 {
		r &= ^(int64(0x80) << uint8(8*(len(v)-1)))
		return -r, nil
	}
	return r, nil
}

// NewEngineBTC returns a new *Engine for bitcoin
func NewEngineBTC() *Engine { return newEngine(NewGeneratorBTC()) }
