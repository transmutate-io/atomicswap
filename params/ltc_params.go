package params

type ltcParams = btcParams

var (
	// LTC_MainNet represents the litecoin main net
	LTC_MainNet = &ltcParams{
		pubKeyHashAddrID: 0x30, // starts with L
		scriptHashAddrID: 0x32, // starts with M
		privateKeyID:     0xB0, // starts with 6 (uncompressed) or T (compressed)
	}
	// LTC_TestNet represents the litecoin test net
	LTC_TestNet = &ltcParams{
		pubKeyHashAddrID: 0x6f, // starts with m or n
		scriptHashAddrID: 0x3a, // starts with Q
		privateKeyID:     0xef, // starts with 9 (uncompressed) or c (compressed)
	}
	// LTC_SimNet represents the litecoin simulation net
	LTC_SimNet = &ltcParams{
		pubKeyHashAddrID: 0x3f, // starts with S
		scriptHashAddrID: 0x7b, // starts with s
		privateKeyID:     0x64, // starts with 4 (uncompressed) or F (compressed)
	}
	// LTC_RegressionNet represents the litecoin regression test net
	LTC_RegressionNet = &ltcParams{
		pubKeyHashAddrID: 0x6f, // starts with m or n
		scriptHashAddrID: 0x3a, // starts with Q
		privateKeyID:     0xef, // starts with 9 (uncompressed) or c (compressed)
	}
)

func init() {
	Networks["litecoin"] = map[Chain]Params{
		MainNet:       LTC_MainNet,
		TestNet:       LTC_TestNet,
		SimNet:        LTC_SimNet,
		RegressionNet: LTC_RegressionNet,
	}
}
