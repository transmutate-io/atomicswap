package cryptos

import "transmutate.io/pkg/atomicswap/cryptotypes"

func newCryptoBTC() *Crypto {
	return &Crypto{
		Name:     "bitcoin",
		Short:    "BTC",
		Decimals: 8,
		Type:     cryptotypes.UTXO,
	}
}
