package cryptos

import (
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/transaction"
)

func newCryptoLTC() Crypto {
	return &crypto{
		name:       "litecoin",
		short:      "LTC",
		newPrivKey: key.NewPrivateLTC,
		newTx:      transaction.NewLTC,
	}
}
