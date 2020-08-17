package trade

import (
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/networks"
	"github.com/transmutate-io/atomicswap/params"
	"github.com/transmutate-io/cryptocore/types"
)

type fundsDataDOGE struct{ *fundsDataBTC }

func newFundsDataDOGE() FundsData {
	return &fundsDataDOGE{
		fundsDataBTC: newFundsDataBTC().(*fundsDataBTC),
	}
}

// MarshalYAML implement yaml.Marshaler
func (fd *fundsDataDOGE) MarshalYAML() (interface{}, error) {
	return fd.fundsDataBTC, nil
}

// UnmarshalYAML implement yaml.Unmarshaler
func (fd *fundsDataDOGE) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := &fundsDataBTC{}
	if err := unmarshal(r); err != nil {
		return err
	}
	fd.fundsDataBTC = r
	return nil
}

// Lock implement FundsData
func (fd *fundsDataDOGE) Lock() Lock { return &fundsLockDOGE{fundsLockBTC(fd.LockScript)} }

type fundsLockDOGE struct{ fundsLockBTC }

func newFundsLockDOGE(l types.Bytes) Lock {
	return &fundsLockDOGE{
		fundsLockBTC: newFundsLockBTC(l).(fundsLockBTC),
	}
}

// LockData implement Lock
func (fl *fundsLockDOGE) LockData() (*LockData, error) {
	return parseLockScript(cryptos.Dogecoin, fl.fundsLockBTC)
}

// Address implement Lock
func (fl *fundsLockDOGE) Address(chain params.Chain) (string, error) {
	return networks.All[cryptos.Dogecoin][chain].P2SHFromScript(fl.fundsLockBTC)
}

// MarshalYAML implement yaml.Marshaler
func (fl *fundsLockDOGE) MarshalYAML() (interface{}, error) {
	return fl.fundsLockBTC.Bytes().Hex(), nil
}

// UnmarshalYAML implement yaml.Unmarshaler
func (fl *fundsLockDOGE) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := fundsLockBTC{}
	if err := unmarshal(&r); err != nil {
		return err
	}
	fl.fundsLockBTC = r
	return nil
}
