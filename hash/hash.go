package hash

import (
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
)

// Hash256 is a wrapper for the double hash sha256(sha256(b))
func Hash256(b []byte) []byte { return Sha256Sum(Sha256Sum(b)) }

// Hash160 is a wrapper for the double hash ripemd160(sha256(b))
func Hash160(b []byte) []byte { return Ripemd160Sum(Sha256Sum(b)) }

// Sha256sum returns the sha256 for b
func Sha256Sum(b []byte) []byte { r := sha256.Sum256(b); return r[:] }

// Ripemd160Sum returns the ripemd160 for b
func Ripemd160Sum(b []byte) []byte {
	h := ripemd160.New()
	h.Write(b)
	return h.Sum(nil)
}
