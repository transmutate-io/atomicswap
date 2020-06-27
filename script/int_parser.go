package script

import (
	"transmutate.io/pkg/atomicswap/cryptos"
)

type IntParser interface{ ParseInt64(v []byte) (int64, error) }

func ParseInt64(c *cryptos.Crypto, v []byte) (int64, error) {
	p, ok := intParsers[c.Name]
	if !ok {
		return 0, cryptos.InvalidCryptoError(c.Name)
	}
	return p.ParseInt64(v)
}
