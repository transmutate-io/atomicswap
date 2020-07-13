package cmds

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/stages"
	"github.com/transmutate-io/atomicswap/trade"
	"gopkg.in/yaml.v2"
)

const (
	ECNoOutput           = -1
	ECCantOpenOutput     = -2
	ECCantGetFlag        = -3
	ECUnknownShell       = -4
	ECUnknownCrypto      = -5
	ECInvalidDuration    = -6
	ECCantCreateTrade    = -7
	ECCantFindTrade      = -8
	ECBadTemplate        = -9
	ECCantListTrades     = -10
	ECCantExportTrades   = -11
	ECCantExportProposal = -12
	ECCantListProposal   = -13
	ECCantOpenTrade      = -14
	ECCantOpenProposal   = -15
	ECInvalidProposal    = -16
	ECCantImportTrades   = -17
	ECNoInput            = -18
)

func openOutput(cmd *cobra.Command) (io.Writer, func() error) {
	outfn, err := cmd.Root().PersistentFlags().GetString("output")
	if err != nil {
		errorExit(ECNoOutput, "can't get output: %#v\n", err)
	}
	if outfn == "-" {
		return os.Stdout, func() error { return nil }
	}
	f, err := os.Create(outfn)
	if err != nil {
		errorExit(ECCantOpenOutput, "can't create output file: %#v\n", err)
	}
	return f, func() error { return f.Close() }
}

func openInput(cmd *cobra.Command) (io.Reader, func() error) {
	infn, err := cmd.Root().PersistentFlags().GetString("input")
	if err != nil {
		errorExit(ECNoInput, "can't get input: %#v\n", err)
	}
	if infn == "-" {
		return os.Stdin, func() error { return nil }
	}
	f, err := os.Open(infn)
	if err != nil {
		errorExit(ECNoInput, "can't open input file: %#v\n", err)
	}
	return f, func() error { return f.Close() }
}

func errorExit(code int, f string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, f, a...)
	os.Exit(code)
}

func dataDir(cmd *cobra.Command) string {
	return filepath.Clean(flagString(cmd.Root().PersistentFlags(), "data"))
}

func tradesDir(dataDir string) string { return filepath.Join(dataDir, "trades") }

func createFile(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return nil, err
	}
	return os.Create(p)
}

func verboseLevel(fs *pflag.FlagSet, max int) int {
	r, err := fs.GetCount("verbose")
	if err != nil {
		errorExit(ECCantGetFlag, "can't get flag: %#v\n", err)
	}
	if r > max {
		return max
	}
	return r
}

func parseCrypto(c string) (*cryptos.Crypto, error) {
	if r, err := cryptos.ParseShort(c); err == nil {
		return r, nil
	}
	r, err := cryptos.Parse(c)
	if err != nil {
		return nil, err
	}
	return r, nil
}

var filepathSeparator = string([]rune{filepath.Separator})

func eachTrade(td string, f func(string, trade.Trade) error) error {
	tdPrefix := td + filepathSeparator
	return filepath.Walk(td, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}
		tf, err := os.Open(path)
		if err != nil {
			return err
		}
		defer tf.Close()
		tr := &trade.OnChainTrade{}
		if err = yaml.NewDecoder(tf).Decode(tr); err != nil {
			return err
		}
		return f(strings.TrimPrefix(path, tdPrefix), tr)
	})
}

func eachProposal(td string, f func(string, trade.Trade) error) error {
	return eachTrade(td, func(name string, tr trade.Trade) error {
		if tr.Stager().Stage() != stages.SendProposal {
			return nil
		}
		return f(name, tr)
	})
}

func eachLockSet(td string, f func(string, trade.Trade) error) error {
	return eachTrade(td, func(name string, tr trade.Trade) error {
		if tr.Stager().Stage() != stages.ReceiveProposalResponse {
			return nil
		}
		return f(name, tr)
	})
}

func openTrade(cmd *cobra.Command, name string) (trade.Trade, error) {
	f, err := os.Open(filepath.Join(tradesDir(dataDir(cmd)), name))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := &trade.OnChainTrade{}
	if err = yaml.NewDecoder(f).Decode(r); err != nil {
		return nil, err
	}
	return r, nil
}

func saveTrade(cmd *cobra.Command, name string, tr trade.Trade) error {
	f, err := createFile(filepath.Join(tradesDir(dataDir(cmd)), name))
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewEncoder(f).Encode(tr)
}

func flagString(fs *pflag.FlagSet, name string) string {
	r, err := fs.GetString(name)
	if err != nil {
		errorExit(ECCantGetFlag, "can't get flag: %#v\n", err)
	}
	return r
}

func flagDuration(fs *pflag.FlagSet, name string) time.Duration {
	r, err := fs.GetDuration(name)
	if err != nil {
		errorExit(ECCantGetFlag, "can't get flag: %#v\n", err)
	}
	return r
}

func flagBool(fs *pflag.FlagSet, name string) bool {
	r, err := fs.GetBool(name)
	if err != nil {
		errorExit(ECCantGetFlag, "can't get flag: %#v\n", err)
	}
	return r
}

func addFlagVerbose(fs *pflag.FlagSet) {
	fs.CountP("verbose", "v", "increse verbose level")
}

func addFlagFormat(fs *pflag.FlagSet) {
	fs.StringP("format", "g", "", "go template format string for output")
}

func addFlagForce(fs *pflag.FlagSet) {
	fs.BoolP("force", "f", false, "force")
}

func addFlagAll(fs *pflag.FlagSet) {
	fs.BoolP("all", "a", false, "all")
}
