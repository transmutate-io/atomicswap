package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"transmutate.io/pkg/atomicswap/hash"
)

func errorExit(code int, f string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, f, a...)
	os.Exit(code)
}

type hashFunc func([]byte) []byte

var (
	hashFuncs = map[string]hashFunc{
		"sha256":      hash.Sha256Sum,
		"blake256":    hash.Blake256Sum,
		"ripemd160":   hash.Ripemd160Sum,
		"btc_hash256": hash.NewBTC().Hash256,
		"btc_hash160": hash.NewBTC().Hash160,
		"dcr_hash256": hash.NewDCR().Hash256,
		"dcr_hash160": hash.NewDCR().Hash160,
	}
	hashFuncsKeys []string
)

func init() {
	hashFuncsKeys = make([]string, 0, len(hashFuncs))
	for i := range hashFuncs {
		hashFuncsKeys = append(hashFuncsKeys, i)
	}
	sort.Strings(hashFuncsKeys)
}

func main() {
	if len(os.Args) < 2 {
		s := append(make([]string, 0, len(hashFuncs)+1), "expecting hash type\navailable:")
		for _, i := range hashFuncsKeys {
			s = append(s, "    "+i)
		}
		errorExit(-1, strings.Join(s, "\n"))
	}
	hf, ok := hashFuncs[strings.ToLower(os.Args[1])]
	if !ok {
		errorExit(-2, "invalid hash type: %s\n", os.Args[1])
	}
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	var (
		fName  string
		useArg bool
	)
	fs.BoolVar(&useArg, "x", false, "use provided hex string")
	fs.StringVar(&fName, "f", "-", "read from file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(-3)
	}
	if args := fs.Args(); useArg && len(args) < 1 {
		errorExit(-4, "need a hex string")
	} else if useArg {
		for _, i := range args {
			b, err := hex.DecodeString(i)
			if err != nil {
				errorExit(-5, "can't decode hex string: %v\n", err)
			}
			fmt.Fprintln(os.Stdout, hex.EncodeToString(hf(b)))
		}
		return
	}
	var fin io.Reader
	if fName == "-" {
		fin = os.Stdin
	} else {
		f, err := os.Open(fName)
		if err != nil {
			errorExit(-6, "can't open file: %v\n", err)
		}
		defer f.Close()
		fin = f
	}
	b, err := ioutil.ReadAll(fin)
	if err != nil {
		errorExit(-7, "can't read: %v\n", err)
	}
	b = hf(b)
	fmt.Fprintln(os.Stdout, hex.EncodeToString(b))
}
