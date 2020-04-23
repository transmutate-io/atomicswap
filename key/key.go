package key

import (
	"fmt"

	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/yamltypes"
)

type (
	NewPrivateFunc   = func() (Private, error)
	NewPublicFunc    = func(b []byte) (Public, error)
	ParsePrivateFunc = func(b []byte) (Private, error)

	Keyer interface{ Key() interface{} }

	KeyData interface{}

	Private interface {
		Public() Public
		Sign(b []byte) ([]byte, error)
		yamltypes.MarshalerUnmarshaler
		Serialize() []byte
		Keyer
	}

	Public interface {
		KeyData() KeyData
		Verify(sig, msg []byte) error
		yamltypes.MarshalerUnmarshaler
		SerializeCompressed() []byte
		Keyer
	}
)

type newFuncs struct {
	parsePriv ParsePrivateFunc
	priv      NewPrivateFunc
	pub       NewPublicFunc
}

var cryptoFuncs = map[string]newFuncs{
	"bitcoin": newFuncs{
		priv: NewPrivateBTC,
		pub:  NewPublicBTC,
	},
	"litecoin": newFuncs{
		priv: NewPrivateLTC,
		pub:  NewPublicLTC,
	},
	"dogecoin": newFuncs{
		priv: NewPrivateDOGE,
		pub:  NewPublicDOGE,
	},
	"bitcoin-cash": newFuncs{
		priv: NewPrivateBCH,
		pub:  NewPublicBCH,
	},
}

type KeysError cryptos.Crypto

func (e *KeysError) Error() string {
	return fmt.Sprintf(`can't create keys for crypto: "%s"`, (*cryptos.Crypto)(e).Name)
}

func getCryptoFuncs(c *cryptos.Crypto) (*newFuncs, error) {
	cf, ok := cryptoFuncs[c.String()]
	if !ok {
		return nil, (*KeysError)(c)
	}
	return &cf, nil
}

func ParsePrivate(c *cryptos.Crypto, b []byte) (Private, error) {
	cf, err := getCryptoFuncs(c)
	if err != nil {
		return nil, (*KeysError)(c)
	}
	return cf.parsePriv(b)
}

func NewPrivate(c *cryptos.Crypto) (Private, error) {
	cf, err := getCryptoFuncs(c)
	if err != nil {
		return nil, (*KeysError)(c)
	}
	return cf.priv()
}

func NewPublic(c *cryptos.Crypto, b []byte) (Public, error) {
	cf, err := getCryptoFuncs(c)
	if err != nil {
		return nil, (*KeysError)(c)
	}
	return cf.pub(b)
}
