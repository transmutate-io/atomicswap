package params

import "fmt"

type InvalidCryptoError string

func (e InvalidCryptoError) Error() string { return fmt.Sprintf("invalid crypto: \"%s\"", string(e)) }

type Crypto int

func ParseCrypto(s string) (Crypto, error) {
	var r Crypto
	if err := (&r).Set(s); err != nil {
		return 0, err
	}
	return r, nil
}

func (v Crypto) String() string { return _Crypto[v] }

func (v *Crypto) Set(sv string) error {
	nv, ok := _CryptoNames[sv]
	if !ok {
		return InvalidCryptoError(sv)
	}
	*v = nv
	return nil
}

func (v Crypto) MarshalYAML() (interface{}, error) { return v.String(), nil }

func (v *Crypto) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	return v.Set(r)
}

const (
 	Bitcoin Crypto = iota
 	Litecoin
)

var (
	_Crypto = map[Crypto]string{
		Bitcoin:  "bitcoin",
		Litecoin: "litecoin",
	}
	_CryptoNames map[string]Crypto
)

func init() {
	_CryptoNames = make(map[string]Crypto, len(_Crypto))
	for k, v := range _Crypto {
		_CryptoNames[v] = k
	}
}
