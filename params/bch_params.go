package params

import (
	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchutil"
	"transmutate.io/pkg/atomicswap/hash"
)

// bchParams represents a network parameter set
type bchParams struct {
	prefix           string
	pubKeyHashAddrID byte // First byte of a P2PKH address
	scriptHashAddrID byte // First byte of a P2SH address
	privateKeyID     byte // First byte of a WIF private key
}

// bchParams returns the chain params as a *chaincfg.bchParams
func (p *bchParams) params() *chaincfg.Params {
	return &chaincfg.Params{
		CashAddressPrefix:      p.prefix,
		LegacyPubKeyHashAddrID: p.pubKeyHashAddrID,
		LegacyScriptHashAddrID: p.scriptHashAddrID,
		PrivateKeyID:           p.privateKeyID,
	}
}

// P2PK returns the p2pk address for a key (same as p2pkh)
func (p *bchParams) P2PK(pub []byte) (string, error) {
	r, err := bchutil.NewAddressPubKey(pub, p.params())
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// P2PKH returns the p2pkh address for a key hash
func (p *bchParams) P2PKH(pubHash []byte) (string, error) {
	r, err := bchutil.NewAddressPubKeyHash(pubHash, p.params())
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// P2PKHFromKey returns the p2pkh address for a key
func (p *bchParams) P2PKHFromKey(pub []byte) (string, error) {
	return p.P2PKH(hash.Hash160(pub))
}

// P2SH returns the p2sh address for a script hash
func (p *bchParams) P2SH(scriptHash []byte) (string, error) {
	r, err := bchutil.NewAddressScriptHashFromHash(scriptHash, p.params())
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// P2SHFromScript returns the p2sh address for a script
func (p *bchParams) P2SHFromScript(script []byte) (string, error) {
	return p.P2SH(hash.Hash160(script))
}

var (
	// BCH_MainNet represents the bitcoin main net
	BCH_MainNet = &bchParams{
		prefix:           "bitcoincash",
		pubKeyHashAddrID: 0x00, // starts with 1
		scriptHashAddrID: 0x05, // starts with 3
		privateKeyID:     0x80, // starts with 5 (uncompressed) or K (compressed)
	}
	// BCH_TestNet represents the bitcoin test net
	BCH_TestNet = &bchParams{
		prefix:           "bchtest",
		pubKeyHashAddrID: 0x6f,
		scriptHashAddrID: 0xc4,
		privateKeyID:     0xef,
	}
	// BCH_RegressionNet represents the bitcoin regression test net
	BCH_RegressionNet = &bchParams{
		prefix:           "bchreg",
		pubKeyHashAddrID: 0x6f,
		scriptHashAddrID: 0xc4,
		privateKeyID:     0xef,
	}
	// BCH_SimNet represents the bitcoin simulation net
	BCH_SimNet = &bchParams{
		prefix:           "bchsim",
		pubKeyHashAddrID: 0x3f,
		scriptHashAddrID: 0x7b,
		privateKeyID:     0x64,
	}
)
