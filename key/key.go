package key

import (
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/cryptocore/types"
)

type (
	NewPrivateFunc   = func() (Private, error)
	NewPublicFunc    = func(b []byte) (Public, error)
	ParsePrivateFunc = func(b []byte) (Private, error)

	Keyer interface{ Key() interface{} }

	KeyData = types.Bytes

	Private interface {
		Public() Public
		Sign(b []byte) ([]byte, error)
		Serialize() []byte
		Keyer
	}

	Public interface {
		KeyData() KeyData
		Verify(sig, msg []byte) error
		SerializeCompressed() []byte
		Keyer
	}
)

type newFuncs struct {
	parsePriv ParsePrivateFunc
	priv      NewPrivateFunc
	pub       NewPublicFunc
}

func getCryptoFuncs(c *cryptos.Crypto) (*newFuncs, error) {
	cf, ok := cryptoFuncs[c.String()]
	if !ok {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return &cf, nil
}

func ParsePrivate(c *cryptos.Crypto, b []byte) (Private, error) {
	cf, err := getCryptoFuncs(c)
	if err != nil {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return cf.parsePriv(b)
}

func NewPrivate(c *cryptos.Crypto) (Private, error) {
	cf, err := getCryptoFuncs(c)
	if err != nil {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return cf.priv()
}

func NewPublic(c *cryptos.Crypto, b []byte) (Public, error) {
	cf, err := getCryptoFuncs(c)
	if err != nil {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return cf.pub(b)
}
