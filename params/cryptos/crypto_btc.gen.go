package cryptos

import (
	"transmutate.io/pkg/atomicswap/types/transaction"
	"transmutate.io/pkg/atomicswap/types/key"
)

func newCryptoBTC() Crypto {
	return &crypto{
		name:       "bitcoin",
		short:      "BTC",
		newPrivKey: key.NewPrivateBTC,
		newTx:      transaction.NewBTC,
	}
}
