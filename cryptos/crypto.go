package cryptos

import (
	"fmt"

	"transmutate.io/pkg/atomicswap/cryptotypes"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/transaction"
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
	Name       string
	Short      string
	newPrivKey func() (key.Private, error)
	newTx      func() transaction.Tx
	Type       cryptotypes.CryptoType
}

func (c Crypto) String() string { return c.Name }

func (c Crypto) NewPrivateKey() (key.Private, error) { return c.newPrivKey() }

func (c Crypto) NewTx() transaction.Tx { return c.newTx() }

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
