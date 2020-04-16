package cryptos

import "transmutate.io/pkg/atomicswap/types/key"

func newCryptoBCH() Crypto {
	return &crypto{
		name:       "bitcoin-cash",
		short:      "BCH",
		newPrivKey: key.NewPrivateBCH,
	}
}
