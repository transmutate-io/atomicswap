package trade

import "transmutate.io/pkg/cryptocore/types"

var (
	newFundsDataFuncs = map[string]func() FundsData{
		"bitcoin-cash": newFundsDataBCH,
		"bitcoin":      newFundsDataBTC,
		"decred":       newFundsDataDCR,
		"dogecoin":     newFundsDataDOGE,
		"litecoin":     newFundsDataLTC,
	}
	newFundsLockFuncs = map[string]func(types.Bytes) Lock{
		"bitcoin-cash": newFundsLockBCH,
		"bitcoin":      newFundsLockBTC,
		"decred":       newFundsLockDCR,
		"dogecoin":     newFundsLockDOGE,
		"litecoin":     newFundsLockLTC,
	}
)
