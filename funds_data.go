package atomicswap

import (
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/cryptotypes"
)

type Funds interface {
	Len() int
	Idx(idx int) interface{}
	Add(fd interface{})
	Data() interface{}
}

func newFundsData(c *cryptos.Crypto) Funds {
	switch c.Type {
	case cryptotypes.UTXO:
		return newFundsUTXO()
	default:
		panic("not supported")
	}
}
