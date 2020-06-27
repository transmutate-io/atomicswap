package {{ .Values.package }}

import (
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/networks"
	"transmutate.io/pkg/atomicswap/params"
	"transmutate.io/pkg/cryptocore/types"
)

type fundsData{{ .Values.short }} struct{ *fundsDataBTC }

func newFundsData{{ .Values.short }}() FundsData {
	return &fundsData{{ .Values.short }}{
		fundsDataBTC: newFundsDataBTC().(*fundsDataBTC),
	}
}

func (fd *fundsData{{ .Values.short }}) MarshalYAML() (interface{}, error) {
	return fd.fundsDataBTC, nil
}

func (fd *fundsData{{ .Values.short }}) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := &fundsDataBTC{}
	if err := unmarshal(r); err != nil {
		return err
	}
	fd.fundsDataBTC = r
	return nil
}

func (fd *fundsData{{ .Values.short }}) Lock() Lock { return &fundsLock{{ .Values.short }}{fundsLockBTC(fd.LockScript)} }

type fundsLock{{ .Values.short }} struct{ fundsLockBTC }

func newFundsLock{{ .Values.short }}(l types.Bytes) Lock {
	return &fundsLock{{ .Values.short }}{
		fundsLockBTC: newFundsLockBTC(l).(fundsLockBTC),
	}
}

func (fl *fundsLock{{ .Values.short }}) LockData() (*LockData, error) {
	return parseLockScript(cryptos.{{ title ( dashed_to_camel .Values.name ) }}, fl.fundsLockBTC)
}

func (fl *fundsLock{{ .Values.short }}) Address(chain params.Chain) (string, error) {
	return networks.All[cryptos.{{ title ( dashed_to_camel .Values.name ) }}][chain].P2SHFromScript(fl.fundsLockBTC)
}


func (fl *fundsLock{{ .Values.short }}) MarshalYAML() (interface{}, error) { return fl.fundsLockBTC, nil }

func (fl *fundsLock{{ .Values.short }}) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := fundsLockBTC{}
	if err := unmarshal(&r); err != nil {
		return err
	}
	fl.fundsLockBTC = r
	return nil
}
