package cryptos

import (
	"transmutate.io/pkg/atomicswap/types/transaction"
	"transmutate.io/pkg/atomicswap/types/key"
)

func newCryptoLTC() Crypto {
	return &crypto{
		name:       "litecoin",
		short:      "LTC",
		newPrivKey: key.NewPrivateLTC,
		newTx:      transaction.NewLTC,
	}
}
