package cryptos

var (
	BitcoinCash = &Crypto{
		Name:     "bitcoin-cash",
		Short:    "BCH",
		Decimals: 8,
		Type:     UTXO,
	}
	Bitcoin = &Crypto{
		Name:     "bitcoin",
		Short:    "BTC",
		Decimals: 8,
		Type:     UTXO,
	}
	Decred = &Crypto{
		Name:     "decred",
		Short:    "DCR",
		Decimals: 8,
		Type:     UTXO,
	}
	Dogecoin = &Crypto{
		Name:     "dogecoin",
		Short:    "DOGE",
		Decimals: 8,
		Type:     UTXO,
	}
	Litecoin = &Crypto{
		Name:     "litecoin",
		Short:    "LTC",
		Decimals: 8,
		Type:     UTXO,
	}

	Cryptos = map[string]*Crypto{
		"bitcoin-cash": BitcoinCash,
		"bitcoin":      Bitcoin,
		"decred":       Decred,
		"dogecoin":     Dogecoin,
		"litecoin":     Litecoin,
	}
)
