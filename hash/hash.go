package hash

import (
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
)

func Hash256(b []byte) []byte   { return Sha256Sum(Sha256Sum(b)) }
func Hash160(b []byte) []byte   { return Ripemd160Sum(Sha256Sum(b)) }
func Sha256Sum(b []byte) []byte { r := sha256.Sum256(b); return r[:] }

func Ripemd160Sum(b []byte) []byte {
	h := ripemd160.New()
	h.Write(b)
	return h.Sum(nil)
}
