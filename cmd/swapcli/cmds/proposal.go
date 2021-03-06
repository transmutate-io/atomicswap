package cmds

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/internal/cmdutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil/exitcodes"
	"github.com/transmutate-io/atomicswap/internal/tplutil"
	"github.com/transmutate-io/atomicswap/roles"
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
	flagutil.AddFlags(flagutil.FlagFuncMap{
		listProposalsCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddFormat,
			flagutil.AddVerbose,
			flagutil.AddOutput,
		},
		acceptProposalCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddInput,
		},
		exportProposalCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddOutput,
		},
	})
	cmdutil.AddCommands(ProposalCmd, []*cobra.Command{
		listProposalsCmd,
		exportProposalCmd,
		acceptProposalCmd,
	})
}

func listProposals(td string, out io.Writer, tpl *template.Template) error {
	return eachProposal(td, func(name string, tr trade.Trade) error {
		return tpl.Execute(out, newTradeInfo(name, tr))
	})
}

func cmdListProposals(cmd *cobra.Command, args []string) {
	tpl := tplutil.MustOpenTemplate(cmd.Flags(), tradeListTemplates, nil)
	out, closeOut := flagutil.MustOpenOutput(cmd.Flags())
	defer closeOut()
	if err := listProposals(tradesDir(cmd), out, tpl); err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
}

func exportProposal(tr trade.Trade, out io.Writer) error {
	if tr.Role() != roles.Buyer {
		cmdutil.ErrorExit(exitcodes.NotABuyer)
	}
	btr, err := tr.Buyer()
	if err != nil {
		return err
	}
	prop, err := btr.GenerateBuyProposal()
	if err != nil {
		return err
	}
	return yaml.NewEncoder(out).Encode(prop)
}

func cmdExportProposal(cmd *cobra.Command, args []string) {
	out, closeOut := flagutil.MustOpenOutput(cmd.Flags())
	defer closeOut()
	if err := exportProposal(mustOpenTrade(cmd, args[0]), out); err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
}

func acceptProposal(tp string, name string, prop *trade.BuyProposal) error {
	newTrade, err := trade.AcceptProposal(prop)
	if err != nil {
		return err
	}
	return saveTrade(filepath.Join(tp, filepath.FromSlash(name)), newTrade)
}

func cmdAcceptProposal(cmd *cobra.Command, args []string) {
	in, inClose := flagutil.MustOpenInput(cmd.Flags())
	defer inClose()
	b, err := ioutil.ReadAll(in)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
	prop, err := trade.UnamrshalBuyProposal(b)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
	if err = acceptProposal(tradesDir(cmd), args[0], prop); err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
}
