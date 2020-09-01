package flagutil

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
	"github.com/transmutate-io/atomicswap/internal/cmdutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil/exitcodes"
	"github.com/transmutate-io/atomicswap/internal/testutil"
	"github.com/transmutate-io/atomicswap/params"
	"github.com/transmutate-io/cryptocore"
)

func init() { cobra.EnableCommandSorting = false }

type FlagFunc = func(*pflag.FlagSet)

type FlagFuncMap map[*pflag.FlagSet][]FlagFunc

func AddFlags(fm FlagFuncMap) {
	for fs, flags := range fm {
		for _, i := range flags {
			i(fs)
		}
	}
}

func Count(fs *pflag.FlagSet, name string) (int, error) { return fs.GetCount(name) }

func MustCount(fs *pflag.FlagSet, name string) int {
	r, err := Count(fs, name)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.CantGetFlag, err)
	}
	return r
}

func String(fs *pflag.FlagSet, name string) (string, error) { return fs.GetString(name) }

func MustString(fs *pflag.FlagSet, name string) string {
	r, err := String(fs, name)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.CantGetFlag, err)
	}
	return r
}

func Duration(fs *pflag.FlagSet, name string) (time.Duration, error) { return fs.GetDuration(name) }

func MustDuration(fs *pflag.FlagSet, name string) time.Duration {
	r, err := Duration(fs, name)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.CantGetFlag, err)
	}
	return r
}

func Bool(fs *pflag.FlagSet, name string) (bool, error) { return fs.GetBool(name) }

func MustBool(fs *pflag.FlagSet, name string) bool {
	r, err := Bool(fs, name)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.CantGetFlag, err)
	}
	return r
}

func UInt64(fs *pflag.FlagSet, name string) (uint64, error) { return fs.GetUint64(name) }

func MustUInt64(fs *pflag.FlagSet, name string) uint64 {
	r, err := UInt64(fs, name)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.CantGetFlag, err)
	}
	return r
}

func AddVerbose(fs *pflag.FlagSet)           { fs.CountP("verbose", "v", "increse verbose level") }
func Verbose(fs *pflag.FlagSet) (int, error) { return Count(fs, "verbose") }
func MustVerbose(fs *pflag.FlagSet) int      { return MustCount(fs, "verbose") }

func VerboseLevel(fs *pflag.FlagSet, max int) (int, error) {
	v, err := Verbose(fs)
	if err != nil {
		return 0, err
	}
	return verboseLevel(v, max), nil
}

func verboseLevel(vl, max int) int {
	if vl > max {
		return max
	}
	return vl
}

func MustVerboseLevel(fs *pflag.FlagSet, max int) int {
	return verboseLevel(MustVerbose(fs), max)
}

func AddFormat(fs *pflag.FlagSet) {
	fs.StringP("format", "f", "", "go template format string for output")
}

func Format(fs *pflag.FlagSet) (string, error) { return String(fs, "format") }
func MustFormat(fs *pflag.FlagSet) string      { return MustString(fs, "format") }

func AddForce(fs *pflag.FlagSet)            { fs.BoolP("force", "f", false, "force") }
func Force(fs *pflag.FlagSet) (bool, error) { return Bool(fs, "force") }
func MustForce(fs *pflag.FlagSet) bool      { return MustBool(fs, "force") }

func AddAll(fs *pflag.FlagSet)            { fs.BoolP("all", "a", false, "all") }
func All(fs *pflag.FlagSet) (bool, error) { return Bool(fs, "all") }
func MustAll(fs *pflag.FlagSet) bool      { return MustBool(fs, "all") }

func AddConfirmations(fs *pflag.FlagSet) {
	fs.Uint64P("confirmations", "c", 0, "number of confirmations")
}

func Confirmations(fs *pflag.FlagSet) (uint64, error) { return UInt64(fs, "confirmations") }
func MustConfirmations(fs *pflag.FlagSet) uint64      { return MustUInt64(fs, "confirmations") }

func AddIgnoreTarget(fs *pflag.FlagSet) {
	fs.BoolP("ignoretarget", "t", false, "ignore target amount and continue watching")
}

func IgnoreTarget(fs *pflag.FlagSet) (bool, error) { return Bool(fs, "ignoretarget") }
func MustIgnoreTarget(fs *pflag.FlagSet) bool      { return MustBool(fs, "ignoretarget") }

func AddFirstBlock(fs *pflag.FlagSet) {
	fs.Uint64P("firstblock", "b", 1, "set the first block where is possible to find an input")
}

func FirstBlock(fs *pflag.FlagSet) (uint64, error) { return UInt64(fs, "firstblock") }
func MustFirstBlock(fs *pflag.FlagSet) uint64      { return MustUInt64(fs, "firstblock") }

func AddRPC(fs *pflag.FlagSet) {
	fs.StringP("rpcaddr", "a", "127.0.0.1:3333", "set RPC host:port")
	fs.StringP("rpcusername", "u", "admin", "set RPC username")
	fs.StringP("rpcpassword", "p", "password", "set RPC password")
	fs.String("rpctlscacert", "", "set RPC CA certificate")
	fs.Bool("rpctlsskipverify", false, "skip TLS verification")
	fs.String("rpctlsclientcert", "", "RPC client certificate")
	fs.String("rpctlsclientkey", "", "RPC client key")
}

