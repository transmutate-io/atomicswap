package cryptos

import (
	"transmutate.io/pkg/atomicswap/cryptotypes"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/transaction"
)

func newCryptoLTC() *Crypto {
	return &Crypto{
		Name:       "litecoin",
		Short:      "LTC",
		newPrivKey: key.NewPrivateLTC,
		newTx:      transaction.NewLTC,
		Type:       cryptotypes.UTXO,
	}
}
