package cryptos

import (
	"transmutate.io/pkg/atomicswap/types/transaction"
	"transmutate.io/pkg/atomicswap/types/key"
)

func newCryptoDOGE() Crypto {
	return &crypto{
		name:       "dogecoin",
		short:      "DOGE",
		newPrivKey: key.NewPrivateDOGE,
		newTx:      transaction.NewDOGE,
	}
}
