package params

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"transmutate.io/pkg/atomicswap/hash"
)

// btcParams represents a network parameter set
type btcParams struct {
	pubKeyHashAddrID byte // First byte of a P2PKH address
	scriptHashAddrID byte // First byte of a P2SH address
	privateKeyID     byte // First byte of a WIF private key
}

// Params returns the chain params as a *chaincfg.Params
func (p *btcParams) params() *chaincfg.Params {
	return &chaincfg.Params{
		PubKeyHashAddrID: p.pubKeyHashAddrID,
		ScriptHashAddrID: p.scriptHashAddrID,
		PrivateKeyID:     p.privateKeyID,
	}
}

// P2PK returns the p2pk address for a key (same as p2pkh)
func (p *btcParams) P2PK(pub []byte) (string, error) {
	r, err := btcutil.NewAddressPubKey(pub, p.params())
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// P2PKH returns the p2pkh address for a key hash
func (p *btcParams) P2PKH(pubHash []byte) (string, error) {
	r, err := btcutil.NewAddressPubKeyHash(pubHash, p.params())
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// P2PKHFromKey returns the p2pkh address for a key
func (p *btcParams) P2PKHFromKey(pub []byte) (string, error) {
	return p.P2PKH(hash.Hash160(pub))
}

// P2SH returns the p2sh address for a script hash
func (p *btcParams) P2SH(scriptHash []byte) (string, error) {
	r, err := btcutil.NewAddressScriptHashFromHash(scriptHash, p.params())
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// P2SHFromScript returns the p2sh address for a script
func (p *btcParams) P2SHFromScript(script []byte) (string, error) {
	return p.P2SH(hash.Hash160(script))
}

var (
	// BTC_MainNet represents the bitcoin main net
	BTC_MainNet = &btcParams{
		pubKeyHashAddrID: 0x00, // starts with 1
		scriptHashAddrID: 0x05, // starts with 3
		privateKeyID:     0x80, // starts with 5 (uncompressed) or K (compressed)
	}
	// BTC_TestNet represents the bitcoin test net
	BTC_TestNet = &btcParams{
		pubKeyHashAddrID: 0x6f, // starts with m or n
		scriptHashAddrID: 0xc4, // starts with 2
		privateKeyID:     0xef, // starts with 9 (uncompressed) or c (compressed)
	}
	// BTC_RegressionNet represents the bitcoin regression test net
	BTC_RegressionNet = &btcParams{
		pubKeyHashAddrID: 0x6f, // starts with m or n
		scriptHashAddrID: 0xc4, // starts with 2
		privateKeyID:     0xef, // starts with 9 (uncompressed) or c (compressed)
	}
	// BTC_SimNet represents the bitcoin simulation net
	BTC_SimNet = &btcParams{
		pubKeyHashAddrID: 0x3f, // starts with S
		scriptHashAddrID: 0x7b, // starts with s
		privateKeyID:     0x64, // starts with 4 (uncompressed) or F (compressed)
	}
)

func init() {
	Networks["bitcoin"] = map[Chain]Params{
		MainNet:       BTC_MainNet,
		TestNet:       BTC_TestNet,
		SimNet:        BTC_SimNet,
		RegressionNet: BTC_RegressionNet,
	}
}
