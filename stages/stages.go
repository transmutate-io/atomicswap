package stages

import (
	"transmutate.io/pkg/atomicswap/roles"
)

func stagesConcat(s ...[]Stage) []Stage {
	r := make([]Stage, 0, 1024)
	for _, i := range s {
		r = append(r, i...)
	}
	return r
}

var (
	genBuyer  = []Stage{GenerateKeys, GenerateToken}
	propBuyer = []Stage{ShareProposal, ReceiveProposalResponse}
	lockBuyer = []Stage{
		ReceiveKeyData,
		GenerateLock,
		ShareLock,
		ShareKeyData,
		ReceiveLock,
	}
	fundsBuyer = []Stage{LockFunds, WaitLockedFunds, RedeemFunds}
	genSeller  = []Stage{GenerateKeys}
	propSeller = []Stage{ReceiveProposal, ShareProposalResponse}
	lockSeller = []Stage{
		ShareKeyData,
		ReceiveLock,
		ReceiveKeyData,
		GenerateLock,
		ShareLock,
	}
	fundsSeller = []Stage{
		WaitLockedFunds,
		LockFunds,
		WaitRedeemableFunds,
		RedeemFunds,
	}
	done = []Stage{Done}

	StagesManualExchange = map[roles.Role][]Stage{
		roles.Buyer: stagesConcat(
			genBuyer,
			propBuyer,
			lockBuyer,
			fundsBuyer,
			done,
		),
		roles.Seller: stagesConcat(
			genSeller,
			propSeller,
			lockSeller,
			fundsSeller,
			done,
		),
	}
)
