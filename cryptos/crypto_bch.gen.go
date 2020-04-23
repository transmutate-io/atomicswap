package cryptos

import "transmutate.io/pkg/atomicswap/cryptotypes"

func newCryptoBCH() *Crypto {
	return &Crypto{
		Name:     "bitcoin-cash",
		Short:    "BCH",
		Decimals: 8,
		Type:     cryptotypes.UTXO,
	}
}
