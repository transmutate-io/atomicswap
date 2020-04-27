package cryptos

import "fmt"

type InvalidTypeError string

func (e InvalidTypeError) Error() string {
	return fmt.Sprintf("invalid crypto type: \"%s\"", string(e))
}

type Type int

func ParseType(s string) (Type, error) {
	var r Type
	if err := (&r).Set(s); err != nil {
		return 0, err
	}
	return r, nil
}

func (v Type) String() string { return _Type[v] }

func (v *Type) Set(sv string) error {
	nv, ok := _TypeNames[sv]
	if !ok {
		return InvalidTypeError(sv)
	}
	*v = nv
	return nil
}

func (v Type) MarshalYAML() (interface{}, error) { return v.String(), nil }

func (v *Type) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	return v.Set(r)
}

const (
	UTXO Type = iota
	StateBased
)

var (
	_Type = map[Type]string{
		UTXO: "utxo",
		StateBased: "state-based",
	}
	_TypeNames map[string]Type
)

func init() {
	_TypeNames = make(map[string]Type, len(_Type))
	for k, v := range _Type {
		_TypeNames[v] = k
	}
}

