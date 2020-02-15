package params

import (
	"github.com/btcsuite/btcd/chaincfg"
)

var Networks = make(map[Crypto]map[Chain]*Params, 64)

type Params struct {
	// // Name defines a human-readable identifier for the network.
	// Name string
	// // Net defines the magic bytes used to identify the network.
	// Net wire.BitcoinNet
	// // DefaultPort defines the default peer-to-peer port for the network.
	// DefaultPort string
	// // GenesisBlock defines the first block of the chain.
	// GenesisBlock *wire.MsgBlock
	// // GenesisHash is the starting block hash.
	// GenesisHash *chainhash.Hash
	// // PowLimit defines the highest allowed proof of work value for a block
	// // as a uint256.
	// PowLimit *big.Int
	// // PowLimitBits defines the highest allowed proof of work value for a
	// // block in compact form.
	// PowLimitBits uint32
	// // These fields define the block heights at which the specified softfork
	// // BIP became active.
	// BIP0034Height int32
	// BIP0065Height int32
	// BIP0066Height int32
	// // CoinbaseMaturity is the number of blocks required before newly mined
	// // coins (coinbase transactions) can be spent.
	// CoinbaseMaturity uint16
	// // SubsidyReductionInterval is the interval of blocks before the subsidy
	// // is reduced.
	// SubsidyReductionInterval int32
	// // TargetTimespan is the desired amount of time that should elapse
	// // before the block difficulty requirement is examined to determine how
	// // it should be changed in order to maintain the desired block
	// // generation rate.
	// TargetTimespan time.Duration
	// // TargetTimePerBlock is the desired amount of time to generate each
	// // block.
	// TargetTimePerBlock time.Duration
	// // RetargetAdjustmentFactor is the adjustment factor used to limit
	// // the minimum and maximum amount of adjustment that can occur between
	// // difficulty retargets.
	// RetargetAdjustmentFactor int64
	// // ReduceMinDifficulty defines whether the network should reduce the
	// // minimum required difficulty after a long enough period of time has
	// // passed without finding a block.  This is really only useful for test
	// // networks and should not be set on a main network.
	// ReduceMinDifficulty bool
	// // MinDiffReductionTime is the amount of time after which the minimum
	// // required difficulty should be reduced when a block hasn't been found.
	// //
	// // NOTE: This only applies if ReduceMinDifficulty is true.
	// MinDiffReductionTime time.Duration
	// // GenerateSupported specifies whether or not CPU mining is allowed.
	// GenerateSupported bool
	// // Checkpoints ordered from oldest to newest.
	// Checkpoints []Checkpoint
	// // These fields are related to voting on consensus rule changes as
	// // defined by BIP0009.
	// //
	// // RuleChangeActivationThreshold is the number of blocks in a threshold
	// // state retarget window for which a positive vote for a rule change
	// // must be cast in order to lock in a rule change. It should typically
	// // be 95% for the main network and 75% for test networks.
	// //
	// // MinerConfirmationWindow is the number of blocks in each threshold
	// // state retarget window.
	// //
	// // Deployments define the specific consensus rule changes to be voted
	// // on.
	// RuleChangeActivationThreshold uint32
	// MinerConfirmationWindow       uint32
	// Deployments                   [DefinedDeployments]ConsensusDeployment
	// // Mempool parameters
	// RelayNonStdTxs bool
	// // Human-readable part for Bech32 encoded segwit addresses, as defined
	// // in BIP 173.
	// Bech32HRPSegwit string
	// // BIP32 hierarchical deterministic extended key magics
	// HDPrivateKeyID [4]byte
	// HDPublicKeyID  [4]byte

	// // BIP44 coin type used in the hierarchical deterministic path for
	// // address generation.
	// HDCoinType uint32

	// Address encoding magics
	pubKeyHashAddrID byte // First byte of a P2PKH address
	scriptHashAddrID byte // First byte of a P2SH address
	privateKeyID     byte // First byte of a WIF private key
	// WitnessPubKeyHashAddrID byte // First byte of a P2WPKH address
	// WitnessScriptHashAddrID byte // First byte of a P2WSH address
}

func (p Params) Params() *chaincfg.Params {
	return &chaincfg.Params{
		PubKeyHashAddrID: p.pubKeyHashAddrID,
		ScriptHashAddrID: p.scriptHashAddrID,
		PrivateKeyID:     p.privateKeyID,
	}
}
