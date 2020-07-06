package script

import (
	"time"

	"github.com/transmutate-io/atomicswap/cryptos"
)

type (
	Generator interface {
		If(i, e []byte) []byte
		Data(b []byte) []byte
		Int64(n int64) []byte
		P2PKHHash(hash []byte) []byte
		P2PKHPublic(pub []byte) []byte
		P2PKPublic(pub []byte) []byte
		P2SHHash(h []byte) []byte
		P2SHScript(s []byte) []byte
		P2SHRedeem(s []byte, pref ...[]byte) []byte
		P2MS(verify bool, nRequired int64, pubKeys ...[]byte) []byte
		LockTime(lock int64) []byte
		LockTimeTime(t time.Time) []byte
		Sequence(lock int64) []byte
		HashLock(h []byte, verify bool) []byte
		HTLC(lockScript, tokenHash, timeLockedScript, hashLockedScript []byte) []byte
		HTLCRedeem(sig, key, token, locksScript []byte) []byte
		HTLCRecover(sig, key, locksScript []byte) []byte
		MSTLC(lockScript, timeLockedScript []byte, nRequired int64, pubKeys ...[]byte) []byte
	}
)

func NewGenerator(c *cryptos.Crypto) (Generator, error) {
	g, ok := generators[c.Name]
	if !ok {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return g, nil
}
