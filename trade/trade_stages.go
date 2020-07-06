package trade

import (
	"github.com/transmutate-io/atomicswap/roles"
	"github.com/transmutate-io/atomicswap/stages"
)

type (
	tradeStages   []stages.Stage
	tradeStageMap map[roles.Role]tradeStages
)

var (
	onChainTradeStages = tradeStageMap{
		roles.Buyer: tradeStages{
			// generate keys and token
			stages.GenerateKeys,
			stages.GenerateToken,
			// proposal exchange
			stages.SendProposal,
			stages.ReceiveProposalResponse,
			// funds locking
			stages.LockFunds,
			stages.WaitLockedFunds,
			// funds redeeming
			stages.RedeemFunds,
			// finished
			stages.Done,
		},
		roles.Seller: []stages.Stage{
			// proposal exchange
			stages.ReceiveProposal,
			stages.SendProposalResponse,
			// funds locking
			stages.WaitLockedFunds,
			stages.LockFunds,
			// funds redeeming
			stages.WaitFundsRedeemed,
			stages.RedeemFunds,
			// finished
			stages.Done,
		},
	}
	offChainTradeStages = tradeStageMap{
		roles.Buyer:  tradeStages{},
		roles.Seller: tradeStages{},
	}
)
