package cryptos

import (
	"transmutate.io/pkg/atomicswap/cryptotypes"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/transaction"
)

func newCryptoBTC() *Crypto {
	return &Crypto{
		Name:       "bitcoin",
		Short:      "BTC",
		newPrivKey: key.NewPrivateBTC,
		newTx:      transaction.NewBTC,
		Type:       cryptotypes.UTXO,
	}
}
