package key

import (
	"fmt"

	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/yamltypes"
)

type (
	NewPrivateFunc = func() (Private, error)
	NewPublicFunc  = func(b []byte) (Public, error)

	Keyer interface{ Key() interface{} }

	Private interface {
		Public() Public
		Sign(b []byte) ([]byte, error)
		yamltypes.MarshalerUnmarshaler
		Serialize() []byte
		Keyer
	}

	Public interface {
		Hash160() []byte
		Verify(sig, msg []byte) error
		yamltypes.MarshalerUnmarshaler
		SerializeCompressed() []byte
		Keyer
	}
)

type newFuncs struct {
	priv NewPrivateFunc
	pub  NewPublicFunc
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

func NewPrivate(c *cryptos.Crypto) (Private, error) {
	cf, ok := cryptoFuncs[c.String()]
	if !ok {
		return nil, (*KeysError)(c)
	}
	return cf.priv()
}
