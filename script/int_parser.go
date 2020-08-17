package script

import (
	"github.com/transmutate-io/atomicswap/cryptos"
)

// IntParser represent and int64 parser
type IntParser interface{ ParseInt64(v []byte) (int64, error) }

// ParseInt64 parses the int64 for the given crypto
func ParseInt64(c *cryptos.Crypto, v []byte) (int64, error) {
	p, ok := intParsers[c.Name]
	if !ok {
		return 0, cryptos.InvalidCryptoError(c.Name)
	}
	return p.ParseInt64(v)
}
