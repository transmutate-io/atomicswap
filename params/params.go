package params

type Params interface {
	P2PK(pub []byte) (string, error)
	P2PKH(pubHash []byte) (string, error)
	P2PKHFromKey(pub []byte) (string, error)
	P2SH(scriptHash []byte) (string, error)
	P2SHFromScript(script []byte) (string, error)
	AddressToScript(addr string) ([]byte, error)
}
