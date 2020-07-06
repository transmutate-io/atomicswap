package hash

import (
	"crypto/sha256"

	"github.com/decred/dcrd/crypto/blake256"
	"github.com/transmutate-io/atomicswap/cryptos"
	"golang.org/x/crypto/ripemd160"
)

type Hasher interface {
	Hash256([]byte) []byte
	Hash160([]byte) []byte
}

func New(c *cryptos.Crypto) (Hasher, error) {
	nf, ok := newHasherFuncs[c.Name]
	if !ok {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return nf(), nil
}

// Sha256Sum returns the sha256 for b
func Sha256Sum(b []byte) []byte { r := sha256.Sum256(b); return r[:] }

// Ripemd160Sum returns the ripemd160 for b
func Ripemd160Sum(b []byte) []byte {
	h := ripemd160.New()
	h.Write(b)
	return h.Sum(nil)
}

// Blake256Sum returns the blake sha256 for b
func Blake256Sum(b []byte) []byte {
	hash := blake256.Sum256(b)
	return hash[:]
}
