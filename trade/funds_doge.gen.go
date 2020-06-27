package trade

import (
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/networks"
	"transmutate.io/pkg/atomicswap/params"
	"transmutate.io/pkg/cryptocore/types"
)

type fundsDataDOGE struct{ *fundsDataBTC }

func newFundsDataDOGE() FundsData {
	return &fundsDataDOGE{
		fundsDataBTC: newFundsDataBTC().(*fundsDataBTC),
	}
}

func (fd *fundsDataDOGE) MarshalYAML() (interface{}, error) {
	return fd.fundsDataBTC, nil
}

func (fd *fundsDataDOGE) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := &fundsDataBTC{}
	if err := unmarshal(r); err != nil {
		return err
	}
	fd.fundsDataBTC = r
	return nil
}

func (fd *fundsDataDOGE) Lock() Lock { return &fundsLockDOGE{fundsLockBTC(fd.LockScript)} }

type fundsLockDOGE struct{ fundsLockBTC }

func newFundsLockDOGE(l types.Bytes) Lock {
	return &fundsLockDOGE{
		fundsLockBTC: newFundsLockBTC(l).(fundsLockBTC),
	}
}

func (fl *fundsLockDOGE) LockData() (*LockData, error) {
	return parseLockScript(cryptos.Dogecoin, fl.fundsLockBTC)
}

func (fl *fundsLockDOGE) Address(chain params.Chain) (string, error) {
	return networks.All[cryptos.Dogecoin][chain].P2SHFromScript(fl.fundsLockBTC)
}

func (fl *fundsLockDOGE) MarshalYAML() (interface{}, error) { return fl.fundsLockBTC, nil }

func (fl *fundsLockDOGE) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := fundsLockBTC{}
	if err := unmarshal(&r); err != nil {
		return err
	}
	fl.fundsLockBTC = r
	return nil
}
