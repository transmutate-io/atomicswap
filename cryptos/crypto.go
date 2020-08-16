package cryptos

import (
	"fmt"
	"strings"
)

// CryptosShort contains all available cryptocurrencies
var CryptosShort map[string]*Crypto

func init() {
	CryptosShort = make(map[string]*Crypto, len(Cryptos))
	for _, c := range Cryptos {
		CryptosShort[c.Short] = c
	}
}

type newCryptoFunc = func() *Crypto

// InvalidCryptoError represents an error parsing a crypto name
type InvalidCryptoError string

// Error implement error
func (e InvalidCryptoError) Error() string {
	return fmt.Sprintf("invalid crypto: \"%s\"", string(e))
}

// Parse parses a cryptocurrency name
func Parse(s string) (*Crypto, error) {
	r, ok := Cryptos[strings.ToLower(s)]
	if !ok {
		return nil, InvalidCryptoError(s)
	}
	return r, nil
}

// ParseShort parses a cryptocurrency ticker
func ParseShort(s string) (*Crypto, error) {
	r, ok := CryptosShort[strings.ToUpper(s)]
	if !ok {
		return nil, InvalidCryptoError(s)
	}
	return r, nil
}

// Crypto represents a cryptocurrency
type Crypto struct {
	// Name of the crypto
	Name string
	// Short is the ticker
	Short string
	// Decimals used
	Decimals int
	// Type of crypto
	Type Type
}

// String implement fmt.Stringer
func (c Crypto) String() string { return c.Name }

// MarshalYAML implement yaml.Marshaler
func (c Crypto) MarshalYAML() (interface{}, error) { return c.Name, nil }

// UnmarshalYAML implement yaml.Unmarshaler
func (c *Crypto) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	cr, err := Parse(r)
	if err != nil {
		return err
	}
	*c = *cr
	return nil
}
