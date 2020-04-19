package networks

import (
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/params"
)

type chains = map[params.Chain]params.Params

// Networks contains all available cryptos/networks
var (
	Networks       = make(map[*cryptos.Crypto]map[params.Chain]params.Params, len(NetworksByName))
	NetworksByName = map[string]chains{
		"bitcoin": chains{
			params.MainNet:       params.BTC_MainNet,
			params.TestNet:       params.BTC_TestNet,
			params.SimNet:        params.BTC_SimNet,
			params.RegressionNet: params.BTC_RegressionNet,
		},
		"litecoin": chains{
			params.MainNet:       params.LTC_MainNet,
			params.TestNet:       params.LTC_TestNet,
			params.SimNet:        params.LTC_SimNet,
			params.RegressionNet: params.LTC_RegressionNet,
		},
		"dogecoin": chains{
			params.MainNet:       params.DOGE_MainNet,
			params.TestNet:       params.DOGE_TestNet,
			params.RegressionNet: params.DOGE_RegressionNet,
		},
		"bitcoin-cash": chains{
			params.MainNet:       params.BCH_MainNet,
			params.TestNet:       params.BCH_TestNet,
			params.SimNet:        params.BCH_SimNet,
			params.RegressionNet: params.BCH_RegressionNet,
		},
	}
)

func init() {
	for cn, c := range NetworksByName {
		cr, err := cryptos.ParseCrypto(cn)
		if err != nil {
			panic("unknown crypto: " + cn)
		}
		Networks[cr] = c
	}
}
