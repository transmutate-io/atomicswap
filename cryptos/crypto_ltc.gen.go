package cryptos

import "transmutate.io/pkg/atomicswap/cryptotypes"

func newCryptoLTC() *Crypto {
	return &Crypto{
		Name:     "litecoin",
		Short:    "LTC",
		Decimals: 8,
		Type:     cryptotypes.UTXO,
	}
}
