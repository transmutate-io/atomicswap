package params

// Params represents the parameters of a utxo crypto
type Params interface {
	// P2PK returns the p2pk address of the public key
	P2PK(pub []byte) (string, error)
	// P2PKH returns the p2pkh address of the public key hash
	P2PKH(pubHash []byte) (string, error)
	// P2PKHFromKey returns the p2pkh address of the public key
	P2PKHFromKey(pub []byte) (string, error)
	// P2SH returns the p2sh address of the script hash
	P2SH(scriptHash []byte) (string, error)
	// P2SHFromScript returns the p2sh address of the script
	P2SHFromScript(script []byte) (string, error)
	// AddressToScript converts an addres to a script
	AddressToScript(addr string) ([]byte, error)
}
