package script

import (
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
	r, _ := txscript.NewScriptBuilder().AddData(b).Script()
	return r
}

// Int64 add an int64 as data
func Int64(n int64) []byte {
	b, _ := txscript.NewScriptBuilder().AddInt64(n).Script()
	return b
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

func ParseInt64(v []byte) (int64, error) {
	if len(v) == 0 {
		return 0, nil
	}
	var result int64
	for i, val := range v {
		result |= int64(val) << uint8(8*i)
	}
	if v[len(v)-1]&0x80 != 0 {
		result &= ^(int64(0x80) << uint8(8*(len(v)-1)))
		return -result, nil
	}
	return result, nil
}
