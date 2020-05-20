package trade

import (
	"transmutate.io/pkg/atomicswap/roles"
	"transmutate.io/pkg/atomicswap/stages"
)

var (
	tradeStages = map[roles.Role][]stages.Stage{
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
	// _tradeStages = make(map[roles.Role]map[stages.Stage]stages.Stage, 2)
)

// func init() {
// 	for r, s := range tradeStages {
// 		ts := make(map[stages.Stage]stages.Stage, len(s)+1)
// 		for i := 0; i < len(s)-1; i++ {
// 			ts[s[i]] = s[i+1]
// 		}
// 		ts[stages.Done] = stages.Done
// 		_tradeStages[r] = ts
// 	}
// }
