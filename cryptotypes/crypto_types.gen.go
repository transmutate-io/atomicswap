package cryptotypes

import "fmt"

type InvalidCryptoTypeError string

func (e InvalidCryptoTypeError) Error() string { return fmt.Sprintf("invalid cryptotype: \"%s\"", string(e)) }

type CryptoType int

func ParseCryptoType(s string) (CryptoType, error) {
	var r CryptoType
	if err := (&r).Set(s); err != nil {
		return 0, err
	}
	return r, nil
}

func (v CryptoType) String() string { return _CryptoType[v] }

func (v *CryptoType) Set(sv string) error {
	nv, ok := _CryptoTypeNames[sv]
	if !ok {
		return InvalidCryptoTypeError(sv)
	}
	*v = nv
	return nil
}

func (v CryptoType) MarshalYAML() (interface{}, error) { return v.String(), nil }

func (v *CryptoType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	return v.Set(r)
}

const (
 	UTXO CryptoType = iota
 	StateBased
)

var (
	_CryptoType = map[CryptoType]string{
		UTXO:       "utxo",
		StateBased: "state-based",
	}
	_CryptoTypeNames map[string]CryptoType
)

func init() {
	_CryptoTypeNames = make(map[string]CryptoType, len(_CryptoType))
	for k, v := range _CryptoType {
		_CryptoTypeNames[v] = k
	}
}
