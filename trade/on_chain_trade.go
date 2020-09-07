package trade

import (
	"time"

	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/duration"
	"github.com/transmutate-io/atomicswap/key"
	"github.com/transmutate-io/atomicswap/roles"
	"github.com/transmutate-io/cryptocore/types"
)

type OnChainTrade struct{ *baseTrade }

// NewOnChainTrade returns a new on-chain buyer trade
func NewOnChainTrade(
	ownAmount types.Amount,
	ownCrypto *cryptos.Crypto,
	traderAmount types.Amount,
	traderCrypto *cryptos.Crypto,
	dur time.Duration,
) (Trade, error) {
	bt, err := newBuyerBaseTrade(
		dur,
		ownAmount,
		ownCrypto,
		traderAmount,
		traderCrypto,
	)
	if err != nil {
		return nil, err
	}
	return &OnChainTrade{baseTrade: bt}, nil
}

// AcceptProposal accepts a proposal and returns a new on-chain seller trade
func AcceptProposal(prop *BuyProposal) (Trade, error) {
	r := &OnChainTrade{baseTrade: &baseTrade{Role: roles.Seller}}
	if err := r.AcceptBuyProposal(prop); err != nil {
		return nil, err
	}
	return r, nil
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

func (t *OnChainTrade) MarshalYAML() (interface{}, error) { return t.baseTrade, nil }

func (t *OnChainTrade) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := &baseTrade{}
	if err := unmarshal(r); err != nil {
		return err
	}
	t.baseTrade = r
	return nil
}

func (t *OnChainTrade) Buyer() (BuyerTrade, error) {
	if t.baseTrade.Role != roles.Buyer {
		return nil, ErrNotABuyerTrade
	}
	return t, nil
}

func (t *OnChainTrade) Seller() (SellerTrade, error) {
	if t.baseTrade.Role != roles.Seller {
		return nil, ErrNotASellerTrade
	}
	return t, nil
}
