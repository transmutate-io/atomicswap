package cmds

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/networks"
	"github.com/transmutate-io/atomicswap/stages"
	"github.com/transmutate-io/atomicswap/trade"
	"github.com/transmutate-io/atomicswap/tx"
)

var (
	RedeemCmd = &cobra.Command{
		Use:     "redeem",
		Short:   "redeem commands",
		Aliases: []string{"r", "red"},
	}
	listRedeeamableCmd = &cobra.Command{
		Use:     "list",
		Short:   "list redeeamable trades",
		Aliases: []string{"l", "ls"},
		Args:    cobra.NoArgs,
		Run:     cmdListRedeeamable,
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
	fs := listRedeeamableCmd.Flags()
	addFlagVerbose(fs)
	addFlagFormat(fs)
	addFlagOutput(fs)
	fs = redeemToAddressCmd.Flags()
	addFlagCryptoChain(fs)
	addFlagFee(fs)
	addFlagsRPC(fs)
	addFlagOutput(fs)
	addFlagVerbose(fs)
	for _, i := range []*cobra.Command{
		listRedeeamableCmd,
		redeemToAddressCmd,
	} {
		RedeemCmd.AddCommand(i)
	}
}

func cmdListRedeeamable(cmd *cobra.Command, args []string) {
	out, closeOut := openOutput(cmd.Flags())
	defer closeOut()
	tpl := outputTemplate(cmd.Flags(), tradeListTemplates, nil)
	err := eachTrade(tradesDir(cmd), func(name string, tr trade.Trade) error {
		if tr.Stager().Stage() != stages.RedeemFunds {
			return nil
		}
		return tpl.Execute(out, newTradeInfo(name, tr))
	})
	if err != nil {
		errorExit(ecCantListTrades, err)
	}
}

func cmdRedeemToAddress(cmd *cobra.Command, args []string) {
	tr := openTrade(cmd, args[0])
	th := trade.NewHandler(trade.StageHandlerMap{
		stages.RedeemFunds: func(tr trade.Trade) error {
			out, closeOut := openOutput(cmd.Flags())
			defer closeOut()
			addrScript, err := networks.AllByName[tr.TraderInfo().Crypto.Name][flagCryptoChain(tr.TraderInfo().Crypto)].
				AddressToScript(args[1])
			if err != nil {
				return err
			}
			fs := cmd.Flags()
			var redeemFunc func([]byte, uint64) (tx.Tx, error)
			if flagFeeFixed(fs) {
				redeemFunc = tr.RedeemTxFixedFee
			} else {
				redeemFunc = tr.RedeemTx
			}
			tx, err := redeemFunc(addrScript, flagFee(fs))
			if err != nil {
				return err
			}
			b, err := tx.Serialize()
			if err != nil {
				return err
			}
			if verboseLevel(fs, 1) > 0 {
				fmt.Fprintf(out, "raw transaction: %s\n", hex.EncodeToString(b))
			}
			cl := newClient(cmd.Flags(), tr.TraderInfo().Crypto)
			txID, err := cl.SendRawTransaction(b)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "funds redeemed (tx id): %s\n", txID.Hex())
			return nil
		},
	})
	for _, i := range th.Unhandled(tr.Stager().Stages()...) {
		th.InstallStageHandler(i, trade.NoOpHandler)
	}
	if err := th.HandleTrade(tr); err != nil {
		errorExit(ecCantRedeem, err)
	}
	saveTrade(cmd, args[0], tr)
}
