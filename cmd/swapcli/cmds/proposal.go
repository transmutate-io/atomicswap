package cmds

import (
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/roles"
	"github.com/transmutate-io/atomicswap/stages"
	"github.com/transmutate-io/atomicswap/trade"
	"gopkg.in/yaml.v2"
)

var (
	ProposalCmd = &cobra.Command{
		Use:     "proposal <command>",
		Short:   "proposal commands",
		Aliases: []string{"prop", "p"},
	}
	listProposalsCmd = &cobra.Command{
		Use:     "list",
		Short:   "list trade names with exportable proposals to output",
		Aliases: []string{"ls", "l"},
		Args:    cobra.NoArgs,
		Run:     cmdListProposals,
	}
	exportProposalCmd = &cobra.Command{
		Use:     "export <trade_name>",
		Short:   "export a proposal to output",
		Aliases: []string{"exp", "e"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdExportProposal,
	}
	acceptProposalCmd = &cobra.Command{
		Use:     "accept <trade_name>",
		Short:   "accept a proposal from input",
		Aliases: []string{"a"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdAcceptProposal,
	}
)

func init() {
	addFlags(flagMap{
		listProposalsCmd.Flags(): []flagFunc{
			addFlagFormat,
			addFlagVerbose,
			addFlagOutput,
		},
		acceptProposalCmd.Flags(): []flagFunc{
			addFlagInput,
		},
		exportProposalCmd.Flags(): []flagFunc{
			addFlagOutput,
		},
	})
	addCommands(ProposalCmd, []*cobra.Command{
		listProposalsCmd,
		exportProposalCmd,
		acceptProposalCmd,
	})
}

func cmdListProposals(cmd *cobra.Command, args []string) {
	tpl := outputTemplate(cmd.Flags(), tradeListTemplates, nil)
	out, closeOut := openOutput(cmd.Flags())
	defer closeOut()
	err := eachProposal(tradesDir(cmd), func(name string, tr trade.Trade) error {
		return tpl.Execute(out, newTradeInfo(name, tr))
	})
	if err != nil {
		errorExit(ecCantListProposals, err)
	}
}

func cmdExportProposal(cmd *cobra.Command, args []string) {
	tr := openTrade(cmd, args[0])
	if tr.Role() != roles.Buyer {
		errorExit(ecNotABuyer)
	}
	btr, err := tr.Buyer()
	if err != nil {
		errorExit(ecCantExportProposal, err)
	}
	prop, err := btr.GenerateBuyProposal()
	if err != nil {
		errorExit(ecCantExportProposal, err)
	}
	out, closeOut := openOutput(cmd.Flags())
	defer closeOut()
	if err = yaml.NewEncoder(out).Encode(prop); err != nil {
		errorExit(ecCantExportProposal, err)
	}
}

func cmdAcceptProposal(cmd *cobra.Command, args []string) {
	newTrade := trade.NewOnChainSell()
	th := trade.NewHandler(trade.DefaultStageHandlers)
	th.InstallStageHandlers(trade.StageHandlerMap{
		stages.ReceiveProposal: func(tr trade.Trade) error {
			in, inClose := openInput(cmd.Flags())
			defer inClose()
			b, err := ioutil.ReadAll(in)
			if err != nil {
				return err
			}
			prop, err := trade.UnamrshalBuyProposal(b)
			if err != nil {
				return err
			}
			str, err := tr.Seller()
			if err != nil {
				return err
			}
			if err := str.AcceptBuyProposal(prop); err != nil {
				return err
			}
			return nil
		},
		stages.SendProposalResponse: trade.InterruptHandler,
	})
	for _, i := range th.Unhandled(newTrade.Stager().Stages()...) {
		th.InstallStageHandler(i, trade.NoOpHandler)
	}
	if err := th.HandleTrade(newTrade); err != nil && err != trade.ErrInterruptTrade {
		errorExit(ecCantCreateTrade, err)
	}
	saveTrade(cmd, args[0], newTrade)
}
