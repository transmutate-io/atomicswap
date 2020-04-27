package key

var cryptoFuncs = map[string]newFuncs{
	"bitcoin": newFuncs{
		parsePriv: ParsePrivateBTC,
		priv:      NewPrivateBTC,
		pub:       NewPublicBTC,
	},
	"litecoin": newFuncs{
		parsePriv: ParsePrivateLTC,
		priv:      NewPrivateLTC,
		pub:       NewPublicLTC,
	},
	"dogecoin": newFuncs{
		parsePriv: ParsePrivateDOGE,
		priv:      NewPrivateDOGE,
		pub:       NewPublicDOGE,
	},
	"bitcoin-cash": newFuncs{
		parsePriv: ParsePrivateBCH,
		priv:      NewPrivateBCH,
		pub:       NewPublicBCH,
	},
}
