package atomicswap

import (
	"transmutate.io/pkg/cryptocore/types"
)

// Output represents an output
type Output struct {
	TxID   types.Bytes `yaml:"txid"`
	N      uint32      `yaml:"n"`
	Amount uint64      `yaml:"amount"`
}

type fundsUTXO struct {
	Outputs []*Output `yaml:"outputs"`
}

func (f *fundsUTXO) Len() int                { return len(f.Outputs) }
func (f *fundsUTXO) Idx(idx int) interface{} { return f.Outputs[idx] }
func (f *fundsUTXO) Data() interface{}       { return f.Outputs }

func (f *fundsUTXO) Add(fd interface{}) {
	o, ok := fd.(*Output)
	if !ok {
		panic("not an output")
	}
	f.Outputs = append(f.Outputs, o)
}

func newFundsUTXO() *fundsUTXO {
	return &fundsUTXO{Outputs: make([]*Output, 0, 4)}
}

// type tradeInfo struct {
// 	Crypto      string `yaml:"crypto"`
// 	RedeemKey   string `yaml:"redeem_key,omitempty"`
// 	RecoveryKey string `yaml:"recover_key,omitempty"`
// }

// func (ti *TradeInfo) UnmarshalYAML(unmarshal func(interface{}) error) error {
// 	t := &tradeInfo{}
// 	err := unmarshal(t)
// 	if err != nil {
// 		return err
// 	}
// 	ti.Crypto, err = cryptos.ParseCrypto(t.Crypto)
// 	if err != nil {
// 		return err
// 	}
// 	priv, err := ti.Crypto.NewPrivateKey()
// 	if err != nil {
// 		return err
// 	}
// 	if err = 	yaml.Unmarshal([]byte(t.RecoveryKey), priv); err != nil {
// 		return err
// 	}
// 	ti.RecoveryKey = priv
// 	if priv, err = ti.Crypto.NewPrivateKey(); err != nil {
// 		return err
// 	}
// 	if err = yaml.Unmarshal([]byte(t.RedeemKey), priv); err != nil {
// 		return err
// 	}
// 	ti.RedeemKey = priv
// 	return nil
// }
