package script

import (
	"time"

	"github.com/transmutate-io/atomicswap/cryptos"
)

// Engine represents a scripting engine
type Engine struct {
	b         []byte
	Generator Generator
}

// NewEngine returns a scripting engine for the given crypto
func NewEngine(c *cryptos.Crypto) (*Engine, error) {
	gen, ok := generators[c.Name]
	if !ok {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return newEngine(gen), nil
}

func newEngine(gen Generator) *Engine {
	return &Engine{
		b:         make([]byte, 0, 1024),
		Generator: gen,
	}
}

// If adds an if-else to the script
func (eng *Engine) If(i, e []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.If(i, e)...)
	return eng
}

// Data adds b as data the script
func (eng *Engine) Data(b []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.Data(b)...)
	return eng
}

// Int64 adds n as data to the script
func (eng *Engine) Int64(n int64) *Engine {
	eng.b = append(eng.b, eng.Generator.Int64(n)...)
	return eng
}

// P2PKHHash adds a p2pkh contract to the script using the key hash
func (eng *Engine) P2PKHHash(hash []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.P2PKHHash(hash)...)
	return eng
}

// P2PKHPublic adds a p2pkh contract to the script using the public key
func (eng *Engine) P2PKHPublic(pub []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.P2PKHPublic(pub)...)
	return eng
}

// P2PKPublic adds a p2pk contract to the script using the public key
func (eng *Engine) P2PKPublic(pub []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.P2PKPublic(pub)...)
	return eng
}

// P2SHHash adds a p2sh contract to the script using the script hash
func (eng *Engine) P2SHHash(h []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.P2SHHash(h)...)
	return eng
}

// P2SHScript adds a p2sh contract to the script using the script s
func (eng *Engine) P2SHScript(s []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.P2SHScript(s)...)
	return eng
}

// P2MS adds a p2ms contract to the script
func (eng *Engine) P2MS(verify bool, nRequired int64, pubKeys ...[]byte) *Engine {
	eng.b = append(eng.b, eng.Generator.P2MS(verify, nRequired, pubKeys...)...)
	return eng
}

// LockTime adds a timelock to the script using a fixed int
func (eng *Engine) LockTime(lock int64) *Engine {
	eng.b = append(eng.b, eng.Generator.LockTime(lock)...)
	return eng
}

// LockTimeTime adds a timelock to the script using an absolute time
func (eng *Engine) LockTimeTime(t time.Time) *Engine {
	eng.b = append(eng.b, eng.Generator.LockTimeTime(t)...)
	return eng
}

// Sequence adds a sequence check to the script
func (eng *Engine) Sequence(lock int64) *Engine {
	eng.b = append(eng.b, eng.Generator.Sequence(lock)...)
	return eng
}

// HashLock adds an hashlock to the script
func (eng *Engine) HashLock(h []byte, verify bool) *Engine {
	eng.b = append(eng.b, eng.Generator.HashLock(h, verify)...)
	return eng
}

// HTLC adds an hash-time-locked contract to the script
func (eng *Engine) HTLC(lockScript, tokenHash, timeLockedScript, hashLockedScript []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.HTLC(lockScript, tokenHash, timeLockedScript, hashLockedScript)...)
	return eng
}

// HTLCRedeem redeems an htlc to the script
func (eng *Engine) HTLCRedeem(sig, key, token, locksScript []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.HTLCRedeem(sig, key, token, locksScript)...)
	return eng
}

// HTLCRecover recovers an htlc to the script
func (eng *Engine) HTLCRecover(sig, key, locksScript []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.HTLCRecover(sig, key, locksScript)...)
	return eng
}

// MSTLC adds a multisig-time-lock contract to the script
func (eng *Engine) MSTLC(lockScript, timeLockedScript []byte, nRequired int64, pubKeys ...[]byte) *Engine {
	eng.b = append(eng.b, eng.Generator.MSTLC(lockScript, timeLockedScript, nRequired, pubKeys...)...)
	return eng
}

// Bytes returns the script bytes
func (eng *Engine) Bytes() []byte { return eng.b }

// SetBytes sets the script bytes
func (eng *Engine) SetBytes(b []byte) { eng.b = b }
