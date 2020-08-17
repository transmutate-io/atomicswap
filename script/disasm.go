package script

import (
	"strings"

	"github.com/transmutate-io/atomicswap/cryptos"
)

// Disassembler represents a script disassembler
type Disassembler interface {
	DisassembleString(s []byte) (string, error)
}

// DisassembleString disassembles a script into a string
func DisassembleString(c *cryptos.Crypto, s []byte) (string, error) {
	dis, ok := disassemblers[c.Name]
	if !ok {
		return "", cryptos.InvalidCryptoError(c.Name)
	}
	return dis.DisassembleString(s)
}

// DisassembleStrings disassembles a script into multiple strings
func DisassembleStrings(c *cryptos.Crypto, s []byte) ([]string, error) {
	r, err := DisassembleString(c, s)
	if err != nil {
		return nil, err
	}
	return strings.Split(r, " "), nil
}
