package params

import (
	"fmt"
	"strings"
)

// UnknownCryptoError is returned when parsing an invalid crypto name
type UnknownCryptoError string

func (e UnknownCryptoError) Error() string { return fmt.Sprintf(`unknown crypto: "%s"`, string(e)) }

// UnknownTickerError is returned when parsing an invalid crypto ticker
type UnknownTickerError string

func (e UnknownTickerError) Error() string { return fmt.Sprintf(`unknown ticker: "%s"`, string(e)) }

// Crypto represents a cryptocurrency
type Crypto int

func (c Crypto) String() string { return _cryptoNames[c] }

// ParseCrypto parses a cryptocurrency name
func ParseCrypto(c string) (Crypto, error) {
	r, ok := _cryptosByName[c]
	if !ok {
		return 0, UnknownCryptoError(c)
	}
	return r, nil
}

// ParseTicker parses a cryptocurrency ticker
func ParseTicker(t string) (Crypto, error) {
	r, ok := _cryptosByTicker[strings.ToUpper(t)]
	if !ok {
		return 0, UnknownTickerError(t)
	}
	return r, nil
}
