package cmds

import (
	"text/template"
	"time"

	"github.com/spf13/cobra"
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

func init() {
	addFlags(flagMap{
		listLockSetsCmd.Flags(): []flagFunc{
			addFlagVerbose,
			addFlagFormat,
			addFlagOutput,
		},
		showLockSetInfoCmd.Flags(): []flagFunc{
			addFlagVerbose,
			addFlagFormat,
			addFlagOutput,
			addFlagCryptoChain,
			addFlagInput,
		},
		acceptLockSetCmd.Flags(): []flagFunc{
			addFlagInput,
		},
		exportLockSetCmd.Flags(): []flagFunc{
			addFlagOutput,
		},
	})
	addCommands(LockSetCmd, []*cobra.Command{
		listLockSetsCmd,
		exportLockSetCmd,
		acceptLockSetCmd,
		showLockSetInfoCmd,
	})
}

func cmdListLockSets(cmd *cobra.Command, args []string) {
	tpl := mustOutputTemplate(cmd.Flags(), tradeListTemplates, nil)
	out, closeOut := mustOpenOutput(cmd.Flags())
	defer closeOut()
	err := eachLockSet(tradesDir(cmd), func(name string, tr trade.Trade) error {
		return tpl.Execute(out, newTradeInfo(name, tr))
	})
	if err != nil {
		errorExit(ecCantListLockSets, err)
	}
}

func cmdExportLockSet(cmd *cobra.Command, args []string) {
	tr := mustOpenTrade(cmd, args[0])
	str, err := tr.Seller()
	if err != nil {
		errorExit(ecCantExportLockSet, err)
	}
	ls := str.Locks()
	out, outClose := mustOpenOutput(cmd.Flags())
	defer outClose()
	if err := yaml.NewEncoder(out).Encode(ls); err != nil {
		errorExit(ecCantExportLockSet, err)
	}
}

func cmdAcceptLockSet(cmd *cobra.Command, args []string) {
	tr := mustOpenTrade(cmd, args[0])
	th := trade.NewHandler(trade.DefaultStageHandlers)
	th.InstallStageHandlers(trade.StageHandlerMap{
		stages.ReceiveProposalResponse: func(t trade.Trade) error {
			in, inClose := mustOpenInput(cmd.Flags())
			defer inClose()
			btr, err := tr.Buyer()
			if err != nil {
				return err
			}
			return btr.SetLocks(openLockSet(in, tr.OwnInfo().Crypto, tr.TraderInfo().Crypto))
		},
		stages.LockFunds: trade.InterruptHandler,
	})
	for _, i := range th.Unhandled(tr.Stager().Stages()...) {
		th.InstallStageHandler(i, trade.NoOpHandler)
	}
	if err := th.HandleTrade(tr); err != nil && err != trade.ErrInterruptTrade {
		errorExit(ecCantAcceptLockSet, err)
	}
	mustSaveTrade(cmd, args[0], tr)
}

func cmdShowLockSetInfo(cmd *cobra.Command, args []string) {
	tr := mustOpenTrade(cmd, args[0])
	if _, err := tr.Buyer(); err != nil {
		errorExit(ecCantOpenTrade, err)
	}
	in, inClose := mustOpenInput(cmd.Flags())
	defer inClose()
	ls := openLockSet(in, tr.OwnInfo().Crypto, tr.TraderInfo().Crypto)
	out, outClose := mustOpenOutput(cmd.Flags())
	defer outClose()
	tpl := mustOutputTemplate(cmd.Flags(), lockSetInfoTemplates, template.FuncMap{"now": time.Now})
	err := tpl.Execute(out, newLockSetInfo(
		tr,
		mustNewLockInfo(cmd, ls.Buyer, tr.OwnInfo().Crypto),
		mustNewLockInfo(cmd, ls.Seller, tr.TraderInfo().Crypto),
	))
	if err != nil {
		errorExit(ecBadTemplate, err)
	}
}
