package atomicswap

import (
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/cryptotypes"
)

type Funds interface{ CryptoType() cryptotypes.CryptoType }

func newFunds(c *cryptos.Crypto) Funds {
	switch c.Type {
	case cryptotypes.UTXO:
		return newFundsUTXO()
	default:
		panic("not supported")
	}
}
