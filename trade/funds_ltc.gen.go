package trade

import (
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/networks"
	"github.com/transmutate-io/atomicswap/params"
	"github.com/transmutate-io/cryptocore/types"
)

type fundsDataLTC struct{ *fundsDataBTC }

func newFundsDataLTC() FundsData {
	return &fundsDataLTC{
		fundsDataBTC: newFundsDataBTC().(*fundsDataBTC),
	}
}

func (fd *fundsDataLTC) MarshalYAML() (interface{}, error) {
	return fd.fundsDataBTC, nil
}

func (fd *fundsDataLTC) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := &fundsDataBTC{}
	if err := unmarshal(r); err != nil {
		return err
	}
	fd.fundsDataBTC = r
	return nil
}

func (fd *fundsDataLTC) Lock() Lock { return &fundsLockLTC{fundsLockBTC(fd.LockScript)} }

type fundsLockLTC struct{ fundsLockBTC }

func newFundsLockLTC(l types.Bytes) Lock {
	return &fundsLockLTC{
		fundsLockBTC: newFundsLockBTC(l).(fundsLockBTC),
	}
}

func (fl *fundsLockLTC) LockData() (*LockData, error) {
	return parseLockScript(cryptos.Litecoin, fl.fundsLockBTC)
}

func (fl *fundsLockLTC) Address(chain params.Chain) (string, error) {
	return networks.All[cryptos.Litecoin][chain].P2SHFromScript(fl.fundsLockBTC)
}

func (fl *fundsLockLTC) MarshalYAML() (interface{}, error) { return fl.fundsLockBTC, nil }

func (fl *fundsLockLTC) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := fundsLockBTC{}
	if err := unmarshal(&r); err != nil {
		return err
	}
	fl.fundsLockBTC = r
	return nil
}
