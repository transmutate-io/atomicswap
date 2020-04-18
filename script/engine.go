package script

import (
	"time"
)

type (
	Disassembler interface {
		DisassembleString(s []byte) (string, error)
		DisassembleStrings(s []byte) ([]string, error)
	}

	Validator interface {
		Validate() ([]byte, error)
	}

	Logical interface {
		IfBytes(i, e []byte) []byte
		If(i, e []byte) Engine
	}

	Data interface {
		DataBytes(b []byte) []byte
		Data(b []byte) Engine

		Int64Bytes(n int64) []byte
		Int64(n int64) Engine
		ParseInt64(v []byte) (int64, error)
	}

	Bytes interface {
		Append(b []byte) []byte
		Bytes() []byte
		SetBytes(b []byte) Engine
	}

	P2PKH interface {
		P2PKHHashBytes(hash []byte) []byte
		P2PKHHash(hash []byte) Engine
		P2PKHPublicBytes(pub []byte) []byte
		P2PKHPublic(pub []byte) Engine
	}

	P2PK interface {
		P2PKPublicBytes(pub []byte) []byte
		P2PKPublic(pub []byte) Engine
	}

	P2SH interface {
		P2SHHashBytes(h []byte) []byte
		P2SHHash(h []byte) Engine
		P2SHScriptBytes(s []byte) []byte
		P2SHScript(s []byte) Engine
	}

	P2MS interface {
		P2MSBytes(verify bool, nRequired int64, pubKeys ...[]byte) []byte
		P2MS(verify bool, nRequired int64, pubKeys ...[]byte) Engine
	}

	LockTime interface {
		LockTimeBytes(lock int64) []byte
		LockTime(lock int64) Engine
		LockTimeTimeBytes(t time.Time) []byte
		LockTimeTime(t time.Time) Engine
	}

	Sequence interface {
		SequenceBytes(lock int64) []byte
		Sequence(lock int64) Engine
	}

	HashLock interface {
		HashLockBytes(h []byte, verify bool) []byte
		HashLock(h []byte, verify bool) Engine
	}

	HTLC interface {
		HTLCBytes(lockScript, tokenHash, timeLockedScript, hashLockedScript []byte) []byte
		HTLC(lockScript, tokenHash, timeLockedScript, hashLockedScript []byte) Engine
	}

	MSTLC interface {
		MSTLCBytes(lockScript, timeLockedScript []byte, nRequired int64, pubKeys ...[]byte) []byte
		MSTLC(lockScript, timeLockedScript []byte, nRequired int64, pubKeys ...[]byte) Engine
	}

	Engine interface {
		Disassembler
		Validator
		Logical
		Data
		Bytes
		P2PK
		P2PKH
		P2SH
		P2MS
		LockTime
		Sequence
		HashLock
		HTLC
		MSTLC
	}
)
