package cryptos

import (
	"fmt"

	"transmutate.io/pkg/atomicswap/cryptotypes"
)

type newCryptoFunc = func() *Crypto

type InvalidCryptoError string

func (e InvalidCryptoError) Error() string {
	return fmt.Sprintf("invalid crypto: \"%s\"", string(e))
}

func ParseCrypto(s string) (*Crypto, error) {
	r, ok := Cryptos[s]
	if !ok {
		return nil, InvalidCryptoError(s)
	}
	return r(), nil
}

type Crypto struct {
	Name     string
	Short    string
	Decimals int
	Type     cryptotypes.CryptoType
}

func (c Crypto) String() string { return c.Name }

func (c Crypto) MarshalYAML() (interface{}, error) { return c.Name, nil }

func (c *Crypto) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	cr, err := ParseCrypto(r)
	if err != nil {
		return err
	}
	*c = *cr
	return nil
}
