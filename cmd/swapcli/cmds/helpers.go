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
	"github.com/transmutate-io/atomicswap/trade"
	"gopkg.in/yaml.v2"
)

const (
	noOutput         = -1
	cantOpenOutput   = -2
	cantGetFlag      = -3
	unknownShell     = -4
	unknownCrypto    = -5
	invalidDuration  = -6
	cantCreateTrade  = -7
	cantFindTrade    = -8
	badTemplate      = -9
	cantListTrades   = -10
	cantShowTrades   = -11
	cantShowProposal = -12
	cantListProposal = -13
	cantOpenTrade    = -14
)

func openOutput(cmd *cobra.Command) (io.Writer, func() error) {
	outfn, err := cmd.Root().PersistentFlags().GetString("output")
	if err != nil {
		errorExit(noOutput, "can't get output: %#v\n", err)
	}
	if outfn == "-" {
		return os.Stdout, func() error { return nil }
	}
	f, err := os.Create(outfn)
	if err != nil {
		errorExit(cantOpenOutput, "can't create output file: %#v\n", err)
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
		errorExit(cantGetFlag, "can't get flag: %#v\n", err)
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
		if info.IsDir() {
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

func flagString(fs *pflag.FlagSet, name string) string {
	r, err := fs.GetString(name)
	if err != nil {
		errorExit(cantGetFlag, "can't get flag: %#v\n", err)
	}
	return r
}

func flagDuration(fs *pflag.FlagSet, name string) time.Duration {
	r, err := fs.GetDuration(name)
	if err != nil {
		errorExit(cantGetFlag, "can't get flag: %#v\n", err)
	}
	return r
}

func flagBool(fs *pflag.FlagSet, name string) bool {
	r, err := fs.GetBool(name)
	if err != nil {
		errorExit(cantGetFlag, "can't get flag: %#v\n", err)
	}
	return r
}

func addVerboseFlag(fs *pflag.FlagSet) {
	fs.CountP("verbose", "v", "increse verbose level")
}

func addFormatFlag(fs *pflag.FlagSet) {
	fs.StringP("format", "g", "", "go template format string for output")
}

func addForceFlag(fs *pflag.FlagSet) {
	fs.BoolP("force", "f", false, "force")
}

func addAllFlag(fs *pflag.FlagSet) {
	fs.BoolP("all", "a", false, "all")
}
