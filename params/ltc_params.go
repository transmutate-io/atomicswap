package params

var (
	LTC_MainNet = &Params{
		// Name:        "mainnet",
		// Net:         wire.MainNet,
		// DefaultPort: "9333",
		// DNSSeeds: []DNSSeed{
		// 	{"seed-a.litecoin.loshan.co.uk", true},
		// 	{"dnsseed.thrasher.io", true},
		// 	{"dnsseed.litecointools.com", false},
		// 	{"dnsseed.litecoinpool.org", false},
		// 	{"dnsseed.koin-project.com", false},
		// },
		// // Chain parameters
		// GenesisBlock:             &genesisBlock,
		// GenesisHash:              &genesisHash,
		// PowLimit:                 mainPowLimit,
		// PowLimitBits:             504365055,
		// BIP0034Height:            710000,
		// BIP0065Height:            918684,
		// BIP0066Height:            811879,
		// CoinbaseMaturity:         100,
		// SubsidyReductionInterval: 840000,
		// TargetTimespan:           (time.Hour * 24 * 3) + (time.Hour * 12), // 3.5 days
		// TargetTimePerBlock:       (time.Minute * 2) + (time.Second * 30),  // 2.5 minutes
		// RetargetAdjustmentFactor: 4,                                       // 25% less, 400% more
		// ReduceMinDifficulty:      false,
		// MinDiffReductionTime:     0,
		// GenerateSupported:        false,
		// // Checkpoints ordered from oldest to newest.
		// Checkpoints: []Checkpoint{
		// 	{1500, newHashFromStr("841a2965955dd288cfa707a755d05a54e45f8bd476835ec9af4402a2b59a2967")},
		// 	{4032, newHashFromStr("9ce90e427198fc0ef05e5905ce3503725b80e26afd35a987965fd7e3d9cf0846")},
		// 	{8064, newHashFromStr("eb984353fc5190f210651f150c40b8a4bab9eeeff0b729fcb3987da694430d70")},
		// 	{16128, newHashFromStr("602edf1859b7f9a6af809f1d9b0e6cb66fdc1d4d9dcd7a4bec03e12a1ccd153d")},
		// 	{23420, newHashFromStr("d80fdf9ca81afd0bd2b2a90ac3a9fe547da58f2530ec874e978fce0b5101b507")},
		// 	{50000, newHashFromStr("69dc37eb029b68f075a5012dcc0419c127672adb4f3a32882b2b3e71d07a20a6")},
		// 	{80000, newHashFromStr("4fcb7c02f676a300503f49c764a89955a8f920b46a8cbecb4867182ecdb2e90a")},
		// 	{120000, newHashFromStr("bd9d26924f05f6daa7f0155f32828ec89e8e29cee9e7121b026a7a3552ac6131")},
		// 	{161500, newHashFromStr("dbe89880474f4bb4f75c227c77ba1cdc024991123b28b8418dbbf7798471ff43")},
		// 	{179620, newHashFromStr("2ad9c65c990ac00426d18e446e0fd7be2ffa69e9a7dcb28358a50b2b78b9f709")},
		// 	{240000, newHashFromStr("7140d1c4b4c2157ca217ee7636f24c9c73db39c4590c4e6eab2e3ea1555088aa")},
		// 	{383640, newHashFromStr("2b6809f094a9215bafc65eb3f110a35127a34be94b7d0590a096c3f126c6f364")},
		// 	{409004, newHashFromStr("487518d663d9f1fa08611d9395ad74d982b667fbdc0e77e9cf39b4f1355908a3")},
		// 	{456000, newHashFromStr("bf34f71cc6366cd487930d06be22f897e34ca6a40501ac7d401be32456372004")},
		// 	{638902, newHashFromStr("15238656e8ec63d28de29a8c75fcf3a5819afc953dcd9cc45cecc53baec74f38")},
		// 	{721000, newHashFromStr("198a7b4de1df9478e2463bd99d75b714eab235a2e63e741641dc8a759a9840e5")},
		// },
		// // Consensus rule change deployments.
		// //
		// // The miner confirmation window is defined as:
		// //   target proof of work timespan / target proof of work spacing
		// RuleChangeActivationThreshold: 6048, // 75% of MinerConfirmationWindow
		// MinerConfirmationWindow:       8064, //
		// Deployments: [DefinedDeployments]ConsensusDeployment{
		// 	DeploymentTestDummy: {
		// 		BitNumber:  28,
		// 		StartTime:  1199145601, // January 1, 2008 UTC
		// 		ExpireTime: 1230767999, // December 31, 2008 UTC
		// 	},
		// 	DeploymentCSV: {
		// 		BitNumber:  0,
		// 		StartTime:  1485561600, // January 28, 2017 UTC
		// 		ExpireTime: 1517356801, // January 31st, 2018 UTC
		// 	},
		// 	DeploymentSegwit: {
		// 		BitNumber:  1,
		// 		StartTime:  1485561600, // January 28, 2017 UTC
		// 		ExpireTime: 1517356801, // January 31st, 2018 UTC.
		// 	},
		// },
		// // Mempool parameters
		// RelayNonStdTxs: false,
		// // Human-readable part for Bech32 encoded segwit addresses, as defined in
		// // BIP 173.
		// Bech32HRPSegwit: "ltc", // always ltc for main net
		// Address encoding magics
		pubKeyHashAddrID: 0x30, // starts with L
		scriptHashAddrID: 0x32, // starts with M
		privateKeyID:     0xB0, // starts with 6 (uncompressed) or T (compressed)
		// WitnessPubKeyHashAddrID: 0x06, // starts with p2
		// WitnessScriptHashAddrID: 0x0A, // starts with 7Xh
		// // BIP32 hierarchical deterministic extended key magics
		// HDPrivateKeyID: [4]byte{0x04, 0x88, 0xad, 0xe4}, // starts with xprv
		// HDPublicKeyID:  [4]byte{0x04, 0x88, 0xb2, 0x1e}, // starts with xpub
		// // BIP44 coin type used in the hierarchical deterministic path for
		// // address generation.
		// HDCoinType: 2,
	}
	LTC_TestNet = &Params{
		// Name:        "testnet4",
		// Net:         wire.TestNet4,
		// DefaultPort: "19335",
		// DNSSeeds: []DNSSeed{
		// 	{"testnet-seed.litecointools.com", false},
		// 	{"seed-b.litecoin.loshan.co.uk", true},
		// 	{"dnsseed-testnet.thrasher.io", true},
		// },
		// // Chain parameters
		// GenesisBlock:             &testNet4GenesisBlock,
		// GenesisHash:              &testNet4GenesisHash,
		// PowLimit:                 testNet4PowLimit,
		// PowLimitBits:             504365055,
		// BIP0034Height:            76,
		// BIP0065Height:            76,
		// BIP0066Height:            76,
		// CoinbaseMaturity:         100,
		// SubsidyReductionInterval: 840000,
		// TargetTimespan:           (time.Hour * 24 * 3) + (time.Hour * 12), // 3.5 days
		// TargetTimePerBlock:       (time.Minute * 2) + (time.Second * 30),  // 2.5 minutes
		// RetargetAdjustmentFactor: 4,                                       // 25% less, 400% more
		// ReduceMinDifficulty:      true,
		// MinDiffReductionTime:     time.Minute * 5, // TargetTimePerBlock * 2
		// GenerateSupported:        false,
		// // Checkpoints ordered from oldest to newest.
		// Checkpoints: []Checkpoint{
		// 	{26115, newHashFromStr("817d5b509e91ab5e439652eee2f59271bbc7ba85021d720cdb6da6565b43c14f")},
		// 	{43928, newHashFromStr("7d86614c153f5ef6ad878483118ae523e248cd0dd0345330cb148e812493cbb4")},
		// 	{69296, newHashFromStr("66c2f58da3cfd282093b55eb09c1f5287d7a18801a8ff441830e67e8771010df")},
		// 	{99949, newHashFromStr("8dd471cb5aecf5ead91e7e4b1e932c79a0763060f8d93671b6801d115bfc6cde")},
		// 	{159256, newHashFromStr("ab5b0b9968842f5414804591119d6db829af606864b1959a25d6f5c114afb2b7")},
		// },
		// // Consensus rule change deployments.
		// //
		// // The miner confirmation window is defined as:
		// //   target proof of work timespan / target proof of work spacing
		// RuleChangeActivationThreshold: 1512, // 75% of MinerConfirmationWindow
		// MinerConfirmationWindow:       2016,
		// Deployments: [DefinedDeployments]ConsensusDeployment{
		// 	DeploymentTestDummy: {
		// 		BitNumber:  28,
		// 		StartTime:  1199145601, // January 1, 2008 UTC
		// 		ExpireTime: 1230767999, // December 31, 2008 UTC
		// 	},
		// 	DeploymentCSV: {
		// 		BitNumber:  0,
		// 		StartTime:  1483228800, // January 1, 2017
		// 		ExpireTime: 1517356801, // January 31st, 2018
		// 	},
		// 	DeploymentSegwit: {
		// 		BitNumber:  1,
		// 		StartTime:  1483228800, // January 1, 2017
		// 		ExpireTime: 1517356801, // January 31st, 2018
		// 	},
		// },
		// // Mempool parameters
		// RelayNonStdTxs: true,
		// // Human-readable part for Bech32 encoded segwit addresses, as defined in
		// // BIP 173.
		// Bech32HRPSegwit: "tltc", // always tltc for test net
		// Address encoding magics
		pubKeyHashAddrID: 0x6f, // starts with m or n
		scriptHashAddrID: 0x3a, // starts with Q
		privateKeyID:     0xef, // starts with 9 (uncompressed) or c (compressed)
		// WitnessPubKeyHashAddrID: 0x52, // starts with QW
		// WitnessScriptHashAddrID: 0x31, // starts with T7n
		// // BIP32 hierarchical deterministic extended key magics
		// HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94}, // starts with tprv
		// HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xcf}, // starts with tpub
		// // BIP44 coin type used in the hierarchical deterministic path for
		// // address generation.
		// HDCoinType: 1,
	}
	LTC_SimNet = &Params{
		// Name:        "simnet",
		// Net:         wire.SimNet,
		// DefaultPort: "18555",
		// DNSSeeds:    []DNSSeed{}, // NOTE: There must NOT be any seeds.
		// // Chain parameters
		// GenesisBlock:             &simNetGenesisBlock,
		// GenesisHash:              &simNetGenesisHash,
		// PowLimit:                 simNetPowLimit,
		// PowLimitBits:             0x207fffff,
		// BIP0034Height:            0, // Always active on simnet
		// BIP0065Height:            0, // Always active on simnet
		// BIP0066Height:            0, // Always active on simnet
		// CoinbaseMaturity:         100,
		// SubsidyReductionInterval: 210000,
		// TargetTimespan:           time.Hour * 24 * 14, // 14 days
		// TargetTimePerBlock:       time.Minute * 10,    // 10 minutes
		// RetargetAdjustmentFactor: 4,                   // 25% less, 400% more
		// ReduceMinDifficulty:      true,
		// MinDiffReductionTime:     time.Minute * 20, // TargetTimePerBlock * 2
		// GenerateSupported:        true,
		// // Checkpoints ordered from oldest to newest.
		// Checkpoints: nil,
		// // Consensus rule change deployments.
		// //
		// // The miner confirmation window is defined as:
		// //   target proof of work timespan / target proof of work spacing
		// RuleChangeActivationThreshold: 75, // 75% of MinerConfirmationWindow
		// MinerConfirmationWindow:       100,
		// Deployments: [DefinedDeployments]ConsensusDeployment{
		// 	DeploymentTestDummy: {
		// 		BitNumber:  28,
		// 		StartTime:  0,             // Always available for vote
		// 		ExpireTime: math.MaxInt64, // Never expires
		// 	},
		// 	DeploymentCSV: {
		// 		BitNumber:  0,
		// 		StartTime:  0,             // Always available for vote
		// 		ExpireTime: math.MaxInt64, // Never expires
		// 	},
		// 	DeploymentSegwit: {
		// 		BitNumber:  1,
		// 		StartTime:  0,             // Always available for vote
		// 		ExpireTime: math.MaxInt64, // Never expires.
		// 	},
		// },
		// // Mempool parameters
		// RelayNonStdTxs: true,
		// // Human-readable part for Bech32 encoded segwit addresses, as defined in
		// // BIP 173.
		// Bech32HRPSegwit: "sltc", // always lsb for sim net
		// Address encoding magics
		pubKeyHashAddrID: 0x3f, // starts with S
		scriptHashAddrID: 0x7b, // starts with s
		privateKeyID:     0x64, // starts with 4 (uncompressed) or F (compressed)
		// WitnessPubKeyHashAddrID: 0x19, // starts with Gg
		// WitnessScriptHashAddrID: 0x28, // starts with ?
		// // BIP32 hierarchical deterministic extended key magics
		// HDPrivateKeyID: [4]byte{0x04, 0x20, 0xb9, 0x00}, // starts with sprv
		// HDPublicKeyID:  [4]byte{0x04, 0x20, 0xbd, 0x3a}, // starts with spub
		// // BIP44 coin type used in the hierarchical deterministic path for
		// // address generation.
		// HDCoinType: 115, // ASCII for s
	}
	LTC_RegressionNet = &Params{
		// Name:        "regtest",
		// Net:         wire.TestNet,
		// DefaultPort: "18444",
		// DNSSeeds:    []DNSSeed{},
		// // Chain parameters
		// GenesisBlock:             &regTestGenesisBlock,
		// GenesisHash:              &regTestGenesisHash,
		// PowLimit:                 regressionPowLimit,
		// PowLimitBits:             0x207fffff,
		// CoinbaseMaturity:         100,
		// BIP0034Height:            100000000, // Not active - Permit ver 1 blocks
		// BIP0065Height:            1351,      // Used by regression tests
		// BIP0066Height:            1251,      // Used by regression tests
		// SubsidyReductionInterval: 150,
		// TargetTimespan:           time.Hour * 24 * 14, // 14 days
		// TargetTimePerBlock:       time.Minute * 10,    // 10 minutes
		// RetargetAdjustmentFactor: 4,                   // 25% less, 400% more
		// ReduceMinDifficulty:      true,
		// MinDiffReductionTime:     time.Minute * 20, // TargetTimePerBlock * 2
		// GenerateSupported:        true,
		// // Checkpoints ordered from oldest to newest.
		// Checkpoints: nil,
		// // Consensus rule change deployments.
		// //
		// // The miner confirmation window is defined as:
		// //   target proof of work timespan / target proof of work spacing
		// RuleChangeActivationThreshold: 108, // 75%  of MinerConfirmationWindow
		// MinerConfirmationWindow:       144,
		// Deployments: [DefinedDeployments]ConsensusDeployment{
		// 	DeploymentTestDummy: {
		// 		BitNumber:  28,
		// 		StartTime:  0,             // Always available for vote
		// 		ExpireTime: math.MaxInt64, // Never expires
		// 	},
		// 	DeploymentCSV: {
		// 		BitNumber:  0,
		// 		StartTime:  0,             // Always available for vote
		// 		ExpireTime: math.MaxInt64, // Never expires
		// 	},
		// 	DeploymentSegwit: {
		// 		BitNumber:  1,
		// 		StartTime:  0,             // Always available for vote
		// 		ExpireTime: math.MaxInt64, // Never expires.
		// 	},
		// },
		// // Mempool parameters
		// RelayNonStdTxs: true,
		// // Human-readable part for Bech32 encoded segwit addresses, as defined in
		// // BIP 173.
		// Bech32HRPSegwit: "rltc", // always rltc for reg test net
		// Address encoding magics
		pubKeyHashAddrID: 0x6f, // starts with m or n
		scriptHashAddrID: 0x3a, // starts with Q
		privateKeyID:     0xef, // starts with 9 (uncompressed) or c (compressed)
		// // BIP32 hierarchical deterministic extended key magics
		// HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94}, // starts with tprv
		// HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xcf}, // starts with tpub
		// // BIP44 coin type used in the hierarchical deterministic path for
		// // address generation.
		// HDCoinType: 1,
	}
)

func init() {
	Networks[Litecoin] = map[Chain]*Params{
		MainNet:       LTC_MainNet,
		TestNet:       LTC_TestNet,
		SimNet:        LTC_SimNet,
		RegressionNet: LTC_RegressionNet,
	}
}
