package trade

import (
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/networks"
	"github.com/transmutate-io/atomicswap/params"
	"github.com/transmutate-io/cryptocore/types"
)

type fundsDataBTC struct {
	Outputs    []*Output   `yaml:"outputs"`
	LockScript types.Bytes `yaml:"lock_script"`
}

func newFundsDataBTC() FundsData { return &fundsDataBTC{Outputs: make([]*Output, 0, 4)} }

// Funds implement FundsData
func (f *fundsDataBTC) Funds() interface{} { return f.Outputs }

// AddFunds implement FundsData
func (f *fundsDataBTC) AddFunds(funds interface{}) {
	f.Outputs = append(f.Outputs, funds.(*Output))
}

// Lock implement FundsData
func (f fundsDataBTC) Lock() Lock { return fundsLockBTC(f.LockScript) }

// SetLock implement FundsData
func (f *fundsDataBTC) SetLock(lock Lock) { f.LockScript = lock.Bytes() }

type fundsLockBTC types.Bytes

func newFundsLockBTC(l types.Bytes) Lock { return fundsLockBTC(l) }

// Bytes implement Lock
func (fl fundsLockBTC) Bytes() types.Bytes { return types.Bytes(fl) }

// Data implement Lock
func (fl fundsLockBTC) Data() types.Bytes { return fl.Bytes() }

// LockData implement Lock
func (fl fundsLockBTC) LockData() (*LockData, error) {
	return parseLockScript(cryptos.Bitcoin, fl)
}

// Address implement Lock
func (fl fundsLockBTC) Address(chain params.Chain) (string, error) {
	return networks.All[cryptos.Bitcoin][chain].P2SHFromScript(fl)
}

// MarshalYAML implement yaml.Marshaler
func (fl fundsLockBTC) MarshalYAML() (interface{}, error) { return types.Bytes(fl).Hex(), nil }

// UnmarshalYAML implement yaml.Unmarshaler
func (fl *fundsLockBTC) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := types.Bytes([]byte{})
	if err := unmarshal(&r); err != nil {
		return err
	}
	*fl = fundsLockBTC(r)
	return nil
}
