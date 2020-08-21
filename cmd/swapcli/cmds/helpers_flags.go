package cmds

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/internal/testutil"
	"github.com/transmutate-io/atomicswap/params"
	"github.com/transmutate-io/cryptocore"
)

func init() { cobra.EnableCommandSorting = false }

type flagFunc = func(*pflag.FlagSet)

type flagMap map[*pflag.FlagSet][]flagFunc

func addFlags(fm flagMap) {
	for fs, flags := range fm {
		for _, i := range flags {
			i(fs)
		}
	}
}

func addCommands(cmd *cobra.Command, sub []*cobra.Command) {
	for _, i := range sub {
		cmd.AddCommand(i)
	}
}

func flagCount(fs *pflag.FlagSet, name string) (int, error) { return fs.GetCount(name) }

func mustFlagCount(fs *pflag.FlagSet, name string) int {
	r, err := flagCount(fs, name)
	if err != nil {
		errorExit(ecCantGetFlag, err)
	}
	return r
}

func flagString(fs *pflag.FlagSet, name string) (string, error) { return fs.GetString(name) }

func mustFlagString(fs *pflag.FlagSet, name string) string {
	r, err := flagString(fs, name)
	if err != nil {
		errorExit(ecCantGetFlag, err)
	}
	return r
}

func flagDuration(fs *pflag.FlagSet, name string) (time.Duration, error) { return fs.GetDuration(name) }

func mustFlagDuration(fs *pflag.FlagSet, name string) time.Duration {
	r, err := fs.GetDuration(name)
	if err != nil {
		errorExit(ecCantGetFlag, err)
	}
	return r
}

func flagBool(fs *pflag.FlagSet, name string) (bool, error) { return fs.GetBool(name) }

func mustFlagBool(fs *pflag.FlagSet, name string) bool {
	r, err := fs.GetBool(name)
	if err != nil {
		errorExit(ecCantGetFlag, err)
	}
	return r
}

func flagUInt64(fs *pflag.FlagSet, name string) (uint64, error) { return fs.GetUint64(name) }

func mustFlagUInt64(fs *pflag.FlagSet, name string) uint64 {
	r, err := fs.GetUint64(name)
	if err != nil {
		errorExit(ecCantGetFlag, err)
	}
	return r
}

func addFlagVerbose(fs *pflag.FlagSet) { fs.CountP("verbose", "v", "increse verbose level") }

func verboseLevel(vl, max int) int {
	if vl > max {
		return max
	}
	return vl
}

func mustVerboseLevel(fs *pflag.FlagSet, max int) int {
	return verboseLevel(mustFlagCount(fs, "verbose"), max)
}

func addFlagFormat(fs *pflag.FlagSet) {
	fs.StringP("format", "f", "", "go template format string for output")
}

func mustFlagFormat(fs *pflag.FlagSet) string { return mustFlagString(fs, "format") }

func addFlagForce(fs *pflag.FlagSet) { fs.BoolP("force", "f", false, "force") }

func mustFlagForce(fs *pflag.FlagSet) bool { return mustFlagBool(fs, "force") }

func addFlagAll(fs *pflag.FlagSet) { fs.BoolP("all", "a", false, "all") }

func mustFlagAll(fs *pflag.FlagSet) bool { return mustFlagBool(fs, "all") }

func addFlagConfirmations(fs *pflag.FlagSet) {
	fs.Uint64P("confirmations", "c", 0, "number of confirmations")
}

func mustFlagConfirmations(fs *pflag.FlagSet) uint64 { return mustFlagUInt64(fs, "confirmations") }

func addFlagIgnoreTarget(fs *pflag.FlagSet) {
	fs.BoolP("ignoretarget", "t", false, "ignore target amount and continue watching")
}

func mustFlagIgnoreTarget(fs *pflag.FlagSet) bool { return mustFlagBool(fs, "ignoretarget") }

func addFlagFirstBlock(fs *pflag.FlagSet) {
	fs.Uint64P("firstblock", "b", 1, "set the first block where is possible to find an input")
}

func flagFirstBlock(fs *pflag.FlagSet) uint64 { return mustFlagUInt64(fs, "firstblock") }

