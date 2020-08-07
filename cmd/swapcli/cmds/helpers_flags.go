package cmds

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/pflag"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/internal/testutil"
	"github.com/transmutate-io/atomicswap/params"
	"github.com/transmutate-io/cryptocore"
)

func flagCount(fs *pflag.FlagSet, name string) int {
	r, err := fs.GetCount(name)
	if err != nil {
		errorExit(ecCantGetFlag, err)
	}
	return r
}

func flagString(fs *pflag.FlagSet, name string) string {
	r, err := fs.GetString(name)
	if err != nil {
		errorExit(ecCantGetFlag, err)
	}
	return r
}

func flagDuration(fs *pflag.FlagSet, name string) time.Duration {
	r, err := fs.GetDuration(name)
	if err != nil {
		errorExit(ecCantGetFlag, err)
	}
	return r
}

func flagBool(fs *pflag.FlagSet, name string) bool {
	r, err := fs.GetBool(name)
	if err != nil {
		errorExit(ecCantGetFlag, err)
	}
	return r
}

func flagUInt64(fs *pflag.FlagSet, name string) uint64 {
	r, err := fs.GetUint64(name)
	if err != nil {
		errorExit(ecCantGetFlag, err)
	}
	return r
}

type templateData = map[string]interface{}

func outputTemplate(fs *pflag.FlagSet, tpls []string, funcs template.FuncMap) *template.Template {
	var tplStr string
	if tplStr = flagFormat(fs); tplStr == "" {
		tplStr = tpls[verboseLevel(fs, len(tpls)-1)]
	}
	var err error
	r := template.New("main")
	if funcs != nil {
		r = r.Funcs(funcs)
	}
	if r, err = r.Parse(tplStr); err != nil {
		errorExit(ecBadTemplate, err)
	}
	return r
}

func addFlagOutput(fs *pflag.FlagSet) { fs.StringP("output", "o", "-", "set output") }

func openOutput(fs *pflag.FlagSet) (io.Writer, func() error) {
	outfn := flagString(fs, "output")
	if outfn == "-" {
		return os.Stdout, func() error { return nil }
	}
	f, err := os.Create(outfn)
	if err != nil {
		errorExit(ecCantOpenOutput, err)
	}
	return f, func() error { return f.Close() }
}

func addFlagInput(fs *pflag.FlagSet) { fs.StringP("input", "i", "-", "set input") }

func openInput(fs *pflag.FlagSet) (io.Reader, func() error) {
	infn := flagString(fs, "input")
	if infn == "-" {
		return os.Stdin, func() error { return nil }
	}
	f, err := os.Open(infn)
	if err != nil {
		errorExit(ecNoInput, err)
	}
	return f, func() error { return f.Close() }
}

func addFlagVerbose(fs *pflag.FlagSet) { fs.CountP("verbose", "v", "increse verbose level") }

func verboseLevel(fs *pflag.FlagSet, max int) int {
	r := flagCount(fs, "verbose")
	if r > max {
		return max
	}
	return r
}

type cryptoChain string

func (c *cryptoChain) Set(v string) error {
	for _, i := range cryptoChains {
		if v == i {
			*c = cryptoChain(v)
			return nil
		}
	}
	return errors.New(fmt.Sprintf("invalid network: %s", v))
}

func (c cryptoChain) String() string { return string(c) }
func (c cryptoChain) Type() string   { return "string" }

var cryptoChains = []string{"mainnet", "testnet", "localnet"}

var _cryptoChain = cryptoChain(cryptoChains[0])

func addFlagCryptoChain(fs *pflag.FlagSet) {
	fs.VarP(&_cryptoChain, "network", "n", "set the network to use ("+strings.Join(cryptoChains, ", ")+")")
}

func flagCryptoChain(c *cryptos.Crypto) params.Chain {
	switch _cryptoChain {
	case "mainnet":
		return params.MainNet
	case "testnet":
		return params.TestNet
	}
	for _, i := range testutil.Cryptos {
		if i.Name == c.Name {
			return i.Chain
		}
	}
	return params.MainNet
}