func RPCUsername(fs *pflag.FlagSet) (string, error) { return String(fs, "rpcusername") }
func MustRPCUsername(fs *pflag.FlagSet) string      { return MustString(fs, "rpcusername") }
func RPCPassword(fs *pflag.FlagSet) (string, error) { return String(fs, "rpcpassword") }
func MustRPCPassword(fs *pflag.FlagSet) string      { return MustString(fs, "rpcpassword") }
func RPCAddress(fs *pflag.FlagSet) (string, error)  { return String(fs, "rpcaddr") }
func MustRPCAddress(fs *pflag.FlagSet) string       { return MustString(fs, "rpcaddr") }

func RPCTLSConfig(fs *pflag.FlagSet) (*cryptocore.TLSConfig, error) {
	var changed bool
	r := &cryptocore.TLSConfig{}
	if v, err := String(fs, "rpctlscacert"); err != nil {
		return nil, err
	} else if v != "" {
		changed = true
		r.CA = v
	}
	if v, err := String(fs, "rpctlsclientcert"); err != nil {
		return nil, err
	} else if v != "" {
		changed = true
		r.ClientCertificate = v
	}
	if v, err := String(fs, "rpctlsclientkey"); err != nil {
		return nil, err
	} else if v != "" {
		changed = true
		r.ClientKey = v
	}
	if v, err := Bool(fs, "rpctlsskipverify"); err != nil {
		return nil, err
	} else if v {
		changed = true
		r.SkipVerify = v
	}
	if changed {
		return r, nil
	}
	return nil, nil
}

func MustRPCTLSConfig(fs *pflag.FlagSet) *cryptocore.TLSConfig {
	r, err := RPCTLSConfig(fs)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.CantGetFlag, err)
	}
	return r
}

type FeeFlag struct {
	Value uint64
	Fixed bool
}

func (f *FeeFlag) Set(v string) error {
	if strings.HasSuffix(v, "b") {
		f.Fixed = false
		v = strings.TrimSuffix(v, "b")
	} else {
		f.Fixed = false
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return err
	}
	f.Value = uint64(i)
	return nil
}

func (f *FeeFlag) String() string {
	r := strconv.Itoa(int(f.Value))
	if !f.Fixed {
		r += "b"
	}
	return r
}

func (f *FeeFlag) Type() string              { return "string" }
func (f *FeeFlag) AddFlag(fs *pflag.FlagSet) { fs.VarP(f, "fee", "f", "set fee per byte") }

type NetworkFlag string

var availableNetworks = []string{"mainnet", "testnet", "localnet"}

func (c *NetworkFlag) Set(v string) error {
	for _, i := range availableNetworks {
		if v == i {
			*c = NetworkFlag(v)
			return nil
		}
	}
	return errors.New(fmt.Sprintf("invalid network: %s", v))
}

func (c NetworkFlag) String() string { return string(c) }

func (c NetworkFlag) Type() string { return "string" }

func (c *NetworkFlag) AddFlag(fs *pflag.FlagSet) {
	fs.VarP(c, "network", "n", "set the network to use ("+strings.Join(availableNetworks, ", ")+")")
}

func (c NetworkFlag) Network(cn string) (params.Chain, error) {
	switch c {
	case "mainnet":
		return params.MainNet, nil
	case "testnet":
		return params.TestNet, nil
	case "localnet":
		for _, i := range testutil.Cryptos {
			if i.Name == cn {
				return i.Chain, nil
			}
		}
	default:
	}
	return 0, params.InvalidChainError(c)
}

func (c NetworkFlag) MustNetwork(cn string) params.Chain {
	r, err := c.Network(cn)
	if err != nil {
		panic(err)
	}
	return r
}

func AddInput(fs *pflag.FlagSet) { fs.StringP("input", "i", "-", "set input") }

func OpenInput(fs *pflag.FlagSet) (io.Reader, func() error, error) {
	fn, err := String(fs, "input")
	if err != nil {
		return nil, nil, err
	}
	if fn == "-" {
		return os.Stdin, func() error { return nil }, nil
	}
	f, err := os.Open(fn)
	if err != nil {
		return nil, nil, err
	}
	return f, func() error { return f.Close() }, nil
}

func MustOpenInput(fs *pflag.FlagSet) (io.Reader, func() error) {
	r, c, err := OpenInput(fs)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.BadInput, err)
	}
	return r, c
}

func AddOutput(fs *pflag.FlagSet) { fs.StringP("output", "o", "-", "set output") }

func OpenOutput(fs *pflag.FlagSet) (io.Writer, func() error, error) {
	fn, err := String(fs, "output")
	if err != nil {
		return nil, nil, err
	}
	if fn == "-" {
		return os.Stdout, func() error { return nil }, nil
	}
	f, err := os.Create(fn)
	if err != nil {
		return nil, nil, err
	}
	return f, func() error { return f.Close() }, nil
}

func MustOpenOutput(fs *pflag.FlagSet) (io.Writer, func() error) {
	w, c, err := OpenOutput(fs)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.BadOutput, err)
	}
	return w, c
}
