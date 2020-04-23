package cryptos

import "transmutate.io/pkg/atomicswap/cryptotypes"

func newCryptoDOGE() *Crypto {
	return &Crypto{
		Name:     "dogecoin",
		Short:    "DOGE",
		Decimals: 8,
		Type:     cryptotypes.UTXO,
	}
}
