package stages

func stagesConcat(s ...[]Stage) []Stage {
	r := make([]Stage, 0, 1024)
	for _, i := range s {
		r = append(r, i...)
	}
	return r
}

// var (
// 	genBuyer  = []Stage{GenerateKeys, GenerateToken}
// 	propBuyer = []Stage{ShareProposal, ReceiveProposalResponse}
// 	lockBuyer = []Stage{
// 		ReceiveKeyData,
// 		GenerateLock,
// 		ShareLock,
// 		ShareKeyData,
// 		ReceiveLock,
// 	}
// 	fundsBuyer = []Stage{LockFunds, WaitLockedFunds, RedeemFunds}
// 	genSeller  = []Stage{GenerateKeys}
// 	propSeller = []Stage{ReceiveProposal, ShareProposalResponse}
// 	lockSeller = []Stage{
// 		ShareKeyData,
// 		ReceiveLock,
// 		ReceiveKeyData,
// 		GenerateLock,
// 		ShareLock,
// 	}
// 	fundsSeller = []Stage{
// 		WaitLockedFunds,
// 		LockFunds,
// 		WaitRedeemableFunds,
// 		RedeemFunds,
// 	}
// 	done = []Stage{Done}

// 	StagesManualExchange = map[roles.Role][]Stage{
// 		roles.Buyer: stagesConcat(
// 			genBuyer,
// 			propBuyer,
// 			lockBuyer,
// 			fundsBuyer,
// 			done,
// 		),
// 		roles.Seller: stagesConcat(
// 			genSeller,
// 			propSeller,
// 			lockSeller,
// 			fundsSeller,
// 			done,
// 		),
// 	}
// )

// buyer                                 | seller
//-----------------------------------------------------------------------------------
// sends proposal to seller              | receives proposal
// generate key                          | generate key
// generates a token                     |
// receives redeem key hash              | sends redeem key hash to buyer
// generates lockscript                  |
// sends lockscript to seller            | receive lockscript and checks
// sends redeem key hash to seller       | receive redeem key hash
//                                       | generates lockscript
// receive lockscript and checks         | sends lockscript to buyer
// locks funds with generated lockscript | waits for buyer lock transaction(s)
// waits for seller lock transaction(s)  | locks funds with generate lockscript
// redeems funds with token              | waits for buyer redeem transaction
//                                       | extract token from buyers redeem transaction
//                                       | redeems funds with token
