package params

import "fmt"

const (
	MainNet Chain = iota
	TestNet
	SimNet
	RegressionNet
)

var (
	_chainsByName map[string]Chain
	_chainNames   = map[Chain]string{
		MainNet:       "mainnet",
		TestNet:       "testnet",
		SimNet:        "simnet",
		RegressionNet: "regnet",
	}
)

type Chain int

func (c Chain) String() string { return _chainNames[c] }

// Parse parses a chain name
func ParseChain(c string) (Chain, error) {
	r, ok := _chainsByName[c]
	if !ok {
		return 0, UnknownChainError(c)
	}
	return r, nil
}

// UnknownChainError is returned when parsing an invalid chain name
type UnknownChainError string

func (e UnknownChainError) Error() string { return fmt.Sprintf(`unknown crypto: "%s"`, string(e)) }

func init() {
	_chainsByName = make(map[string]Chain, len(_chainNames))
	for c, n := range _chainNames {
		_chainsByName[n] = c
	}
}
