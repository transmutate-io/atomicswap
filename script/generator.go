package script

import (
	"time"

	"github.com/transmutate-io/atomicswap/cryptos"
)

type (
	// Generator represents a script generator
	Generator interface {
		// If else statement. If e is nil an else branch will not be present
		If(i, e []byte) []byte
		// Data returns the bytes as data
		Data(b []byte) []byte
		// Int64 returns n as data
		Int64(n int64) []byte
		// P2PKHHash returns a p2pkh contract using the hash
		P2PKHHash(hash []byte) []byte
		// P2PKHPublic returns a p2pkh contract using the public key
		P2PKHPublic(pub []byte) []byte
		// P2PKPublic returns a p2pk contract using the public key
		P2PKPublic(pub []byte) []byte
		// P2SHHash returns a p2sh contract using the hash
		P2SHHash(h []byte) []byte
		// P2SHScript returns a p2sh contract using the script
		P2SHScript(s []byte) []byte
		// P2SHRedeem returns a script to redeem a p2sh contract
		P2SHRedeem(s []byte, pref ...[]byte) []byte
		// P2MS returns a p2ms contract
		P2MS(verify bool, nRequired int64, pubKeys ...[]byte) []byte
		// LockTime returns an absolute timelock using an int
		LockTime(lock int64) []byte
		// LockTimeTime returns an absolute timelock using an time.Time
		LockTimeTime(t time.Time) []byte
		// Sequence returns a relative timelock using an int
		Sequence(lock int64) []byte
		// HashLock returns an hashlock
		HashLock(h []byte, verify bool) []byte
		// HTLC returns returns an hash time locked contract
		HTLC(lockScript, tokenHash, timeLockedScript, hashLockedScript []byte) []byte
		// HTLCRedeem returns the script to redeem an htlc
		HTLCRedeem(sig, key, token, locksScript []byte) []byte
		// HTLCRecover returns the script to recover an htlc
		HTLCRecover(sig, key, locksScript []byte) []byte
		// MSTLC returns a multi-sig time locked contract
		MSTLC(lockScript, timeLockedScript []byte, nRequired int64, pubKeys ...[]byte) []byte
	}
)

// NewGenerator returns a generator for the given crypto
func NewGenerator(c *cryptos.Crypto) (Generator, error) {
	g, ok := generators[c.Name]
	if !ok {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return g, nil
}
