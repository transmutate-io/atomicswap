package cryptos

import (
	"fmt"

	"transmutate.io/pkg/atomicswap/types/transaction"

	"transmutate.io/pkg/atomicswap/types/key"
)

type (
	ParseCryptoFunc          = func(string) (Crypto, error)
	YAMLMarshalerUnmarshaler interface {
		MarshalYAML() (interface{}, error)
		UnmarshalYAML(unmarshal func(interface{}) error) error
	}

	Crypto interface {
		String() string
		Name() string
		Short() string
		NewPrivateKey() (key.Private, error)
		NewTx() transaction.Tx
		YAMLMarshalerUnmarshaler
	}

	newCryptoFunc = func() Crypto
)

type InvalidCryptoError string

func (e InvalidCryptoError) Error() string {
	return fmt.Sprintf("invalid crypto: \"%s\"", string(e))
}

func ParseCrypto(s string) (Crypto, error) {
	r, ok := _cryptos[s]
	if !ok {
		return nil, InvalidCryptoError(s)
	}
	return r(), nil
}

type crypto struct {
	name       string
	short      string
	newPrivKey func() (key.Private, error)
	newTx      func() transaction.Tx
}

func (c crypto) Short() string { return c.short }

func (c crypto) Name() string { return c.name }

func (c crypto) String() string { return c.name }

func (c crypto) MarshalYAML() (interface{}, error) { return c.name, nil }

func (c crypto) NewPrivateKey() (key.Private, error) { return c.newPrivKey() }

func (c crypto) NewTx() transaction.Tx { return c.newTx() }

func (c *crypto) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	c.name = r
	return nil
}
