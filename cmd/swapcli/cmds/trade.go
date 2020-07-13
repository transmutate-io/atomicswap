package cmds

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

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
		Run:     cmdNewTrade,
		Args:    cobra.ExactArgs(6),
	}
	listTradesCmd = &cobra.Command{
		Use:     "list",
		Short:   "list trades",
		Aliases: []string{"l", "ls"},
		Run:     cmdListTrades,
	}
	deleteTradeCmd = &cobra.Command{
		Use:     "delete <name>",
		Short:   "delete trade",
		Aliases: []string{"d", "del", "rm"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdDeleteTrade,
	}
	exportTradesCmd = &cobra.Command{
		Use:     "export [name1] [name2] [...]",
		Short:   "export trade(s)",
		Aliases: []string{"exp", "e"},
		Run:     cmdExportTrades,
	}
	importTradesCmd = &cobra.Command{
		Use:     "import",
		Short:   "import trade(s)",
		Aliases: []string{"imp", "i"},
		Run:     cmdImportTrades,
	}
)

func init() {
	addFlagVerbose(listTradesCmd.Flags())
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
	fs := cmd.Flags()
	ownCrypto, err := parseCrypto(fs.Arg(2))
	if err != nil {
		errorExit(ECUnknownCrypto, "unknown crypto: %s\n", fs.Arg(2))
	}
	traderCrypto, err := parseCrypto(fs.Arg(4))
	if err != nil {
		errorExit(ECUnknownCrypto, "unknown crypto: %s\n", fs.Arg(4))
	}
	dur, err := time.ParseDuration(fs.Arg(5))
	if err != nil {
		errorExit(ECInvalidDuration, "invalid duration: %s\n", fs.Arg(5))
	}
	tr, err := trade.NewOnChainBuy(
		types.Amount(fs.Arg(1)), ownCrypto,
		types.Amount(fs.Arg(3)), traderCrypto,
		dur,
	)
	if err != nil {
		errorExit(ECCantCreateTrade, "can't create trade: %#v\n", err)
	}
	h := trade.NewHandler(trade.DefaultStageHandlers)
	out, closeOut := openOutput(cmd)
	defer closeOut()
	h.InstallStageHandler(stages.SendProposal, func(tr trade.Trade) error {
		prop, err := tr.GenerateBuyProposal()
		if err != nil {
			return err
		}
		if err = yaml.NewEncoder(out).Encode(prop); err != nil {
			return err
		}
		return trade.ErrInterruptTrade
	})
	for _, i := range h.Unhandled(tr.Stager().Stages()...) {
		h.InstallStageHandler(i, trade.NoOpHandler)
	}
	if err = h.HandleTrade(tr); err != nil && err != trade.ErrInterruptTrade {
		errorExit(ECCantCreateTrade, "can't create trade: %#v\n", err.Error())
	}
	if err = saveTrade(cmd, fs.Arg(0), tr); err != nil {
		errorExit(ECCantCreateTrade, "can't create trade: %#v\n", err)
	}
}

func cmdListTrades(cmd *cobra.Command, args []string) {
	vl := verboseLevel(cmd.Flags(), len(tradeListTemplates)-1)
	tpl, err := template.New("main").Parse(tradeListTemplates[vl])
	if err != nil {
		errorExit(ECBadTemplate, "bad template: %#v\n", err)
	}
	out, closeOut := openOutput(cmd)
	defer closeOut()
	err = eachTrade(tradesDir(dataDir(cmd)), func(name string, tr trade.Trade) error {
		return tpl.Execute(out, &listEntry{
			Name:  name,
			Trade: tr,
		})
	})
	if err != nil {
		errorExit(ECCantListTrades, "can't list trades: %#v\n", err)
	}
}

type listEntry struct {
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
		errorExit(ECCantFindTrade, `can't delete trade: %#v\n`, err)
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
		errorExit(ECCantExportTrades, "can't export trades: %#v\n", err)
	}
	n := make([]string, 0, len(names))
	for i := range names {
		n = append(n, i)
	}
	if len(names) > 0 {
		errorExit(ECCantExportTrades, "missing trades: %s\n", strings.Join(n, ", "))
	}
	out, closeOut := openOutput(cmd)
	defer closeOut()
	if err = yaml.NewEncoder(out).Encode(trades); err != nil {
		errorExit(ECCantExportTrades, "can't marshal trades: %#v\n", err)
	}
}

func cmdImportTrades(cmd *cobra.Command, args []string) {
	in, closeIn := openInput(cmd)
	defer closeIn()
	trades := make(map[string]*trade.OnChainTrade, 16)
	if err := yaml.NewDecoder(in).Decode(trades); err != nil {
		errorExit(ECCantImportTrades, "can't import trades: %#v\n", err)
	}
	for n, tr := range trades {
		if err := saveTrade(cmd, n, tr); err != nil {
			errorExit(ECCantImportTrades, "can't import trade \"%s\": %#v\n", n, err)
		}
	}
}
