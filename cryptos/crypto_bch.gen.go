package cryptos

import (
	"transmutate.io/pkg/atomicswap/cryptotypes"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/transaction"
)

func newCryptoBCH() *Crypto {
	return &Crypto{
		Name:       "bitcoin-cash",
		Short:      "BCH",
		newPrivKey: key.NewPrivateBCH,
		newTx:      transaction.NewBCH,
		Type:       cryptotypes.UTXO,
	}
}
