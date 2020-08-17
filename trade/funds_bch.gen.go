package trade

import (
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/networks"
	"github.com/transmutate-io/atomicswap/params"
	"github.com/transmutate-io/cryptocore/types"
)

type fundsDataBCH struct{ *fundsDataBTC }

func newFundsDataBCH() FundsData {
	return &fundsDataBCH{
		fundsDataBTC: newFundsDataBTC().(*fundsDataBTC),
	}
}

// MarshalYAML implement yaml.Marshaler
func (fd *fundsDataBCH) MarshalYAML() (interface{}, error) {
	return fd.fundsDataBTC, nil
}

// UnmarshalYAML implement yaml.Unmarshaler
func (fd *fundsDataBCH) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := &fundsDataBTC{}
	if err := unmarshal(r); err != nil {
		return err
	}
	fd.fundsDataBTC = r
	return nil
}

// Lock implement FundsData
func (fd *fundsDataBCH) Lock() Lock { return &fundsLockBCH{fundsLockBTC(fd.LockScript)} }

type fundsLockBCH struct{ fundsLockBTC }

func newFundsLockBCH(l types.Bytes) Lock {
	return &fundsLockBCH{
		fundsLockBTC: newFundsLockBTC(l).(fundsLockBTC),
	}
}

// LockData implement Lock
func (fl *fundsLockBCH) LockData() (*LockData, error) {
	return parseLockScript(cryptos.BitcoinCash, fl.fundsLockBTC)
}

// Address implement Lock
func (fl *fundsLockBCH) Address(chain params.Chain) (string, error) {
	return networks.All[cryptos.BitcoinCash][chain].P2SHFromScript(fl.fundsLockBTC)
}

// MarshalYAML implement yaml.Marshaler
func (fl *fundsLockBCH) MarshalYAML() (interface{}, error) { return fl.fundsLockBTC.Bytes().Hex(), nil }

// UnmarshalYAML implement yaml.Unmarshaler
func (fl *fundsLockBCH) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := fundsLockBTC{}
	if err := unmarshal(&r); err != nil {
		return err
	}
	fl.fundsLockBTC = r
	return nil
}
