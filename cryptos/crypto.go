package cryptos

import (
	"fmt"
	"strings"
)

var CryptosShort map[string]*Crypto

func init() {
	CryptosShort = make(map[string]*Crypto, len(Cryptos))
	for _, c := range Cryptos {
		CryptosShort[c.Short] = c
	}
}

type newCryptoFunc = func() *Crypto

type InvalidCryptoError string

func (e InvalidCryptoError) Error() string {
	return fmt.Sprintf("invalid crypto: \"%s\"", string(e))
}

func Parse(s string) (*Crypto, error) {
	r, ok := Cryptos[strings.ToLower(s)]
	if !ok {
		return nil, InvalidCryptoError(s)
	}
	return r, nil
}

func ParseShort(s string) (*Crypto, error) {
	r, ok := CryptosShort[strings.ToUpper(s)]
	if !ok {
		return nil, InvalidCryptoError(s)
	}
	return r, nil
}

type Crypto struct {
	Name     string
	Short    string
	Decimals int
	Type     Type
}

func (c Crypto) String() string { return c.Name }

func (c Crypto) MarshalYAML() (interface{}, error) { return c.Name, nil }

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
