package params

import "fmt"

type InvalidChainError string

func (e InvalidChainError) Error() string { return fmt.Sprintf("invalid chain: \"%s\"", string(e)) }

type Chain int

func ParseChain(s string) (Chain, error) {
	var r Chain
	if err := (&r).Set(s); err != nil {
		return 0, err
	}
	return r, nil
}

func (v Chain) String() string { return _Chain[v] }

func (v *Chain) Set(sv string) error {
	nv, ok := _ChainNames[sv]
	if !ok {
		return InvalidChainError(sv)
	}
	*v = nv
	return nil
}

func (v Chain) MarshalYAML() (interface{}, error) { return v.String(), nil }

func (v *Chain) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	return v.Set(r)
}

const (
 	MainNet Chain = iota
 	TestNet
 	SimNet
 	RegressionNet
)

var (
	_Chain = map[Chain]string{
		TestNet:       "testnet",
		SimNet:        "simnet",
		RegressionNet: "regnet",
		MainNet:       "mainnet",
	}
	_ChainNames map[string]Chain
)

func init() {
	_ChainNames = make(map[string]Chain, len(_Chain))
	for k, v := range _Chain {
		_ChainNames[v] = k
	}
}
