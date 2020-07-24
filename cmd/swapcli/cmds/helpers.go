package cmds

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/internal/testutil"
	"github.com/transmutate-io/atomicswap/params"
	"github.com/transmutate-io/atomicswap/stages"
	"github.com/transmutate-io/atomicswap/trade"
	"gopkg.in/yaml.v2"
)

const (
	ecOK = (iota + 1) * -1
	ecBadTemplate
	ecCantCalculateAddress
	ecCantCreateTrade
	ecCantDeleteTrade
	ecCantExportLockSet
	ecCantAcceptLockSet
	ecCantExportProposal
	ecCantExportTrades
	ecCantFindNetwork
	ecCantGetFlag
	ecCantImportTrades
	ecCantRenameTrade
	ecCantListLockSets
	ecCantListProposals
	ecCantListTrades
	ecCantOpenLockSet
	ecCantOpenOutput
	ecCantOpenTrade
	ecInvalidLockData
	ecInvalidDuration
	ecNoInput
	ecNoOutput
	ecOnlyOneNetwork
	ecUnknownCrypto
	ecUnknownShell
	ecNotABuyer
	ecNotASeller
)

var ecMessages = map[int]string{
	ecBadTemplate:          "bad template: %s\n",
	ecCantCalculateAddress: "can't calculate address: %s\n",
	ecCantCreateTrade:      "can't create trade: %s\n",
	ecCantDeleteTrade:      "can't delete trade: %s\n",
	ecCantExportLockSet:    "can't export lock set: %s\n",
	ecCantExportProposal:   "can't export proposal: %s\n",
	ecCantExportTrades:     "can't export trades: %s\n",
	ecCantFindNetwork:      "can't find network for %s\n",
	ecCantGetFlag:          "can't get flag: %s\n",
	ecCantImportTrades:     "can't import trades: %s\n",
	ecCantRenameTrade:      "can't rename trade: %s\n",
	ecCantListLockSets:     "can't list locksets: %s\n",
	ecCantListProposals:    "can't list proposals: %s\n",
	ecCantListTrades:       "can't list trades: %s\n",
	ecCantOpenLockSet:      "can't open lock set: %s\n",
	ecCantOpenOutput:       "can't create output file: %s\n",
	ecCantOpenTrade:        "can't open trade \"%s\": %s\n",
	ecInvalidLockData:      "invalid lock data: %s\n",
	ecInvalidDuration:      "invalid duration: \"%s\"\n",
	ecCantAcceptLockSet:    "can't accept lock set: %s\n",
	ecNoInput:              "can't get input: %s\n",
	ecNoOutput:             "can't get output: %s\n",
	ecOnlyOneNetwork:       "pick only one network\n",
	ecUnknownCrypto:        "unknown crypto: \"%s\"\n",
	ecUnknownShell:         "can't generate completion file: %s\n",
	ecNotABuyer:            "not a buyer\n",
	ecNotASeller:           "not a seller\n",
}

func openOutput(cmd *cobra.Command) (io.Writer, func() error) {
	outfn, err := cmd.Root().PersistentFlags().GetString("output")
	if err != nil {
		errorExit(ecNoOutput, err)
	}
	if outfn == "-" {
		return os.Stdout, func() error { return nil }
	}
	f, err := os.Create(outfn)
	if err != nil {
		errorExit(ecCantOpenOutput, err)
	}
	return f, func() error { return f.Close() }
}

func openInput(cmd *cobra.Command) (io.Reader, func() error) {
	infn, err := cmd.Root().PersistentFlags().GetString("input")
	if err != nil {
		errorExit(ecNoInput, err)
	}
	if infn == "-" {
		return os.Stdin, func() error { return nil }
	}
	f, err := os.Open(infn)
	if err != nil {
		errorExit(ecNoInput, err)
	}
	return f, func() error { return f.Close() }
}

func errorExit(code int, a ...interface{}) {
	f, ok := ecMessages[code]
	if !ok {
		t := make([]string, 0, len(a))
		for i := 0; i < len(a); i++ {
			t = append(t, "%#v")
		}
		f = "args: " + strings.Join(t, " ") + "\n"
	}
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
	r := flagCount(fs, "verbose")
	if r > max {
		return max
	}
	return r
}

func cryptoChain(cmd *cobra.Command, c *cryptos.Crypto) params.Chain {
	fs := cmd.Root().PersistentFlags()
	tn := flagBool(fs, "testnet")
	ln := flagBool(fs, "local")
	if tn && ln {
		errorExit(ecOnlyOneNetwork)
	}
	if tn {
		return params.TestNet
	}
	if ln {
		for _, i := range testutil.Cryptos {
			if i.Name == c.Name {
				return i.Chain
			}
		}
		errorExit(ecCantFindNetwork, c.Name)
	}
	return params.MainNet
}

func parseCrypto(c string) *cryptos.Crypto {
	if r, err := cryptos.ParseShort(c); err == nil {
		return r
	}
	r, err := cryptos.Parse(c)
	if err != nil {
		errorExit(ecUnknownCrypto, c)
	}
	return r
}

func parseDuration(d string) time.Duration {
	r, err := time.ParseDuration(d)
	if err != nil {
		errorExit(ecInvalidDuration, d)
	}
	return r
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
		if tr.Stager().Stage() != stages.SendProposalResponse {
			return nil
		}
		return f(name, tr)
	})
}

func openTrade(cmd *cobra.Command, name string) trade.Trade {
	f, err := os.Open(filepath.Join(tradesDir(dataDir(cmd)), name))
	if err != nil {
		errorExit(ecCantOpenTrade, name, err)
	}
	defer f.Close()
	r := &trade.OnChainTrade{}
	if err = yaml.NewDecoder(f).Decode(r); err != nil {
		errorExit(ecCantOpenTrade, name, err)
	}
	return r
}

func saveTrade(cmd *cobra.Command, name string, tr trade.Trade) error {
	f, err := createFile(filepath.Join(tradesDir(dataDir(cmd)), name))
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewEncoder(f).Encode(tr)
}

func openLockSet(r io.Reader, ownCrypto, traderCrypto *cryptos.Crypto) *trade.BuyProposalResponse {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		errorExit(ecCantOpenLockSet, err)
	}
	ls, err := trade.UnamrshalBuyProposalResponse(ownCrypto, traderCrypto, b)
	if err != nil {
		errorExit(ecCantOpenLockSet, err)
	}
	return ls
}

func outputTemplate(cmd *cobra.Command, tpls []string, funcs template.FuncMap) *template.Template {
	var tplStr string
	if tplStr = flagString(cmd.Flags(), "format"); tplStr == "" {
		tplStr = tpls[verboseLevel(cmd.Flags(), len(tpls)-1)]
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

func addFlagVerbose(fs *pflag.FlagSet) {
	fs.CountP("verbose", "v", "increse verbose level")
}

func addFlagFormat(fs *pflag.FlagSet) {
	fs.StringP("format", "f", "", "go template format string for output")
}

func addFlagForce(fs *pflag.FlagSet) {
	fs.BoolP("force", "f", false, "force")
}

func addFlagAll(fs *pflag.FlagSet) {
	fs.BoolP("all", "a", false, "all")
}
