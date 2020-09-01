package cmds

import (
	"io"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/internal/cmdutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil/exitcodes"
	"github.com/transmutate-io/atomicswap/internal/tplutil"
	"github.com/transmutate-io/atomicswap/stages"
	"github.com/transmutate-io/atomicswap/trade"
	"github.com/transmutate-io/cryptocore/types"
	"gopkg.in/yaml.v2"
)

var (
	TradeCmd = &cobra.Command{
		Use:     "trade",
		Short:   "trade commands",
		Aliases: []string{"t"},
	}
	newTradeCmd = &cobra.Command{
		Use:     "new <name> <own_amount> <own_crypto> <trader_amount> <trader_crypto> <duration>",
		Short:   "create a new trade",
		Aliases: []string{"n"},
		Args:    cobra.ExactArgs(6),
		Run:     cmdNewTrade,
	}
	listTradesCmd = &cobra.Command{
		Use:     "list",
		Short:   "list trades to output",
		Aliases: []string{"l", "ls"},
		Args:    cobra.NoArgs,
		Run:     cmdListTrades,
	}
	renameTradeCmd = &cobra.Command{
		Use:     "rename <old_name> <new_name>",
		Short:   "rename trade",
		Aliases: []string{"ren", "r"},
		Args:    cobra.ExactArgs(2),
		Run:     cmdRenameTrade,
	}
	deleteTradeCmd = &cobra.Command{
		Use:     "delete <name>",
		Short:   "delete a trade",
		Aliases: []string{"d", "del", "rm"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdDeleteTrade,
	}
	exportTradesCmd = &cobra.Command{
		Use:     "export [name1] [name2] [...]",
		Short:   "export trades to output",
		Aliases: []string{"exp", "e"},
		Run:     cmdExportTrades,
	}
	importTradesCmd = &cobra.Command{
		Use:     "import",
		Short:   "import trades from input",
		Aliases: []string{"imp", "i"},
		Args:    cobra.NoArgs,
		Run:     cmdImportTrades,
	}
)

func init() {
	flagutil.AddFlags(flagutil.FlagFuncMap{
		listTradesCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddVerbose,
			flagutil.AddFormat,
			flagutil.AddOutput,
		},
		exportTradesCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddAll,
			flagutil.AddOutput,
		},
		importTradesCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddInput,
		},
	})
	cmdutil.AddCommands(TradeCmd, []*cobra.Command{
		newTradeCmd,
		listTradesCmd,
		renameTradeCmd,
		deleteTradeCmd,
		exportTradesCmd,
		importTradesCmd,
	})
}

func newTrade(ownAmount types.Amount, ownCrypto *cryptos.Crypto, traderAmount types.Amount, traderCrypto *cryptos.Crypto, dur time.Duration) (trade.Trade, error) {
	tr, err := trade.NewOnChainBuy(
		ownAmount, ownCrypto,
		traderAmount, traderCrypto,
		dur,
	)
	if err != nil {
		return nil, err
	}
	th := trade.NewHandler(trade.DefaultStageHandlers)
	th.InstallStageHandler(stages.SendProposal, trade.InterruptHandler)
	for _, i := range th.Unhandled(tr.Stager().Stages()...) {
		th.InstallStageHandler(i, trade.NoOpHandler)
	}
	if err = th.HandleTrade(tr); err != nil {
		return nil, err
	}
	return tr, nil
}

func cmdNewTrade(cmd *cobra.Command, args []string) {
	tr, err := newTrade(
		types.Amount(args[1]), mustParseCrypto(args[2]),
		types.Amount(args[3]), mustParseCrypto(args[4]),
		mustParseDuration(args[5]),
	)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
	mustSaveTrade(cmd, args[0], tr)
}

func listTrades(td string, out io.Writer, tpl *template.Template) error {
	return eachTrade(td, func(name string, tr trade.Trade) error {
		return tpl.Execute(out, newTradeInfo(name, tr))
	})
}

func cmdListTrades(cmd *cobra.Command, args []string) {
	tpl := tplutil.MustOpenTemplate(cmd.Flags(), tradeListTemplates, nil)
	out, closeOut := flagutil.MustOpenOutput(cmd.Flags())
	defer closeOut()
	if err := listTrades(tradesDir(cmd), out, tpl); err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
}

func cmdDeleteTrade(cmd *cobra.Command, args []string) {
	err := os.Remove(tradePath(cmd, args[0]))
	if err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
}

type tradeSelectFunc = func(name string, tr trade.Trade) bool

func exportTrades(td string, tradeSelect tradeSelectFunc) (map[string]trade.Trade, error) {
	trades := make(map[string]trade.Trade, 16)
	err := eachTrade(td, func(name string, tr trade.Trade) error {
		if tradeSelect(name, tr) {
			trades[name] = tr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return trades, nil
}

func cmdExportTrades(cmd *cobra.Command, args []string) {
	var ts tradeSelectFunc
	if flagutil.MustAll(cmd.Flags()) {
		ts = func(name string, tr trade.Trade) bool { return true }
	} else {
		names := make(map[string]struct{}, len(args))
		for _, i := range args {
			names[i] = struct{}{}
		}
		ts = func(name string, tr trade.Trade) bool {
			_, ok := names[name]
			return ok
		}
	}
	trades, err := exportTrades(tradesDir(cmd), ts)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
	if len(trades) == 0 {
		cmdutil.ErrorExit(exitcodes.ExecutionError, "no trades selected")
	}
	out, closeOut := flagutil.MustOpenOutput(cmd.Flags())
	defer closeOut()
	if err = yaml.NewEncoder(out).Encode(trades); err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
}

func cmdImportTrades(cmd *cobra.Command, args []string) {
	in, closeIn := flagutil.MustOpenInput(cmd.Flags())
	defer closeIn()
	trades := make(map[string]*trade.OnChainTrade, 16)
	if err := yaml.NewDecoder(in).Decode(trades); err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
	for n, tr := range trades {
		mustSaveTrade(cmd, filepath.FromSlash(n), tr)
	}
}

func renameFile(oldPath string, newPath string) error {
	d, _ := filepath.Split(newPath)
	if err := os.MkdirAll(d, 0755); err != nil {
		return err
	}
	return os.Rename(oldPath, newPath)
}

func cmdRenameTrade(cmd *cobra.Command, args []string) {
	if err := renameFile(tradePath(cmd, args[0]), tradePath(cmd, args[1])); err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
}
