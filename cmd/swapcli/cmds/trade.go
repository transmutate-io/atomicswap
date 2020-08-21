package cmds

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
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
	addFlags(flagMap{
		listTradesCmd.Flags(): []flagFunc{
			addFlagVerbose,
			addFlagFormat,
			addFlagOutput,
		},
		exportTradesCmd.Flags(): []flagFunc{
			addFlagAll,
		},
		importTradesCmd.Flags(): []flagFunc{
			addFlagInput,
		},
		exportTradesCmd.Flags(): []flagFunc{
			addFlagOutput,
		},
	})
	addCommands(TradeCmd, []*cobra.Command{
		newTradeCmd,
		listTradesCmd,
		renameTradeCmd,
		deleteTradeCmd,
		exportTradesCmd,
		importTradesCmd,
	})
}

func cmdNewTrade(cmd *cobra.Command, args []string) {
	tr, err := trade.NewOnChainBuy(
		types.Amount(args[1]), parseCrypto(args[2]),
		types.Amount(args[3]), parseCrypto(args[4]),
		parseDuration(args[5]),
	)
	if err != nil {
		errorExit(ecCantCreateTrade, err)
	}
	th := trade.NewHandler(trade.DefaultStageHandlers)
	th.InstallStageHandler(stages.SendProposal, trade.InterruptHandler)
	for _, i := range th.Unhandled(tr.Stager().Stages()...) {
		th.InstallStageHandler(i, trade.NoOpHandler)
	}
	if err = th.HandleTrade(tr); err != nil && err != trade.ErrInterruptTrade {
		errorExit(ecCantCreateTrade, err)
	}
	saveTrade(cmd, args[0], tr)
}

func cmdListTrades(cmd *cobra.Command, args []string) {
	tpl := mustOutputTemplate(cmd.Flags(), tradeListTemplates, nil)
	out, closeOut := mustOpenOutput(cmd.Flags())
	defer closeOut()
	err := eachTrade(tradesDir(cmd), func(name string, tr trade.Trade) error {
		return tpl.Execute(out, newTradeInfo(name, tr))
	})
	if err != nil {
		errorExit(ecCantListTrades, err)
	}
}

func cmdDeleteTrade(cmd *cobra.Command, args []string) {
	err := os.Remove(tradePath(cmd, args[0]))
	if err != nil {
		errorExit(ecCantDeleteTrade, err)
	}
}

func cmdExportTrades(cmd *cobra.Command, args []string) {
	names := make(map[string]struct{}, len(args))
	for _, i := range args {
		names[i] = struct{}{}
	}
	trades := make(map[string]trade.Trade, 16)
	err := eachTrade(tradesDir(cmd), func(name string, tr trade.Trade) error {
		if _, exp := names[name]; exp || mustFlagAll(cmd.Flags()) {
			delete(names, name)
			trades[filepath.ToSlash(name)] = tr
		}
		return nil
	})
	if err != nil {
		errorExit(ecCantExportTrades, err)
	}
	n := make([]string, 0, len(names))
	for i := range names {
		n = append(n, i)
	}
	if len(names) > 0 {
		errorExit(ecCantExportTrades, strings.Join(n, ", "))
	}
	out, closeOut := mustOpenOutput(cmd.Flags())
	defer closeOut()
	if err = yaml.NewEncoder(out).Encode(trades); err != nil {
		errorExit(ecCantExportTrades, err)
	}
}

func cmdImportTrades(cmd *cobra.Command, args []string) {
	in, closeIn := mustOpenInput(cmd.Flags())
	defer closeIn()
	trades := make(map[string]*trade.OnChainTrade, 16)
	if err := yaml.NewDecoder(in).Decode(trades); err != nil {
		errorExit(ecCantImportTrades, err)
	}
	for n, tr := range trades {
		n = filepath.Join(strings.Split(n, "/")...)
		saveTrade(cmd, n, tr)
	}
}

func cmdRenameTrade(cmd *cobra.Command, args []string) {
	err := os.Rename(tradePath(cmd, args[0]), tradePath(cmd, args[1]))
	if err != nil {
		errorExit(ecCantRenameTrade, err)
	}
}
