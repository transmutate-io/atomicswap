package cryptos

import (
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/transaction"
)

func newCryptoBTC() Crypto {
	return &crypto{
		name:       "bitcoin",
		short:      "BTC",
		newPrivKey: key.NewPrivateBTC,
		newTx:      transaction.NewBTC,
	}
}
