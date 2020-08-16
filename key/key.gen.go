package key

var cryptoFuncs = map[string]newFuncs{
	"bitcoin-cash": newFuncs{
		parsePriv: ParsePrivateBCH,
		parsePub:  ParsePublicBCH,
		newPriv:   NewPrivateBCH,
	},
	"bitcoin": newFuncs{
		parsePriv: ParsePrivateBTC,
		parsePub:  ParsePublicBTC,
		newPriv:   NewPrivateBTC,
	},
	"decred": newFuncs{
		parsePriv: ParsePrivateDCR,
		parsePub:  ParsePublicDCR,
		newPriv:   NewPrivateDCR,
	},
	"dogecoin": newFuncs{
		parsePriv: ParsePrivateDOGE,
		parsePub:  ParsePublicDOGE,
		newPriv:   NewPrivateDOGE,
	},
	"litecoin": newFuncs{
		parsePriv: ParsePrivateLTC,
		parsePub:  ParsePublicLTC,
		newPriv:   NewPrivateLTC,
	},
}
