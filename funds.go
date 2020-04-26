package atomicswap

import (
	"transmutate.io/pkg/atomicswap/cryptos"
)

type Funds interface {
	CryptoType() cryptos.Type
	AddFunds(funds interface{})
	Funds() interface{}
	SetLock(lock Lock)
	Lock() Lock
}

func newFunds(c *cryptos.Crypto) Funds {
	switch c.Type {
	case cryptos.UTXO:
		return newFundsUTXO()
	default:
		panic("not supported")
	}
}

type (
	// LockData struct {
	// 	Locktime        time.Time
	// 	RedeemKeyData   key.KeyData
	// 	RecoveryKeyData key.KeyData
	// }
	Lock interface {
		Data() interface{}
		// LockData() *LockData
	}
)
