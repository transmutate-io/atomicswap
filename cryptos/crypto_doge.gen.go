package cryptos

import (
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/transaction"
)

func newCryptoDOGE() Crypto {
	return &crypto{
		name:       "dogecoin",
		short:      "DOGE",
		newPrivKey: key.NewPrivateDOGE,
		newTx:      transaction.NewDOGE,
	}
}
