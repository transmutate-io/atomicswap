package cryptos

import (
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/transaction"
)

func newCryptoBCH() Crypto {
	return &crypto{
		name:       "bitcoin-cash",
		short:      "BCH",
		newPrivKey: key.NewPrivateBCH,
		newTx:      transaction.NewBCH,
	}
}
