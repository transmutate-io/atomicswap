package script

import (
	"time"

	"transmutate.io/pkg/atomicswap/cryptos"
)

type Engine struct {
	b         []byte
	Generator Generator
}

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

func (eng *Engine) If(i, e []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.If(i, e)...)
	return eng
}

func (eng *Engine) Data(b []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.Data(b)...)
	return eng
}

func (eng *Engine) Int64(n int64) *Engine {
	eng.b = append(eng.b, eng.Generator.Int64(n)...)
	return eng
}

func (eng *Engine) P2PKHHash(hash []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.P2PKHHash(hash)...)
	return eng
}

func (eng *Engine) P2PKHPublic(pub []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.P2PKHPublic(pub)...)
	return eng
}

func (eng *Engine) P2PKPublic(pub []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.P2PKPublic(pub)...)
	return eng
}

func (eng *Engine) P2SHHash(h []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.P2SHHash(h)...)
	return eng
}

func (eng *Engine) P2SHScript(s []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.P2SHScript(s)...)
	return eng
}

func (eng *Engine) P2MS(verify bool, nRequired int64, pubKeys ...[]byte) *Engine {
	eng.b = append(eng.b, eng.Generator.P2MS(verify, nRequired, pubKeys...)...)
	return eng
}

func (eng *Engine) LockTime(lock int64) *Engine {
	eng.b = append(eng.b, eng.Generator.LockTime(lock)...)
	return eng
}

func (eng *Engine) LockTimeTime(t time.Time) *Engine {
	eng.b = append(eng.b, eng.Generator.LockTimeTime(t)...)
	return eng
}

func (eng *Engine) Sequence(lock int64) *Engine {
	eng.b = append(eng.b, eng.Generator.Sequence(lock)...)
	return eng
}

func (eng *Engine) HashLock(h []byte, verify bool) *Engine {
	eng.b = append(eng.b, eng.Generator.HashLock(h, verify)...)
	return eng
}

func (eng *Engine) HTLC(lockScript, tokenHash, timeLockedScript, hashLockedScript []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.HTLC(lockScript, tokenHash, timeLockedScript, hashLockedScript)...)
	return eng
}

func (eng *Engine) HTLCRedeem(sig, key, token, locksScript []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.HTLCRedeem(sig, key, token, locksScript)...)
	return eng
}

func (eng *Engine) HTLCRecover(sig, key, locksScript []byte) *Engine {
	eng.b = append(eng.b, eng.Generator.HTLCRecover(sig, key, locksScript)...)
	return eng
}

func (eng *Engine) MSTLC(lockScript, timeLockedScript []byte, nRequired int64, pubKeys ...[]byte) *Engine {
	eng.b = append(eng.b, eng.Generator.MSTLC(lockScript, timeLockedScript, nRequired, pubKeys...)...)
	return eng
}

func (eng *Engine) Bytes() []byte { return eng.b }

func (eng *Engine) SetBytes(b []byte) { eng.b = b }
