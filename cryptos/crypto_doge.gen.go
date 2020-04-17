package cryptos

import (
	"transmutate.io/pkg/atomicswap/cryptotypes"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/transaction"
)

func newCryptoDOGE() *Crypto {
	return &Crypto{
		Name:       "dogecoin",
		Short:      "DOGE",
		newPrivKey: key.NewPrivateDOGE,
		newTx:      transaction.NewDOGE,
		Type:       cryptotypes.UTXO,
	}
}
