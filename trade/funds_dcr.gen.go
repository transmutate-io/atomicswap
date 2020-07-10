package trade

import (
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/networks"
	"github.com/transmutate-io/atomicswap/params"
	"github.com/transmutate-io/cryptocore/types"
)

type fundsDataDCR struct{ *fundsDataBTC }

func newFundsDataDCR() FundsData {
	return &fundsDataDCR{
		fundsDataBTC: newFundsDataBTC().(*fundsDataBTC),
	}
}

func (fd *fundsDataDCR) MarshalYAML() (interface{}, error) {
	return fd.fundsDataBTC, nil
}

func (fd *fundsDataDCR) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := &fundsDataBTC{}
	if err := unmarshal(r); err != nil {
		return err
	}
	fd.fundsDataBTC = r
	return nil
}

func (fd *fundsDataDCR) Lock() Lock { return &fundsLockDCR{fundsLockBTC(fd.LockScript)} }

type fundsLockDCR struct{ fundsLockBTC }

func newFundsLockDCR(l types.Bytes) Lock {
	return &fundsLockDCR{
		fundsLockBTC: newFundsLockBTC(l).(fundsLockBTC),
	}
}

func (fl *fundsLockDCR) LockData() (*LockData, error) {
	return parseLockScript(cryptos.Decred, fl.fundsLockBTC)
}

func (fl *fundsLockDCR) Address(chain params.Chain) (string, error) {
	return networks.All[cryptos.Decred][chain].P2SHFromScript(fl.fundsLockBTC)
}

func (fl *fundsLockDCR) MarshalYAML() (interface{}, error) { return fl.fundsLockBTC.Bytes().Hex(), nil }

func (fl *fundsLockDCR) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := fundsLockBTC{}
	if err := unmarshal(&r); err != nil {
		return err
	}
	fl.fundsLockBTC = r
	return nil
}
