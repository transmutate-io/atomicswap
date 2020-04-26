package cryptos

import "transmutate.io/pkg/atomicswap/cryptotypes"

var (
	Cryptos = map[string]*Crypto{
		"bitcoin-cash": &Crypto{
			Name:     "bitcoin-cash",
			Short:    "BCH",
			Decimals: 8,
			Type:     cryptotypes.UTXO,
		},
		"bitcoin": &Crypto{
			Name:     "bitcoin",
			Short:    "BTC",
			Decimals: 8,
			Type:     cryptotypes.UTXO,
		},
		"dogecoin": &Crypto{
			Name:     "dogecoin",
			Short:    "DOGE",
			Decimals: 8,
			Type:     cryptotypes.UTXO,
		},
		"litecoin": &Crypto{
			Name:     "litecoin",
			Short:    "LTC",
			Decimals: 8,
			Type:     cryptotypes.UTXO,
		},
	}
)
