package params

import (
	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/dcrec"
	"github.com/decred/dcrd/dcrutil"
	"github.com/transmutate-io/atomicswap/hash"
)

// dcrParams represents a network parameter set
type dcrParams struct {
	pubKeyHashAddrID [2]byte
	scriptHashAddrID [2]byte
	privateKeyID     [2]byte
}

// Params returns the chain params as a *chaincfg.Params
func (p *dcrParams) params() *chaincfg.Params {
	return &chaincfg.Params{
		PubKeyHashAddrID: p.pubKeyHashAddrID,
		ScriptHashAddrID: p.scriptHashAddrID,
		PrivateKeyID:     p.privateKeyID,
	}
}

// P2PK returns the p2pk address for a key (same as p2pkh)
func (p *dcrParams) P2PK(pub []byte) (string, error) {
	r, err := dcrutil.NewAddressPubKey(pub, p.params())
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// P2PKH returns the p2pkh address for a key hash
func (p *dcrParams) P2PKH(pubHash []byte) (string, error) {
	r, err := dcrutil.NewAddressPubKeyHash(pubHash, p.params(), dcrec.STEcdsaSecp256k1)
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// P2PKHFromKey returns the p2pkh address for a key
func (p *dcrParams) P2PKHFromKey(pub []byte) (string, error) {
	return p.P2PKH(hash.NewDCR().Hash160(pub))
}

// P2SH returns the p2sh address for a script hash
func (p *dcrParams) P2SH(scriptHash []byte) (string, error) {
	r, err := dcrutil.NewAddressScriptHashFromHash(scriptHash, p.params())
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// P2SHFromScript returns the p2sh address for a script
func (p *dcrParams) P2SHFromScript(script []byte) (string, error) {
	return p.P2SH(hash.NewDCR().Hash160(script))
}

var (
	// DCR_MainNet represents the bitcoin main net
	DCR_MainNet = &dcrParams{
		pubKeyHashAddrID: [2]byte{0x07, 0x3f},
		scriptHashAddrID: [2]byte{0x07, 0x1a},
		privateKeyID:     [2]byte{0x22, 0xde},
	}
	// DCR_TestNet represents the bitcoin test net
	DCR_TestNet = &dcrParams{
		pubKeyHashAddrID: [2]byte{0x0f, 0x21},
		scriptHashAddrID: [2]byte{0x0e, 0xfc},
		privateKeyID:     [2]byte{0x23, 0x0e},
	}
	// DCR_RegressionNet represents the bitcoin regression test net
	DCR_RegressionNet = &dcrParams{
		pubKeyHashAddrID: [2]byte{0x0e, 0x00},
		scriptHashAddrID: [2]byte{0x0d, 0xdb},
		privateKeyID:     [2]byte{0x22, 0xfe},
	}
	// DCR_SimNet represents the bitcoin simulation net
	DCR_SimNet = &dcrParams{
		pubKeyHashAddrID: [2]byte{0x0e, 0x91},
		scriptHashAddrID: [2]byte{0x0e, 0x6c},
		privateKeyID:     [2]byte{0x23, 0x07},
	}
)
