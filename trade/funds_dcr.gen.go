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

// MarshalYAML implement yaml.Marshaler
func (fd *fundsDataDCR) MarshalYAML() (interface{}, error) {
	return fd.fundsDataBTC, nil
}

// UnmarshalYAML implement yaml.Unmarshaler
func (fd *fundsDataDCR) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := &fundsDataBTC{}
	if err := unmarshal(r); err != nil {
		return err
	}
	fd.fundsDataBTC = r
	return nil
}

// Lock implement FundsData
func (fd *fundsDataDCR) Lock() Lock { return &fundsLockDCR{fundsLockBTC(fd.LockScript)} }

type fundsLockDCR struct{ fundsLockBTC }

func newFundsLockDCR(l types.Bytes) Lock {
	return &fundsLockDCR{
		fundsLockBTC: newFundsLockBTC(l).(fundsLockBTC),
	}
}

// LockData implement Lock
func (fl *fundsLockDCR) LockData() (*LockData, error) {
	return parseLockScript(cryptos.Decred, fl.fundsLockBTC)
}

// Address implement Lock
func (fl *fundsLockDCR) Address(chain params.Chain) (string, error) {
	return networks.All[cryptos.Decred][chain].P2SHFromScript(fl.fundsLockBTC)
}

// MarshalYAML implement yaml.Marshaler
func (fl *fundsLockDCR) MarshalYAML() (interface{}, error) { return fl.fundsLockBTC.Bytes().Hex(), nil }

// UnmarshalYAML implement yaml.Unmarshaler
func (fl *fundsLockDCR) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := fundsLockBTC{}
	if err := unmarshal(&r); err != nil {
		return err
	}
	fl.fundsLockBTC = r
	return nil
}
