package script

import (
	"bytes"
	"time"

	"github.com/btcsuite/btcd/txscript"
	"transmutate.io/pkg/swapper/hash"
)

func bytesJoin(b ...[]byte) []byte { return bytes.Join(b, []byte{}) }

// P2PKHHash creates a new p2pkh script from a public key hash
func P2PKHHash(hash []byte) []byte {
	return bytesJoin(
		[]byte{txscript.OP_DUP, txscript.OP_HASH160},
		Data(hash),
		[]byte{txscript.OP_EQUALVERIFY, txscript.OP_CHECKSIG},
	)
}

// P2PKHPublicBytes creates a new p2pkh script from a public key bytes
func P2PKHPublicBytes(pub []byte) []byte { return P2PKHHash(hash.Hash160(pub)) }

// P2PKPublicBytes creates a new p2pk script from a public key bytes
func P2PKPublicBytes(pub []byte) []byte {
	return bytesJoin(
		Data(pub),
		[]byte{txscript.OP_CHECKSIG},
	)
}

// P2SHHash creates a new p2sh from a script hash
func P2SHHash(h []byte) []byte {
	return bytesJoin(
		[]byte{txscript.OP_HASH160},
		Data(h),
		[]byte{txscript.OP_EQUAL},
	)
}

// P2SHHash creates a new p2sh from a script
func P2SHScript(s []byte) []byte { return P2SHHash(hash.Hash160(s)) }

// LockTimeInt creates a new OP_CHECKLOCKTIMEVERIFY lock script from an int
func LockTimeInt(lock int64) []byte {
	return bytesJoin(
		Int64(lock),
		[]byte{txscript.OP_CHECKLOCKTIMEVERIFY},
	)
}

// LockTimeTime creates a new OP_CHECKLOCKTIMEVERIFY lock script from a time.Time
func LockTimeTime(t time.Time) []byte { return LockTimeInt(t.Unix()) }

// P2MS creates a new p2ms script from the provided keys
func P2MS(verify bool, nRequired int64, pubKeys ...[]byte) []byte {
	r := append(make([]byte, 0, 1024), Int64(nRequired)...)
	for _, i := range pubKeys {
		r = append(r, Data(i)...)
	}
	var checkOp []byte
	if verify {
		checkOp = []byte{txscript.OP_CHECKMULTISIGVERIFY}
	} else {
		checkOp = []byte{txscript.OP_CHECKMULTISIG}
	}
	return bytesJoin(
		r,
		Int64(int64(len(pubKeys))),
		checkOp,
	)
}

// HashLock creates a new hash1260 lock script
func HashLock(h []byte, verify bool) []byte {
	var checkOp []byte
	if verify {
		checkOp = []byte{txscript.OP_EQUALVERIFY}
	} else {
		checkOp = []byte{txscript.OP_EQUAL}
	}
	return bytesJoin([]byte{txscript.OP_HASH160}, Data(h), checkOp)
}

// HTLC creates a new hash time locked contract script
func HTLC(lockScript, tokenHash, timeLockedScript, hashLockedScript []byte) []byte {
	return If(
		bytesJoin(lockScript, timeLockedScript),
		bytesJoin(HashLock(tokenHash, true), hashLockedScript),
	)
}

// MSTLC creates a multisig time locked contract script
func MSTLC(lockScript, timeLockedScript []byte, nRequired int64, pubKeys ...[]byte) []byte {
	return If(
		bytesJoin(lockScript, timeLockedScript),
		P2MS(false, nRequired, pubKeys...),
	)
}

// SequenceInt creates a new OP_CHECKSEQUENCEVERIFY lock script from an int
func SequenceInt(lock int64) []byte {
	return bytesJoin(
		Int64(lock),
		[]byte{txscript.OP_CHECKSEQUENCEVERIFY},
	)
}

// Sequence  creates a new OP_CHECKSEQUENCEVERIFY lock script from a SequenceNumber
func Sequence(s SequenceNumber) []byte { return SequenceInt(int64(s)) }
