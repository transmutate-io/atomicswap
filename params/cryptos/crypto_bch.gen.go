package cryptos

import (
	"transmutate.io/pkg/atomicswap/types/key"
	"transmutate.io/pkg/atomicswap/types/transaction"
)

func newCryptoBCH() Crypto {
	return &crypto{
		name:       "bitcoin-cash",
		short:      "BCH",
		newPrivKey: key.NewPrivateBCH,
		newTx:      transaction.NewBCH,
	}
}