func addFlagsRPC(fs *pflag.FlagSet) {
	fs.StringP("rpcaddr", "a", "127.0.0.1:3333", "set RPC host:port")
	fs.StringP("rpcusername", "u", "admin", "set RPC username")
	fs.StringP("rpcpassword", "p", "password", "set RPC password")
	fs.String("rpctlscacert", "", "set RPC CA certificate")
	fs.Bool("rpctlsskipverify", false, "skip TLS verification")
	fs.String("rpctlsclientcert", "", "RPC client certificate")
	fs.String("rpctlsclientkey", "", "RPC client key")
}

func mustFlagRPCUsername(fs *pflag.FlagSet) string { return mustFlagString(fs, "rpcusername") }

func mustFlagRPCPassword(fs *pflag.FlagSet) string { return mustFlagString(fs, "rpcpassword") }

func mustFlagRPCAddress(fs *pflag.FlagSet) string { return mustFlagString(fs, "rpcaddr") }

func mustFlagRPCTLSConfig(fs *pflag.FlagSet) *cryptocore.TLSConfig {
	var changed bool
	r := &cryptocore.TLSConfig{}
	if s := mustFlagString(fs, "rpctlscacert"); s != "" {
		changed = true
		r.CA = s
	}
	if s := mustFlagString(fs, "rpctlsclientcert"); s != "" {
		changed = true
		r.ClientCertificate = s
	}
	if s := mustFlagString(fs, "rpctlsclientkey"); s != "" {
		changed = true
		r.ClientKey = s
	}
	if s := mustFlagBool(fs, "rpctlsskipverify"); s {
		changed = true
		r.SkipVerify = s
	}
	if changed {
		return r
	}
	return nil
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

func addFlagFee(fs *pflag.FlagSet) { fs.VarP(_fee, "fee", "f", "set fee per byte") }

func flagFee(fs *pflag.FlagSet) uint64    { return _fee.val }
func flagFeeFixed(fs *pflag.FlagSet) bool { return _fee.fixed }

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

func (c cryptoChain) Type() string { return "string" }

var cryptoChains = []string{"mainnet", "testnet", "localnet"}

var _cryptoChain = cryptoChain(cryptoChains[0])

func addFlagCryptoChain(fs *pflag.FlagSet) {
	fs.VarP(&_cryptoChain, "network", "n", "set the network to use ("+strings.Join(cryptoChains, ", ")+")")
}

func mustFlagCryptoChain(c *cryptos.Crypto) params.Chain {
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
	panic("unknown network")
}

func addFlagInput(fs *pflag.FlagSet) { fs.StringP("input", "i", "-", "set input") }

func openInput(infn string) (io.Reader, func() error, error) {
	if infn == "-" {
		return os.Stdin, func() error { return nil }, nil
	}
	f, err := os.Open(infn)
	if err != nil {
		return nil, nil, err
	}
	return f, func() error { return f.Close() }, nil
}

func mustOpenInput(fs *pflag.FlagSet) (io.Reader, func() error) {
	r, c, err := openInput(mustFlagString(fs, "input"))
	if err != nil {
		errorExit(ecNoInput, err)
	}
	return r, c
}

func addFlagOutput(fs *pflag.FlagSet) { fs.StringP("output", "o", "-", "set output") }

func openOutput(outfn string) (io.Writer, func() error, error) {
	if outfn == "-" {
		return os.Stdout, func() error { return nil }, nil
	}
	f, err := os.Create(outfn)
	if err != nil {
		return nil, nil, err
	}
	return f, func() error { return f.Close() }, nil
}

func mustOpenOutput(fs *pflag.FlagSet) (io.Writer, func() error) {
	w, c, err := openOutput(mustFlagString(fs, "output"))
	if err != nil {
		errorExit(ecCantOpenOutput, err)
	}
	return w, c
}

// func flagConfigFile(fs *pflag.FlagSet) (string, error) { return flagString(fs, "config") }
// func mustFlagConfigFile(fs *pflag.FlagSet) string      { return mustFlagString(fs, "config") }

// func configDir(cmd *cobra.Command) string { return filepath.Join(dataDir(cmd), "config") }

// func configPath(cmd *cobra.Command, name string) string {
// 	return filepath.Join(configDir(cmd), name)
// }

// func loadConfig(cmd *cobra.Command) (*config.Config, error) {
// 	configFile, err := flagConfigFile(cmd.Flags())
// 	if err != nil {
// 		return nil, err
// 	}
// 	if configFile == "" {
// 		configFile = configPath(cmd, "default.yaml")
// 	}
// 	return config.LoadConfig(configFile)
// }

// func mustLoadConfig(cmd *cobra.Command)*config.Config{
// 	r,err:=loadConfig(cmd)
// }
