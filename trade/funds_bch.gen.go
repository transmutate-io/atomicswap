package trade

import (
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/networks"
	"transmutate.io/pkg/atomicswap/params"
	"transmutate.io/pkg/cryptocore/types"
)

type fundsDataBCH struct{ *fundsDataBTC }

func newFundsDataBCH() FundsData {
	return &fundsDataBCH{
		fundsDataBTC: newFundsDataBTC().(*fundsDataBTC),
	}
}

func (fd *fundsDataBCH) MarshalYAML() (interface{}, error) {
	return fd.fundsDataBTC, nil
}

func (fd *fundsDataBCH) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := &fundsDataBTC{}
	if err := unmarshal(r); err != nil {
		return err
	}
	fd.fundsDataBTC = r
	return nil
}

func (fd *fundsDataBCH) Lock() Lock { return &fundsLockBCH{fundsLockBTC(fd.LockScript)} }

type fundsLockBCH struct{ fundsLockBTC }

func newFundsLockBCH(l types.Bytes) Lock {
	return &fundsLockBCH{
		fundsLockBTC: newFundsLockBTC(l).(fundsLockBTC),
	}
}

func (fl *fundsLockBCH) LockData() (*LockData, error) {
	return parseLockScript(cryptos.BitcoinCash, fl.fundsLockBTC)
}

func (fl *fundsLockBCH) Address(chain params.Chain) (string, error) {
	return networks.All[cryptos.BitcoinCash][chain].P2SHFromScript(fl.fundsLockBTC)
}

func (fl *fundsLockBCH) MarshalYAML() (interface{}, error) { return fl.fundsLockBTC, nil }

func (fl *fundsLockBCH) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := fundsLockBTC{}
	if err := unmarshal(&r); err != nil {
		return err
	}
	fl.fundsLockBTC = r
	return nil
}