func flagFormat(fs *pflag.FlagSet) string { return flagString(fs, "format") }

func addFlagFormat(fs *pflag.FlagSet) {
	fs.StringP("format", "f", "", "go template format string for output")
}

func flagForce(fs *pflag.FlagSet) bool { return flagBool(fs, "force") }

func addFlagForce(fs *pflag.FlagSet) { fs.BoolP("force", "f", false, "force") }

func flagAll(fs *pflag.FlagSet) bool { return flagBool(fs, "all") }

func addFlagAll(fs *pflag.FlagSet) { fs.BoolP("all", "a", false, "all") }

func flagRPCUsername(fs *pflag.FlagSet) string { return flagString(fs, "rpcusername") }
func flagRPCPassword(fs *pflag.FlagSet) string { return flagString(fs, "rpcpassword") }
func flagRPCAddress(fs *pflag.FlagSet) string  { return flagString(fs, "rpcaddr") }

func flagRPCTLSConfig(fs *pflag.FlagSet) *cryptocore.TLSConfig {
	var changed bool
	r := &cryptocore.TLSConfig{}
	if s := flagString(fs, "rpctlscacert"); s != "" {
		changed = true
		r.CA = s
	}
	if s := flagString(fs, "rpctlsclientcert"); s != "" {
		changed = true
		r.ClientCertificate = s
	}
	if s := flagString(fs, "rpctlsclientkey"); s != "" {
		changed = true
		r.ClientKey = s
	}
	if s := flagBool(fs, "rpctlsskipverify"); s {
		changed = true
		r.SkipVerify = s
	}
	if changed {
		return r
	}
	return nil
}

func addFlagsRPC(fs *pflag.FlagSet) {
	fs.StringP("rpcaddr", "a", "127.0.0.1:3333", "set RPC host:port")
	fs.StringP("rpcusername", "u", "admin", "set RPC username")
	fs.StringP("rpcpassword", "p", "password", "set RPC password")
	fs.String("rpctlscacert", "", "set RPC CA certificate")
	fs.Bool("rpctlsskipverify", false, "skip TLS verification")
	fs.String("rpctlsclientcert", "", "RPC client certificate")
	fs.String("rpctlsclientkey", "", "RPC client key")
}

func flagFirstBlock(fs *pflag.FlagSet) uint64 { return flagUInt64(fs, "firstblock") }

func addFlagFirstBlock(fs *pflag.FlagSet) {
	fs.Uint64P("firstblock", "b", 1, "set the first block where is possible to find an input")
}

func flagConfirmations(fs *pflag.FlagSet) uint64 { return flagUInt64(fs, "confirmations") }

func addFlagConfirmations(fs *pflag.FlagSet) {
	fs.Uint64P("confirmations", "c", 0, "number of confirmations")
}

func flagIgnoreTarget(fs *pflag.FlagSet) bool { return flagBool(fs, "ignoretarget") }

func addFlagIgnoreTarget(fs *pflag.FlagSet) {
	fs.BoolP("ignoretarget", "t", false, "ignore target amount and continue watching")
}

type fee struct {
	val   uint64
	fixed bool
}

func (f *fee) Set(v string) error {
	if strings.HasSuffix(v, "b") {
		f.fixed = false
		v = strings.TrimSuffix(v, "b")
	} else {
		f.fixed = false
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return err
	}
	f.val = uint64(i)
	return nil
}

func (f *fee) String() string {
	r := strconv.Itoa(int(f.val))
	if !f.fixed {
		r += "b"
	}
	return r
}

func (f *fee) Type() string { return "string" }

var _fee = &fee{val: 1, fixed: true}

func flagFee(fs *pflag.FlagSet) uint64    { return _fee.val }
func flagFeeFixed(fs *pflag.FlagSet) bool { return _fee.fixed }

func addFlagFee(fs *pflag.FlagSet) { fs.VarP(_fee, "fee", "f", "set fee per byte") }
