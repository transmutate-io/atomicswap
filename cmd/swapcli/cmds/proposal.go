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
	fs := listProposalsCmd.Flags()
	addFlagFormat(fs)
	addFlagVerbose(fs)
	for _, i := range []*cobra.Command{
		listProposalsCmd,
		exportProposalCmd,
		acceptProposalCmd,
	} {
		ProposalCmd.AddCommand(i)
	}
}

func cmdListProposals(cmd *cobra.Command, args []string) {
	tpl := outputTemplate(cmd, tradeListTemplates, nil)
	out, closeOut := openOutput(cmd)
	defer closeOut()
	err := eachProposal(tradesDir(dataDir(cmd)), func(name string, tr trade.Trade) error {
		return tpl.Execute(out, &tradeInfo{Name: name, Trade: tr})
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
	out, closeOut := openOutput(cmd)
	defer closeOut()
	if err = yaml.NewEncoder(out).Encode(prop); err != nil {
		errorExit(ecCantExportProposal, err)
	}
}

func cmdAcceptProposal(cmd *cobra.Command, args []string) {
	in, inClose := openInput(cmd)
	defer inClose()
	newTrade := trade.NewOnChainSell()
	th := trade.NewHandler(trade.DefaultStageHandlers)
	th.InstallStageHandler(stages.ReceiveProposal, func(tr trade.Trade) error {
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
	})
	th.InstallStageHandler(stages.SendProposalResponse, func(tr trade.Trade) error {
		return trade.ErrInterruptTrade
	})
	for _, i := range th.Unhandled(newTrade.Stager().Stages()...) {
		th.InstallStageHandler(i, trade.NoOpHandler)
	}
	if err := th.HandleTrade(newTrade); err != nil && err != trade.ErrInterruptTrade {
		errorExit(ecCantCreateTrade, err)
	}
	if err := saveTrade(cmd, args[0], newTrade); err != nil {
		errorExit(ecCantCreateTrade, err)
	}
}
