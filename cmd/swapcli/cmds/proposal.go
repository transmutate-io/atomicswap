package cmds

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
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
		Short:   "list trade names with proposals",
		Aliases: []string{"ls", "l"},
		Run:     cmdListProposals,
	}
	exportProposalCmd = &cobra.Command{
		Use:     "export <name>",
		Short:   "export proposal",
		Aliases: []string{"exp", "e"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdExportProposal,
	}
	acceptProposalCmd = &cobra.Command{
		Use:     "accept <name>",
		Short:   "accept proposal",
		Aliases: []string{"a"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdAcceptProposal,
	}
)

func init() {
	acceptProposalCmd.Flags().StringP("name", "n", "", "set trade name")
	for _, i := range []*cobra.Command{
		listProposalsCmd,
		exportProposalCmd,
		acceptProposalCmd,
	} {
		ProposalCmd.AddCommand(i)
	}
}

func cmdListProposals(cmd *cobra.Command, args []string) {
	out, closeOut := openOutput(cmd)
	defer closeOut()
	err := eachProposal(tradesDir(dataDir(cmd)), func(name string, tr trade.Trade) error {
		fmt.Fprintln(out, name)
		return nil
	})
	if err != nil {
		errorExit(ECCantListProposal, "can't list proposals: %#v\n", err)
	}
}

func cmdExportProposal(cmd *cobra.Command, args []string) {
	tr, err := openTrade(cmd, args[0])
	if err != nil {
		errorExit(ECCantFindTrade, "can't open trade: %#v\n", err)
	}
	prop, err := tr.GenerateBuyProposal()
	if err != nil {
		errorExit(ECCantOpenProposal, "can't open proposal: %#v\n", err)
	}
	out, closeOut := openOutput(cmd)
	defer closeOut()
	if err = yaml.NewEncoder(out).Encode(prop); err != nil {
		errorExit(ECCantExportProposal, "can't export proposal: %#v\n", err)
	}
}

func cmdAcceptProposal(cmd *cobra.Command, args []string) {
	in, inClose := openInput(cmd)
	defer inClose()
	newTrade := trade.NewOnChainSell()
	th := trade.NewHandler(trade.DefaultStageHandlers)
	th.InstallStageHandlers(trade.StageHandlerMap{
		stages.ReceiveProposal: func(tr trade.Trade) error {
			b, err := ioutil.ReadAll(in)
			if err != nil {
				return err
			}
			prop, err := trade.UnamrshalBuyProposal(b)
			if err != nil {
				return err
			}
			if err := tr.AcceptBuyProposal(prop); err != nil {
				return err
			}
			return nil
		},
		stages.SendProposalResponse: func(tr trade.Trade) error {
			return trade.ErrInterruptTrade
		},
	})
	for _, i := range th.Unhandled(newTrade.Stager().Stages()...) {
		th.InstallStageHandler(i, trade.NoOpHandler)
	}
	if err := th.HandleTrade(newTrade); err != nil && err != trade.ErrInterruptTrade {
		errorExit(ECCantCreateTrade, "can't create new trade: %#v\n", err)
	}
	if err := saveTrade(cmd, args[0], newTrade); err != nil {
		errorExit(ECCantCreateTrade, "can't create new trade: %#v\n", err)
	}
}
