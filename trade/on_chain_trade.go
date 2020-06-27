package trade

import (
	"time"

	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/duration"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/roles"
	"transmutate.io/pkg/atomicswap/stages"
	"transmutate.io/pkg/cryptocore/types"
)

type OnChainTrade struct{ *baseTrade }

// NewOnChainBuy returns a new buyer on-chain trade
func NewOnChainBuy(
	ownAmount types.Amount,
	ownCrypto *cryptos.Crypto,
	traderAmount types.Amount,
	traderCrypto *cryptos.Crypto,
	dur time.Duration,
) (Trade, error) {
	ownFundsData, err := newFundsData(ownCrypto)
	if err != nil {
		return nil, err
	}
	traderFundData, err := newFundsData(traderCrypto)
	if err != nil {
		return nil, err
	}
	return &OnChainTrade{
		baseTrade: &baseTrade{
			Role:     roles.Buyer,
			Duration: duration.Duration(dur),
			Stager:   stages.NewStager(tradeStages[roles.Buyer]...),
			OwnInfo: &TraderInfo{
				Amount: ownAmount,
				Crypto: ownCrypto,
			},
			TraderInfo: &TraderInfo{
				Amount: traderAmount,
				Crypto: traderCrypto,
			},
			RecoverableFunds: ownFundsData,
			RedeemableFunds:  traderFundData,
		},
	}, nil
}

// NewOnChainSell returns a new seller on-chain trade
func NewOnChainSell() Trade {
	return &OnChainTrade{
		baseTrade: &baseTrade{
			Role:   roles.Seller,
			Stager: stages.NewStager(tradeStages[roles.Seller]...),
		},
	}
}

func (t *OnChainTrade) Role() roles.Role { return t.baseTrade.Role }

func (t *OnChainTrade) Duration() duration.Duration { return t.baseTrade.Duration }

func (t *OnChainTrade) Token() types.Bytes { return t.baseTrade.Token }

func (t *OnChainTrade) TokenHash() types.Bytes { return t.baseTrade.TokenHash }

func (t *OnChainTrade) OwnInfo() *TraderInfo { return t.baseTrade.OwnInfo }

func (t *OnChainTrade) TraderInfo() *TraderInfo { return t.baseTrade.TraderInfo }

func (t *OnChainTrade) RedeemKey() key.Private { return t.baseTrade.RedeemKey }

func (t *OnChainTrade) RecoveryKey() key.Private { return t.baseTrade.RecoveryKey }

func (t *OnChainTrade) RedeemableFunds() FundsData { return t.baseTrade.RedeemableFunds }

func (t *OnChainTrade) RecoverableFunds() FundsData { return t.baseTrade.RecoverableFunds }

func (t *OnChainTrade) Stager() *stages.Stager { return t.baseTrade.Stager }

func (t *OnChainTrade) MarshalYAML() (interface{}, error) { return t.baseTrade, nil }

func (t *OnChainTrade) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := &baseTrade{}
	if err := unmarshal(r); err != nil {
		return err
	}
	t.baseTrade = r
	return nil
}
