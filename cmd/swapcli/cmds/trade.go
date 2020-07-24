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
		Args:    cobra.NoArgs,
		Run:     cmdExportTrades,
	}
	importTradesCmd = &cobra.Command{
		Use:     "import",
		Short:   "import trades from input",
		Aliases: []string{"imp", "i"},
		Args:    cobra.NoArgs,
		Run:     cmdImportTrades,
	}
	renameTradeCmd = &cobra.Command{
		Use:     "rename <old_name> <new_name>",
		Short:   "rename trade",
		Aliases: []string{"ren", "r"},
		Args:    cobra.ExactArgs(2),
		Run:     cmdRenameTrade,
	}
)

func init() {
	fs := listTradesCmd.Flags()
	addFlagVerbose(fs)
	addFlagFormat(fs)
	addFlagAll(exportTradesCmd.Flags())
	for _, i := range []*cobra.Command{
		newTradeCmd,
		listTradesCmd,
		deleteTradeCmd,
		exportTradesCmd,
		importTradesCmd,
	} {
		TradeCmd.AddCommand(i)
	}
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
	h := trade.NewHandler(trade.DefaultStageHandlers)
	h.InstallStageHandler(stages.SendProposal, func(tr trade.Trade) error {
		return trade.ErrInterruptTrade
	})
	for _, i := range h.Unhandled(tr.Stager().Stages()...) {
		h.InstallStageHandler(i, trade.NoOpHandler)
	}
	if err = h.HandleTrade(tr); err != nil && err != trade.ErrInterruptTrade {
		errorExit(ecCantCreateTrade, err)
	}
	if err = saveTrade(cmd, args[0], tr); err != nil {
		errorExit(ecCantCreateTrade, err)
	}
}

func cmdListTrades(cmd *cobra.Command, args []string) {
	tpl := outputTemplate(cmd, tradeListTemplates, nil)
	out, closeOut := openOutput(cmd)
	defer closeOut()
	err := eachTrade(tradesDir(dataDir(cmd)), func(name string, tr trade.Trade) error {
		return tpl.Execute(out, &tradeInfo{Name: name, Trade: tr})
	})
	if err != nil {
		errorExit(ecCantListTrades, err)
	}
}

type tradeInfo struct {
	Name  string
	Trade trade.Trade
}

var tradeListTemplates = []string{
	"{{ .Name }}\n",
	"{{ .Name }} - {{ .Trade.OwnInfo.Amount }} {{ .Trade.OwnInfo.Crypto.Short }} for {{ .Trade.TraderInfo.Amount }} {{ .Trade.TraderInfo.Crypto.Short }}\n",
	"{{ .Name }} - {{ .Trade.OwnInfo.Amount }} {{ .Trade.OwnInfo.Crypto.Short }} (locked for {{ .Trade.Duration.String }}) for {{ .Trade.TraderInfo.Amount }} {{ .Trade.TraderInfo.Crypto.Short }}\n",
	"{{ .Name }} - {{ .Trade.OwnInfo.Amount }} {{ .Trade.OwnInfo.Crypto.Short }} (locked for {{ .Trade.Duration.String }}) for {{ .Trade.TraderInfo.Amount }} {{ .Trade.TraderInfo.Crypto.Short }} - {{ .Trade.Stager.Stage }}\n",
}

func cmdDeleteTrade(cmd *cobra.Command, args []string) {
	err := os.Remove(filepath.Join(tradesDir(dataDir(cmd)), cmd.Flags().Arg(0)))
	if err != nil {
		errorExit(ecCantDeleteTrade, err)
	}
}

func cmdExportTrades(cmd *cobra.Command, args []string) {
	exportAll := flagBool(cmd.Flags(), "all")
	names := make(map[string]struct{}, len(args))
	for _, i := range args {
		names[i] = struct{}{}
	}
	trades := make(map[string]trade.Trade, 16)
	err := eachTrade(tradesDir(dataDir(cmd)), func(name string, tr trade.Trade) error {
		if _, exp := names[name]; exp || exportAll {
			delete(names, name)
			trades[name] = tr
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
	out, closeOut := openOutput(cmd)
	defer closeOut()
	if err = yaml.NewEncoder(out).Encode(trades); err != nil {
		errorExit(ecCantExportTrades, err)
	}
}

func cmdImportTrades(cmd *cobra.Command, args []string) {
	in, closeIn := openInput(cmd)
	defer closeIn()
	trades := make(map[string]*trade.OnChainTrade, 16)
	if err := yaml.NewDecoder(in).Decode(trades); err != nil {
		errorExit(ecCantImportTrades, err)
	}
	for n, tr := range trades {
		if err := saveTrade(cmd, n, tr); err != nil {
			errorExit(ecCantImportTrades, err)
		}
	}
}

func cmdRenameTrade(cmd *cobra.Command, args []string) {
	err := os.Rename(
		filepath.Join(tradesDir(dataDir(cmd)), args[0]),
		filepath.Join(tradesDir(dataDir(cmd)), args[1]),
	)
	if err != nil {
		errorExit(ecCantRenameTrade, err)
	}
}
