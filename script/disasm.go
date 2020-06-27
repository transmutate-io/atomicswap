package script

import (
	"strings"

	"transmutate.io/pkg/atomicswap/cryptos"
)

type Disassembler interface {
	DisassembleString(s []byte) (string, error)
}

func DisassembleString(c *cryptos.Crypto, s []byte) (string, error) {
	dis, ok := disassemblers[c.Name]
	if !ok {
		return "", cryptos.InvalidCryptoError(c.Name)
	}
	return dis.DisassembleString(s)
}

func DisassembleStrings(c *cryptos.Crypto, s []byte) ([]string, error) {
	r, err := DisassembleString(c, s)
	if err != nil {
		return nil, err
	}
	return strings.Split(r, " "), nil
}
