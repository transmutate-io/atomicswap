package networks

import (
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/params"
)

type chains = map[params.Chain]params.Params

// All contains all available cryptos/networks
var (
	AllByName = map[string]chains{
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
	All              = make(map[*cryptos.Crypto]map[params.Chain]params.Params, len(AllByName))
	Main             = make(map[*cryptos.Crypto]params.Params, len(AllByName))
	MainByName       = make(map[string]params.Params, len(AllByName))
	Test             = make(map[*cryptos.Crypto]params.Params, len(AllByName))
	TestByName       = make(map[string]params.Params, len(AllByName))
	Sim              = make(map[*cryptos.Crypto]params.Params, len(AllByName))
	SimByName        = make(map[string]params.Params, len(AllByName))
	Regression       = make(map[*cryptos.Crypto]params.Params, len(AllByName))
	RegressionByName = make(map[string]params.Params, len(AllByName))
)

func init() {
	for cn, c := range AllByName {
		cr, err := cryptos.Parse(cn)
		if err != nil {
			panic("unknown crypto: " + cn)
		}
		All[cr] = c
		if p, ok := c[params.MainNet]; ok {
			MainByName[cn] = p
			Main[cr] = p
		}
		if p, ok := c[params.TestNet]; ok {
			TestByName[cn] = p
			Test[cr] = p
		}
		if p, ok := c[params.SimNet]; ok {
			SimByName[cn] = p
			Sim[cr] = p
		}
		if p, ok := c[params.RegressionNet]; ok {
			RegressionByName[cn] = p
			Regression[cr] = p
		}
	}
}
