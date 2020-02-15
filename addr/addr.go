package addr

import (
	"github.com/btcsuite/btcutil"
	"transmutate.io/pkg/swapper/hash"
	"transmutate.io/pkg/swapper/params"
)

// P2PK returns the p2pk address for a key (same as p2pkh)
func P2PK(pub []byte, net *params.Params) (string, error) {
	r, err := btcutil.NewAddressPubKey(pub, net.Params())
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// P2PKH returns the p2pkh address for a key hash
func P2PKH(pubHash []byte, net *params.Params) (string, error) {
	r, err := btcutil.NewAddressPubKeyHash(pubHash, net.Params())
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// P2PKHFromKey returns the p2pkh address for a key
func P2PKHFromKey(pub []byte, net *params.Params) (string, error) {
	return P2PKH(hash.Hash160(pub), net)
}

// P2SH returns the p2sh address for a script hash
func P2SH(scriptHash []byte, net *params.Params) (string, error) {
	r, err := btcutil.NewAddressScriptHashFromHash(scriptHash, net.Params())
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// P2SHFromScript returns the p2sh address for a script
func P2SHFromScript(script []byte, net *params.Params) (string, error) {
	return P2SH(hash.Hash160(script), net)
}
