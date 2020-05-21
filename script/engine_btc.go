package script

import (
	"bytes"
	"strings"
	"time"

	"github.com/btcsuite/btcd/txscript"
	"transmutate.io/pkg/atomicswap/hash"
)

type engineBTC struct{ b []byte }

func NewEngineBTC() Engine { return &engineBTC{b: make([]byte, 0, 1024)} }

func (eng *engineBTC) Reset() Engine {
	eng.b = make([]byte, 0, 1024)
	return eng
}

func (eng *engineBTC) Validate() ([]byte, error) {
	return txscript.NewScriptBuilder().AddOps(eng.b).Script()
}

// // Disassemble a script into a string
func (eng *engineBTC) DisassembleString(s []byte) (string, error) {
	return txscript.DisasmString(s)
}

// // DisassembleStrings disassembles a script into a string slice
func (eng *engineBTC) DisassembleStrings(s []byte) ([]string, error) {
	r, err := eng.DisassembleString(s)
	if err != nil {
		return nil, err
	}
	return strings.Split(r, " "), nil
}

func (eng *engineBTC) AppendBytes(b []byte) []byte { eng.b = append(eng.b, b...); return eng.b }

func (eng *engineBTC) Append(b []byte) Engine { eng.b = append(eng.b, b...); return eng }

func (eng *engineBTC) Bytes() []byte { return eng.b }

func (eng *engineBTC) SetBytes(b []byte) Engine {
	eng.b = b
	return eng
}

func (eng *engineBTC) If(i []byte, e []byte) Engine {
	eng.Append(eng.IfBytes(i, e))
	return eng
}

// If else statement. If e is nil an else branch will not be present
func (eng *engineBTC) IfBytes(i []byte, e []byte) []byte {
	r := append(make([]byte, 0, len(i)+len(e)+3), txscript.OP_IF)
	r = append(r, i...)
	if e != nil {
		r = append(r, txscript.OP_ELSE)
		r = append(r, e...)
	}
	return append(r, txscript.OP_ENDIF)
}

func (eng *engineBTC) Data(b []byte) Engine {
	eng.Append(eng.DataBytes(b))
	return eng
}

// Data adds bytes as data
func (eng *engineBTC) DataBytes(b []byte) []byte {
	r, _ := txscript.NewScriptBuilder().AddData(b).Script()
	return r
}

func (eng *engineBTC) Int64Bytes(n int64) []byte {
	b, _ := txscript.NewScriptBuilder().AddInt64(n).Script()
	return b
}

func (eng *engineBTC) Int64(n int64) Engine {
	eng.Append(eng.Int64Bytes(n))
	return eng
}

func (eng *engineBTC) ParseInt64(v []byte) (int64, error) {
	if len(v) == 0 {
		return 0, nil
	}
	var result int64
	for i, val := range v {
		result |= int64(val) << uint8(8*i)
	}
	if v[len(v)-1]&0x80 != 0 {
		result &= ^(int64(0x80) << uint8(8*(len(v)-1)))
		return -result, nil
	}
	return result, nil
}

func bytesJoin(b ...[]byte) []byte { return bytes.Join(b, []byte{}) }

func (eng *engineBTC) P2PKHHash(hash []byte) Engine {
	eng.Append(eng.P2PKHHashBytes(hash))
	return eng
}

func (eng *engineBTC) P2PKHHashBytes(hash []byte) []byte {
	return bytesJoin(
		[]byte{txscript.OP_DUP, txscript.OP_HASH160},
		eng.DataBytes(hash),
		[]byte{txscript.OP_EQUALVERIFY, txscript.OP_CHECKSIG},
	)
}
func (eng *engineBTC) P2PKHPublic(pub []byte) Engine {
	eng.Append(eng.P2PKHPublicBytes(pub))
	return eng
}

func (eng *engineBTC) P2PKHPublicBytes(pub []byte) []byte {
	return eng.P2PKHHashBytes(hash.Hash160(pub))
}

func (eng *engineBTC) P2PKPublic(pub []byte) Engine {
	eng.Append(eng.P2PKPublicBytes(pub))
	return eng
}

func (eng *engineBTC) P2PKPublicBytes(pub []byte) []byte {
	return bytesJoin(
		eng.DataBytes(pub),
		[]byte{txscript.OP_CHECKSIG},
	)
}

func (eng *engineBTC) P2SHHash(h []byte) Engine {
	eng.Append(eng.P2SHHashBytes(h))
	return eng
}

func (eng *engineBTC) P2SHHashBytes(h []byte) []byte {
	return bytesJoin(
		[]byte{txscript.OP_HASH160},
		eng.DataBytes(h),
		[]byte{txscript.OP_EQUAL},
	)
}

func (eng *engineBTC) P2SHScript(s []byte) Engine {
	eng.Append(eng.P2SHScriptBytes(s))
	return eng
}

