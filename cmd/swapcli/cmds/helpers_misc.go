package cmds

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/internal/cmdutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil/exitcodes"
	"github.com/transmutate-io/atomicswap/trade"
	"gopkg.in/yaml.v2"
)

func dataDir(cmd *cobra.Command) string {
	return filepath.Clean(flagutil.MustString(cmd.Root().PersistentFlags(), "data"))
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

func mustCreateFile(p string) *os.File {
	r, err := createFile(p)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.CantCreateFile, err)
	}
	return r
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
		cmdutil.ErrorExit(exitcodes.UnknownCrypto, c)
	}
	return r
}

func mustParseDuration(d string) time.Duration {
	r, err := time.ParseDuration(d)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.InvalidDuration, d)
	}
	return r
}

func eachTrade(td string, f func(string, trade.Trade) error) error {
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
		return f(filepath.ToSlash(cmdutil.TrimPath(path, td)), tr)
	})
}

func eachProposal(td string, f func(string, trade.Trade) error) error {
	return eachTrade(td, func(name string, tr trade.Trade) error {
		return f(name, tr)
	})
}

func openTradeFile(tp string) (trade.Trade, error) {
	f, err := os.Open(tp)
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

func mustOpenTrade(cmd *cobra.Command, name string) trade.Trade {
	r, err := openTradeFile(tradePath(cmd, name))
	if err != nil {
		cmdutil.ErrorExit(exitcodes.CantOpenTrade, name, err)
	}
	return r
}

func saveTrade(tp string, tr trade.Trade) error {
	f, err := createFile(tp)
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
	if err := saveTrade(tradePath(cmd, name), tr); err != nil {
		cmdutil.ErrorExit(exitcodes.CantSaveTrade, err)
	}
}

func openLockSet(r io.Reader, ownCrypto, traderCrypto *cryptos.Crypto) *trade.Locks {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.CantOpenLockSet, err)
	}
	ls, err := trade.UnamrshalLocks(ownCrypto, traderCrypto, b)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.CantOpenLockSet, err)
	}
	return ls
}

func watchDataDir(cmd *cobra.Command) string {
	return filepath.Join(dataDir(cmd), "watch_data")
}

func watchDataPath(cmd *cobra.Command, name string) string {
	return filepath.Join(watchDataDir(cmd), name)
}

func openWatchData(wdPath string) (*watchData, error) {
	r := &watchData{
		Own:    &blockWatchData{Top: 0, Bottom: 0},
		Trader: &blockWatchData{Top: 0, Bottom: 0},
	}
	f, err := os.Open(wdPath)
	if err != nil {
		if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOENT {
			return r, nil
		}
		return nil, err
	}
	defer f.Close()
	if err = yaml.NewDecoder(f).Decode(r); err != nil {
		return nil, err
	}
	return r, nil
}

func mustOpenWatchData(cmd *cobra.Command, name string) *watchData {
	r, err := openWatchData(watchDataPath(cmd, name))
	if err != nil {
		cmdutil.ErrorExit(exitcodes.CantOpenWatchData, err)
	}
	return r
}

func saveWatchData(wdPath string, wd *watchData) error {
	f, err := createFile(wdPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewEncoder(f).Encode(wd)
}

func mustSaveWatchData(cmd *cobra.Command, name string, wd *watchData) {
	if err := saveWatchData(watchDataPath(cmd, name), wd); err != nil {
		cmdutil.ErrorExit(exitcodes.CantSaveWatchData, err)
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
