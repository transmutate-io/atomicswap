package cmds

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/stages"
	"github.com/transmutate-io/atomicswap/trade"
	"gopkg.in/yaml.v2"
)

const (
	ecOK = iota * -1
	ecBadTemplate
	ecCantAcceptLockSet
	ecCantCalculateAddress
	ecCantCreateTrade
	ecCantDeleteTrade
	ecCantExportLockSet
	ecCantExportProposal
	ecCantExportTrades
	ecCantGetFlag
	ecCantImportTrades
	ecCantListLockSets
	ecCantListProposals
	ecCantListTrades
	ecCantLoadConfig
	ecCantOpenLockSet
	ecCantOpenOutput
	ecCantOpenTrade
	ecCantOpenWatchData
	ecCantRecover
	ecCantRedeem
	ecCantRenameTrade
	ecCantSaveTrade
	ecCantSaveWatchData
	ecFailedToWatch
	ecInvalidDuration
	ecInvalidLockData
	ecNoInput
	ecNotABuyer
	ecOnlyOneNetwork
	ecUnknownCrypto
	ecUnknownShell
)

var ecMessages = map[int]string{
	ecBadTemplate:          "bad template: %s\n",
	ecCantAcceptLockSet:    "can't accept lock set: %s\n",
	ecCantCalculateAddress: "can't calculate address: %s\n",
	ecCantCreateTrade:      "can't create trade: %s\n",
	ecCantDeleteTrade:      "can't delete trade: %s\n",
	ecCantExportLockSet:    "can't export lock set: %s\n",
	ecCantExportProposal:   "can't export proposal: %s\n",
	ecCantExportTrades:     "can't export trades: %s\n",
	ecCantGetFlag:          "can't get flag: %s\n",
	ecCantImportTrades:     "can't import trades: %s\n",
	ecCantListLockSets:     "can't list locksets: %s\n",
	ecCantListProposals:    "can't list proposals: %s\n",
	ecCantListTrades:       "can't list trades: %s\n",
	ecCantLoadConfig:       "can't load config: %s\n",
	ecCantOpenLockSet:      "can't open lock set: %s\n",
	ecCantOpenOutput:       "can't create output file: %s\n",
	ecCantOpenTrade:        "can't open trade \"%s\": %s\n",
	ecCantOpenWatchData:    "can't open watch data: %s\n",
	ecCantRecover:          "can't recover funds: %s\n",
	ecCantRedeem:           "can't redeem: %s\n",
	ecCantRenameTrade:      "can't rename trade: %s\n",
	ecCantSaveTrade:        "can't save trade: %s\n",
	ecCantSaveWatchData:    "can't save watch data: %s\n",
	ecFailedToWatch:        "failed to watch blockchain: %s\n",
	ecInvalidDuration:      "invalid duration: \"%s\"\n",
	ecInvalidLockData:      "invalid lock data: %s\n",
	ecNoInput:              "can't get input: %s\n",
	ecNotABuyer:            "not a buyer\n",
	ecOnlyOneNetwork:       "pick only one network\n",
	ecUnknownCrypto:        "unknown crypto: \"%s\"\n",
	ecUnknownShell:         "can't generate completion file: %s\n",
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
	return filepath.Clean(mustFlagString(cmd.Root().PersistentFlags(), "data"))
}

func tradesDir(cmd *cobra.Command) string {
	return filepath.Join(dataDir(cmd), "trades")
}

func tradePath(cmd *cobra.Command, name string) string {
	return filepath.Join(tradesDir(cmd), name)
}

func createFile(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return nil, err
	}
	return os.Create(p)
}

func parseCrypto(c string) (*cryptos.Crypto, error) {
	if r, err := cryptos.ParseShort(c); err == nil {
		return r, nil
	}
	return cryptos.Parse(c)
}

func mustParseCrypto(c string) *cryptos.Crypto {
	r, err := parseCrypto(c)
	if err != nil {
		errorExit(ecUnknownCrypto, c)
	}
	return r
}

func mustParseDuration(d string) time.Duration {
	r, err := time.ParseDuration(d)
	if err != nil {
		errorExit(ecInvalidDuration, d)
	}
	return r
}

func eachTrade(td string, f func(string, trade.Trade) error) error {
	tdPrefix := td + string([]rune{filepath.Separator})
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
	f, err := os.Open(tradePath(cmd, name))
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
	f, err := createFile(tradePath(cmd, name))
	if err != nil {
		return err
	}
	defer f.Close()
	if err = yaml.NewEncoder(f).Encode(tr); err != nil {
		return err
	}
	return nil
}

func mustSaveTrade(cmd *cobra.Command, name string, tr trade.Trade) {
	if err := saveTrade(cmd, name, tr); err != nil {
		errorExit(ecCantSaveTrade, err)
	}
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

func watchDataDir(cmd *cobra.Command) string {
	return filepath.Join(dataDir(cmd), "watch_data")
}

func watchDataPath(cmd *cobra.Command, name string) string {
	return filepath.Join(watchDataDir(cmd), name)
}

func openWatchData(cmd *cobra.Command, name string) *watchData {
	r := &watchData{
		Own:    &blockWatchData{Top: 0, Bottom: 0},
		Trader: &blockWatchData{Top: 0, Bottom: 0},
	}
	f, err := os.Open(watchDataPath(cmd, name))
	if err != nil {
		if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOENT {
			return r
		}
		errorExit(ecCantOpenWatchData, err)
	}
	defer f.Close()
	if err = yaml.NewDecoder(f).Decode(r); err != nil {
		errorExit(ecCantOpenWatchData, err)
	}
	return r
}

func saveWatchData(cmd *cobra.Command, name string, wd *watchData) {
	f, err := createFile(watchDataPath(cmd, name))
	if err != nil {
		errorExit(ecCantSaveWatchData, err)
	}
	defer f.Close()
	if err = yaml.NewEncoder(f).Encode(wd); err != nil {
		errorExit(ecCantSaveWatchData, err)
	}
}

func consoleConfigDir(cmd *cobra.Command) string { return filepath.Join(dataDir(cmd), "config") }

const DEFAULT_CONSOLE_CONFIG_NAME = "console_defaults.yaml"

func consoleConfigPath(cmd *cobra.Command, name string) string {
	if name == "" {
		name = DEFAULT_CONSOLE_CONFIG_NAME
	}
	return filepath.Join(consoleConfigDir(cmd), name)
}