func (eng *engineBTC) P2SHScriptBytes(s []byte) []byte {
	return eng.P2SHHashBytes(hash.Hash160(s))
}

func (eng *engineBTC) P2MS(verify bool, nRequired int64, pubKeys ...[]byte) Engine {
	eng.Append(eng.P2MSBytes(verify, nRequired, pubKeys...))
	return eng
}

func (eng *engineBTC) P2MSBytes(verify bool, nRequired int64, pubKeys ...[]byte) []byte {
	r := append(make([]byte, 0, 1024), eng.Int64Bytes(nRequired)...)
	for _, i := range pubKeys {
		r = append(r, eng.DataBytes(i)...)
	}
	var checkOp []byte
	if verify {
		checkOp = []byte{txscript.OP_CHECKMULTISIGVERIFY}
	} else {
		checkOp = []byte{txscript.OP_CHECKMULTISIG}
	}
	return bytesJoin(
		r,
		eng.Int64Bytes(int64(len(pubKeys))),
		checkOp,
	)
}

func (eng *engineBTC) LockTime(lock int64) Engine {
	eng.Append(eng.LockTimeBytes(lock))
	return eng
}

func (eng *engineBTC) LockTimeBytes(lock int64) []byte {
	return bytesJoin(
		eng.Int64Bytes(lock),
		[]byte{txscript.OP_CHECKLOCKTIMEVERIFY, txscript.OP_DROP},
	)
}

func (eng *engineBTC) LockTimeTime(t time.Time) Engine {
	eng.Append(eng.LockTimeTimeBytes(t))
	return eng
}

func (eng *engineBTC) LockTimeTimeBytes(t time.Time) []byte {
	return eng.LockTimeBytes(t.Unix())
}

func (eng *engineBTC) Sequence(lock int64) Engine {
	eng.Append(eng.SequenceBytes(lock))
	return eng
}

func (eng *engineBTC) SequenceBytes(lock int64) []byte {
	return bytesJoin(
		eng.Int64Bytes(lock),
		[]byte{txscript.OP_CHECKSEQUENCEVERIFY},
	)
}

func (eng *engineBTC) HashLock(h []byte, verify bool) Engine {
	eng.Append(eng.HashLockBytes(h, verify))
	return eng
}

func (eng *engineBTC) HashLockBytes(h []byte, verify bool) []byte {
	var checkOp []byte
	if verify {
		checkOp = []byte{txscript.OP_EQUALVERIFY}
	} else {
		checkOp = []byte{txscript.OP_EQUAL}
	}
	return bytesJoin([]byte{txscript.OP_HASH160}, eng.DataBytes(h), checkOp)
}

func (eng *engineBTC) HTLC(lockScript, tokenHash, timeLockedScript, hashLockedScript []byte) Engine {
	eng.Append(eng.HTLCBytes(lockScript, tokenHash, timeLockedScript, hashLockedScript))
	return eng
}

func (eng *engineBTC) HTLCBytes(lockScript, tokenHash, timeLockedScript, hashLockedScript []byte) []byte {
	return eng.IfBytes(
		bytesJoin(lockScript, timeLockedScript),
		bytesJoin(eng.HashLockBytes(tokenHash, true), hashLockedScript),
	)
}

func (eng *engineBTC) HTLCRedeemBytes(sig, key, token, locksScript []byte) []byte {
	return bytesJoin(
		eng.DataBytes(sig),
		eng.DataBytes(key),
		eng.DataBytes(token),
		eng.Int64Bytes(0),
		eng.DataBytes(locksScript),
	)
}

func (eng *engineBTC) HTLCRedeem(sig, key, token, locksScript []byte) Engine {
	eng.Append(eng.HTLCRedeemBytes(sig, key, token, locksScript))
	return eng
}

func (eng *engineBTC) HTLCRecoverBytes(sig, key, locksScript []byte) []byte {
	return bytesJoin(
		eng.DataBytes(sig),
		eng.DataBytes(key),
		eng.Int64Bytes(1),
		eng.DataBytes(locksScript),
	)
}

func (eng *engineBTC) HTLCRecover(sig, key, locksScript []byte) Engine {
	eng.Append(eng.HTLCRecoverBytes(sig, key, locksScript))
	return eng
}

func (eng *engineBTC) MSTLC(lockScript, timeLockedScript []byte, nRequired int64, pubKeys ...[]byte) Engine {
	eng.Append(eng.MSTLCBytes(lockScript, timeLockedScript, nRequired, pubKeys...))
	return eng
}

func (eng *engineBTC) MSTLCBytes(lockScript, timeLockedScript []byte, nRequired int64, pubKeys ...[]byte) []byte {
	return eng.IfBytes(
		bytesJoin(lockScript, timeLockedScript),
		eng.P2MSBytes(false, nRequired, pubKeys...),
	)
}
