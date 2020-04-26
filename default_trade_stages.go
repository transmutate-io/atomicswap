package atomicswap

import (
	"transmutate.io/pkg/atomicswap/roles"
	"transmutate.io/pkg/atomicswap/stages"
)

var (
	DefaultTradeStages = map[roles.Role][]stages.Stage{
		roles.Buyer: []stages.Stage{
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
)
