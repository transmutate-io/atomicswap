package stages

var (
	SellerStagesManualExchange = []Stage{
		ReceivePublicKeyHash,
		SharePublicKeyHash,
		ShareTokenHash,
		GenerateLockScript,
		ShareLockScript,
		ReceiveLockScript,
		LockFunds,
		WaitLockTransaction,
		RedeemFunds,
		Done,
	}
	BuyerStagesManualExchange = []Stage{
		SharePublicKeyHash,
		ReceivePublicKeyHash,
		ReceiveTokenHash,
		ReceiveLockScript,
		GenerateLockScript,
		ShareLockScript,
		WaitLockTransaction,
		LockFunds,
		WaitRedeemTransaction,
		RedeemFunds,
		Done,
	}
)
