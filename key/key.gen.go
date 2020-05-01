package key

var cryptoFuncs = map[string]newFuncs{
	"bitcoin-cash": newFuncs{
		parsePriv: ParsePrivateBCH,
		priv:      NewPrivateBCH,
		pub:       NewPublicBCH,
	},
	"bitcoin": newFuncs{
		parsePriv: ParsePrivateBTC,
		priv:      NewPrivateBTC,
		pub:       NewPublicBTC,
	},
	"dogecoin": newFuncs{
		parsePriv: ParsePrivateDOGE,
		priv:      NewPrivateDOGE,
		pub:       NewPublicDOGE,
	},
	"litecoin": newFuncs{
		parsePriv: ParsePrivateLTC,
		priv:      NewPrivateLTC,
		pub:       NewPublicLTC,
	},
}