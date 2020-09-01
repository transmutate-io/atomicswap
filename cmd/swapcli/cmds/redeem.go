package cmds

import (
	"encoding/hex"
	"fmt"
	"io"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/internal/cmdutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil/exitcodes"
	"github.com/transmutate-io/atomicswap/internal/tplutil"
	"github.com/transmutate-io/atomicswap/networks"
	"github.com/transmutate-io/atomicswap/stages"
	"github.com/transmutate-io/atomicswap/trade"
	"github.com/transmutate-io/atomicswap/tx"
	"github.com/transmutate-io/cryptocore"
)

var (
	RedeemCmd = &cobra.Command{
		Use:     "redeem",
		Short:   "redeem commands",
		Aliases: []string{"r", "red"},
	}
	listRedeemableCmd = &cobra.Command{
		Use:     "list",
		Short:   "list redeemable trades",
		Aliases: []string{"l", "ls"},
		Args:    cobra.NoArgs,
		Run:     cmdListRedeemable,
	}
	redeemToAddressCmd = &cobra.Command{
		Use:     "toaddress <name> <address>",
		Short:   "toaddress redeems the funds to the provided address",
		Aliases: []string{"t", "to"},
		Args:    cobra.ExactArgs(2),
		Run:     cmdRedeemToAddress,
	}
)

func init() {
	network := &_network
	flagutil.AddFlags(flagutil.FlagFuncMap{
		listRedeemableCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddVerbose,
			flagutil.AddFormat,
			flagutil.AddOutput,
		},
		redeemToAddressCmd.Flags(): []flagutil.FlagFunc{
			network.AddFlag,
			_fee.AddFlag,
			flagutil.AddRPC,
			flagutil.AddOutput,
			flagutil.AddVerbose,
		},
	})
	cmdutil.AddCommands(RedeemCmd, []*cobra.Command{
		listRedeemableCmd,
		redeemToAddressCmd,
	})
}

func listRedeemable(td string, out io.Writer, tpl *template.Template) error {
	return eachTrade(td, func(name string, tr trade.Trade) error {
		if tr.Stager().Stage() != stages.RedeemFunds {
			return nil
		}
		return tpl.Execute(out, newTradeInfo(name, tr))
	})
}

func cmdListRedeemable(cmd *cobra.Command, args []string) {
	out, closeOut := flagutil.MustOpenOutput(cmd.Flags())
	defer closeOut()
	tpl := tplutil.MustOpenTemplate(cmd.Flags(), tradeListTemplates, nil)
	err := listRedeemable(tradesDir(cmd), out, tpl)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
}

func newRedeemHandler(
	addr string,
	feeFixed bool,
	fee uint64,
	out io.Writer,
	verboseRaw bool,
	cl cryptocore.Client,
) func(trade.Trade) error {
	return func(tr trade.Trade) error {
		addrScript, err := networks.
			AllByName[tr.TraderInfo().Crypto.Name][_network.MustNetwork(tr.TraderInfo().Crypto.Name)].
			AddressToScript(addr)
		if err != nil {
			return err
		}
		var redeemFunc func([]byte, uint64) (tx.Tx, error)
		if feeFixed {
			redeemFunc = tr.RedeemTxFixedFee
		} else {
			redeemFunc = tr.RedeemTx
		}
		tx, err := redeemFunc(addrScript, fee)
		if err != nil {
			return err
		}
		b, err := tx.Serialize()
		if err != nil {
			return err
		}
		if verboseRaw {
			fmt.Fprintf(out, "raw transaction: %s\n", hex.EncodeToString(b))
		}

		txID, err := cl.SendRawTransaction(b)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "funds redeemed (tx id): %s\n", txID.Hex())
		return nil
	}
}

func redeemToAddress(
	tr trade.Trade,
	out io.Writer,
	destAddr string,
	addr string,
	username string,
	password string,
	tlsConf *cryptocore.TLSConfig,
	fee uint64,
	fixedFee bool,
) error {
	cl, err := newClient(tr.TraderInfo().Crypto, addr, username, password, tlsConf)
	if err != nil {
		return err
	}
	th := trade.NewHandler(trade.StageHandlerMap{
		stages.RedeemFunds: newRedeemHandler(destAddr, fixedFee, fee, out, true, cl),
	})
	for _, i := range th.Unhandled(tr.Stager().Stages()...) {
		th.InstallStageHandler(i, trade.NoOpHandler)
	}
	return th.HandleTrade(tr)
}

func cmdRedeemToAddress(cmd *cobra.Command, args []string) {
	tr := mustOpenTrade(cmd, args[0])
	out, closeOut := flagutil.MustOpenOutput(cmd.Flags())
	defer closeOut()
	fs := cmd.Flags()
	err := redeemToAddress(
		tr,
		out,
		args[1],
		flagutil.MustRPCAddress(fs),
		flagutil.MustRPCUsername(fs),
		flagutil.MustRPCPassword(fs),
		flagutil.MustRPCTLSConfig(fs),
		_fee.Value,
		_fee.Fixed,
	)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
	mustSaveTrade(cmd, args[0], tr)
}
