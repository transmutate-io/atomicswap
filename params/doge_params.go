package params

type dogeParams = btcParams

var (
	// DOGE_MainNet represents the dogecoin main net
	DOGE_MainNet = &dogeParams{
		pubKeyHashAddrID: 0x1E,
		scriptHashAddrID: 0x16,
		privateKeyID:     0x9E,
	}
	// DOGE_TestNet represents the dogecoin test net
	DOGE_TestNet = &dogeParams{
		pubKeyHashAddrID: 0x71,
		scriptHashAddrID: 0xC4,
		privateKeyID:     0xF1,
	}
	// DOGE_RegressionNet represents the dogecoin regression test net
	DOGE_RegressionNet = &dogeParams{
		pubKeyHashAddrID: 0x6F,
		scriptHashAddrID: 0xc4,
		privateKeyID:     0xEF,
	}
)
