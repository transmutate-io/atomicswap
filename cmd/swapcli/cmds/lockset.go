package cmds

import (
	"io"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/internal/cmdutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil/exitcodes"
	"github.com/transmutate-io/atomicswap/internal/tplutil"
	"github.com/transmutate-io/atomicswap/stages"
	"github.com/transmutate-io/atomicswap/trade"
	"gopkg.in/yaml.v2"
)

var (
	LockSetCmd = &cobra.Command{
		Use:     "lockset <command>",
		Short:   "lockset commands",
		Aliases: []string{"lock", "l"},
	}
	listLockSetsCmd = &cobra.Command{
		Use:     "list",
		Short:   "list trade names with exportable locksets to output",
		Aliases: []string{"ls", "l"},
		Args:    cobra.NoArgs,
		Run:     cmdListLockSets,
	}
	exportLockSetCmd = &cobra.Command{
		Use:     "export <trade_name>",
		Short:   "export a lockset to output",
		Aliases: []string{"exp", "e"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdExportLockSet,
	}
	acceptLockSetCmd = &cobra.Command{
		Use:     "accept <trade_name>",
		Short:   "accept a lockset for a trade from input",
		Aliases: []string{"a"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdAcceptLockSet,
	}
	showLockSetInfoCmd = &cobra.Command{
		Use:     "info <trade_name>",
		Short:   "show a lockset info against a trade to output",
		Aliases: []string{"i"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdShowLockSetInfo,
	}
)

var _network = flagutil.NetworkFlag("mainnet")

func init() {
	network := &_network
	flagutil.AddFlags(flagutil.FlagFuncMap{
		listLockSetsCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddVerbose,
			flagutil.AddFormat,
			flagutil.AddOutput,
		},
		showLockSetInfoCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddVerbose,
			flagutil.AddFormat,
			flagutil.AddOutput,
			network.AddFlag,
			flagutil.AddInput,
		},
		acceptLockSetCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddInput,
		},
		exportLockSetCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddOutput,
		},
	})
	cmdutil.AddCommands(LockSetCmd, []*cobra.Command{
		listLockSetsCmd,
		exportLockSetCmd,
		acceptLockSetCmd,
		showLockSetInfoCmd,
	})
}

func listLockSets(td string, out io.Writer, tpl *template.Template) error {
	return eachLockSet(td, func(name string, tr trade.Trade) error {
		return tpl.Execute(out, newTradeInfo(name, tr))
	})
}

func cmdListLockSets(cmd *cobra.Command, args []string) {
	tpl := tplutil.MustOpenTemplate(cmd.Flags(), tradeListTemplates, nil)
	out, closeOut := flagutil.MustOpenOutput(cmd.Flags())
	defer closeOut()
	if err := listLockSets(tradesDir(cmd), out, tpl); err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
}

func exportLockSet(tp string, out io.Writer) error {
	tr, err := openTradeFile(tp)
	if err != nil {
		return err
	}
	str, err := tr.Seller()
	if err != nil {
		return err
	}
	ls := str.Locks()
	return yaml.NewEncoder(out).Encode(ls)
}

func cmdExportLockSet(cmd *cobra.Command, args []string) {
	out, outClose := flagutil.MustOpenOutput(cmd.Flags())
	defer outClose()
	if err := exportLockSet(tradePath(cmd, args[0]), out); err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
}

func acceptLockSet(tr trade.Trade, lsIn io.Reader) error {
	th := trade.NewHandler(trade.DefaultStageHandlers)
	th.InstallStageHandlers(trade.StageHandlerMap{
		stages.ReceiveProposalResponse: func(t trade.Trade) error {
			btr, err := tr.Buyer()
			if err != nil {
				return err
			}
			return btr.SetLocks(openLockSet(lsIn, tr.OwnInfo().Crypto, tr.TraderInfo().Crypto))
		},
		stages.LockFunds: trade.InterruptHandler,
	})
	for _, i := range th.Unhandled(tr.Stager().Stages()...) {
		th.InstallStageHandler(i, trade.NoOpHandler)
	}
	return th.HandleTrade(tr)
}

func cmdAcceptLockSet(cmd *cobra.Command, args []string) {
	tr := mustOpenTrade(cmd, args[0])
	in, inClose := flagutil.MustOpenInput(cmd.Flags())
	defer inClose()
	if err := acceptLockSet(tr, in); err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
	mustSaveTrade(cmd, args[0], tr)
}

func showLockSetInfo(tp string, lsIn io.Reader, out io.Writer, tpl *template.Template) error {
	tr, err := openTradeFile(tp)
	if err != nil {
		return err
	}
	if _, err := tr.Buyer(); err != nil {
		return err
	}
	ls := openLockSet(lsIn, tr.OwnInfo().Crypto, tr.TraderInfo().Crypto)
	ownLockInfo, err := newLockInfo(ls.Buyer, tr.OwnInfo().Crypto)
	if err != nil {
		return err
	}
	traderLockInfo, err := newLockInfo(ls.Seller, tr.TraderInfo().Crypto)
	if err != nil {
		return err
	}
	return tpl.Execute(out, newLockSetInfo(tr, ownLockInfo, traderLockInfo))
}

func newLockSetTemplate() *template.Template {
	return template.New("main").Funcs(template.FuncMap{"now": time.Now})
}

func cmdShowLockSetInfo(cmd *cobra.Command, args []string) {
	in, inClose := flagutil.MustOpenInput(cmd.Flags())
	defer inClose()
	out, outClose := flagutil.MustOpenOutput(cmd.Flags())
	defer outClose()
	tpl := tplutil.MustOpenTemplate(cmd.Flags(), lockSetInfoTemplates, template.FuncMap{"now": time.Now})
	if err := showLockSetInfo(tradePath(cmd, args[0]), in, out, tpl); err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
}
