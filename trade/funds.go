package trade

import "transmutate.io/pkg/atomicswap/cryptos"

func newFundsData(c *cryptos.Crypto) FundsData {
	switch c.Type {
	case cryptos.UTXO:
		return newFundsUTXO()
	default:
		panic("not supported")
	}
}
