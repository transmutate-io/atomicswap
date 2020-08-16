package networks

import (
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/params"
)

type chains = map[params.Chain]params.Params

var (
	All = map[*cryptos.Crypto]chains{
		cryptos.Bitcoin: chains{
			params.MainNet:       params.BTC_MainNet,
			params.TestNet:       params.BTC_TestNet,
			params.SimNet:        params.BTC_SimNet,
			params.RegressionNet: params.BTC_RegressionNet,
		},
		cryptos.Litecoin: chains{
			params.MainNet:       params.LTC_MainNet,
			params.TestNet:       params.LTC_TestNet,
			params.SimNet:        params.LTC_SimNet,
			params.RegressionNet: params.LTC_RegressionNet,
		},
		cryptos.Dogecoin: chains{
			params.MainNet:       params.DOGE_MainNet,
			params.TestNet:       params.DOGE_TestNet,
			params.RegressionNet: params.DOGE_RegressionNet,
		},
		cryptos.BitcoinCash: chains{
			params.MainNet:       params.BCH_MainNet,
			params.TestNet:       params.BCH_TestNet,
			params.SimNet:        params.BCH_SimNet,
			params.RegressionNet: params.BCH_RegressionNet,
		},
		cryptos.Decred: chains{
			params.MainNet:       params.DCR_MainNet,
			params.TestNet:       params.DCR_TestNet,
			params.SimNet:        params.DCR_SimNet,
			params.RegressionNet: params.DCR_RegressionNet,
		},
	}
	AllByName        = make(map[string]chains, len(All))
	Main             = make(map[*cryptos.Crypto]params.Params, len(All))
	MainByName       = make(map[string]params.Params, len(All))
	Test             = make(map[*cryptos.Crypto]params.Params, len(All))
	TestByName       = make(map[string]params.Params, len(All))
	Sim              = make(map[*cryptos.Crypto]params.Params, len(All))
	SimByName        = make(map[string]params.Params, len(All))
	Regression       = make(map[*cryptos.Crypto]params.Params, len(All))
	RegressionByName = make(map[string]params.Params, len(All))
)

func init() {
	for cn, c := range All {
		AllByName[cn.Name] = c
		if p, ok := c[params.MainNet]; ok {
			MainByName[cn.Name] = p
			Main[cn] = p
		}
		if p, ok := c[params.TestNet]; ok {
			TestByName[cn.Name] = p
			Test[cn] = p
		}
		if p, ok := c[params.SimNet]; ok {
			SimByName[cn.Name] = p
			Sim[cn] = p
		}
		if p, ok := c[params.RegressionNet]; ok {
			RegressionByName[cn.Name] = p
			Regression[cn] = p
		}
	}
}
