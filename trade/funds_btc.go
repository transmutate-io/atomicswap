package trade

import (
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/networks"
	"transmutate.io/pkg/atomicswap/params"
	"transmutate.io/pkg/cryptocore/types"
)

type fundsDataBTC struct {
	Outputs    []*Output   `yaml:"outputs"`
	LockScript types.Bytes `yaml:"lock_script"`
}

func newFundsDataBTC() FundsData { return &fundsDataBTC{Outputs: make([]*Output, 0, 4)} }

func (f *fundsDataBTC) CryptoType() cryptos.Type { return cryptos.UTXO }

func (f *fundsDataBTC) Funds() interface{} { return f.Outputs }

func (f *fundsDataBTC) AddFunds(funds interface{}) {
	f.Outputs = append(f.Outputs, funds.(*Output))
}

func (f fundsDataBTC) Lock() Lock { return fundsLockBTC(f.LockScript) }

func (f *fundsDataBTC) SetLock(lock Lock) { f.LockScript = lock.Bytes() }

type fundsLockBTC types.Bytes

func newFundsLockBTC(l types.Bytes) Lock { return fundsLockBTC(l) }

func (fl fundsLockBTC) Bytes() types.Bytes { return types.Bytes(fl) }

func (fl fundsLockBTC) Data() types.Bytes { return fl.Bytes() }

func (fl fundsLockBTC) LockData() (*LockData, error) {
	return parseLockScript(cryptos.Bitcoin, fl)
}

func (fl fundsLockBTC) Address(chain params.Chain) (string, error) {
	return networks.All[cryptos.Bitcoin][chain].P2SHFromScript(fl)
}
