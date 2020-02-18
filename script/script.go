package script

import (
	"encoding/binary"
	"strings"

	"github.com/btcsuite/btcd/txscript"
)

// Validate a script and return it or an error
func Validate(b []byte) ([]byte, error) {
	return txscript.NewScriptBuilder().AddOps(b).Script()
}

// If else statement. If e is nil a els branch will not be present
func If(i, e []byte) []byte {
	r := append(make([]byte, 0, len(i)+len(e)+3), txscript.OP_IF)
	r = append(r, i...)
	if e != nil {
		r = append(r, txscript.OP_ELSE)
		r = append(r, e...)
	}
	return append(r, txscript.OP_ENDIF)
}

// Data adds bytes as data
func Data(b []byte) []byte {
	r := make([]byte, 0, len(b)+8)
	sz := len(b)
	if sz == 0 || sz == 1 && b[0] == 0 {
		return append(r, txscript.OP_0)
	} else if sz == 1 && b[0] <= 16 {
		return append(r, (txscript.OP_1-1)+b[0])
	} else if sz == 1 && b[0] == 0x81 {
		return append(r, byte(txscript.OP_1NEGATE))
	}
	if sz < txscript.OP_PUSHDATA1 {
		r = append(r, byte((txscript.OP_DATA_1-1)+sz))
	} else if sz <= 0xff {
		r = append(r, txscript.OP_PUSHDATA1, byte(sz))
	} else if sz <= 0xffff {
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, uint16(sz))
		r = append(r, txscript.OP_PUSHDATA2)
		r = append(r, buf...)
	} else {
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(sz))
		r = append(r, txscript.OP_PUSHDATA4)
		r = append(r, buf...)
	}
	return append(r, b...)
}

// Int64 add an int64 as data
func Int64(i int64) []byte {
	if i == 0 {
		return []byte{txscript.OP_0}
	}
	if i == -1 || (i >= 1 && i <= 16) {
		return []byte{byte((txscript.OP_1 - 1) + i)}
	}
	return Data(numBytes(i))
}

func numBytes(n int64) []byte {
	if n == 0 {
		return nil
	}
	isNegative := n < 0
	if isNegative {
		n = -n
	}
	result := make([]byte, 0, 9)
	for n > 0 {
		result = append(result, byte(n&0xff))
		n >>= 8
	}
	if result[len(result)-1]&0x80 != 0 {
		extraByte := byte(0x00)
		if isNegative {
			extraByte = 0x80
		}
		result = append(result, extraByte)

	} else if isNegative {
		result[len(result)-1] |= 0x80
	}
	return result
}

// Disassemble a script into string
func DisassembleString(s []byte) (string, error) { return txscript.DisasmString(s) }

func DisassembleStrings(s []byte) ([]string, error) {
	r, err := DisassembleString(s)
	if err != nil {
		return nil, err
	}
	return strings.Split(r, " "), nil
}
