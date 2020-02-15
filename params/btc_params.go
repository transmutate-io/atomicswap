package params

var (
	BTC_MainNet = &Params{
		pubKeyHashAddrID: 0x00, // starts with 1
		scriptHashAddrID: 0x05, // starts with 3
		privateKeyID:     0x80, // starts with 5 (uncompressed) or K (compressed)
	}
	BTC_TestNet = &Params{
		pubKeyHashAddrID: 0x6f, // starts with m or n
		scriptHashAddrID: 0xc4, // starts with 2
		privateKeyID:     0xef, // starts with 9 (uncompressed) or c (compressed)
	}
	BTC_RegressionNet = &Params{
		pubKeyHashAddrID: 0x6f, // starts with m or n
		scriptHashAddrID: 0xc4, // starts with 2
		privateKeyID:     0xef, // starts with 9 (uncompressed) or c (compressed)
	}
	BTC_SimNet = &Params{
		pubKeyHashAddrID: 0x3f, // starts with S
		scriptHashAddrID: 0x7b, // starts with s
		privateKeyID:     0x64, // starts with 4 (uncompressed) or F (compressed)
	}
)

func init() {
	Networks[Bitcoin] = map[Chain]*Params{
		MainNet:       BTC_MainNet,
		TestNet:       BTC_TestNet,
		SimNet:        BTC_SimNet,
		RegressionNet: BTC_RegressionNet,
	}
}
