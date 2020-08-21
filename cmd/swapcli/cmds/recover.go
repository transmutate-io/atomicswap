package cmds

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/networks"
	"github.com/transmutate-io/atomicswap/stages"
	"github.com/transmutate-io/atomicswap/trade"
	"github.com/transmutate-io/atomicswap/tx"
	"github.com/transmutate-io/cryptocore"
)

var (
	RecoverCmd = &cobra.Command{
		Use:     "recover",
		Short:   "recovery commands",
		Aliases: []string{"rec"},
	}
	listRecoverableCmd = &cobra.Command{
		Use:     "list",
		Short:   "list recoverable trades",
		Aliases: []string{"l", "ls"},
		Args:    cobra.NoArgs,
		Run:     cmdListRecoverable,
	}
	recoverToAddressCmd = &cobra.Command{
		Use:     "toaddress <name> <address>",
		Short:   "toaddress recovers the funds to the provided address",
		Aliases: []string{"t", "to"},
		Args:    cobra.ExactArgs(2),
		Run:     cmdRecoverToAddress,
	}
)

func init() {
	addFlags(flagMap{
		listRecoverableCmd.Flags(): []flagFunc{
			addFlagVerbose,
			addFlagFormat,
			addFlagOutput,
		},
		recoverToAddressCmd.Flags(): []flagFunc{
			addFlagCryptoChain,
			addFlagFee,
			addFlagsRPC,
			addFlagOutput,
			addFlagVerbose,
		},
	})
	addCommands(RecoverCmd, []*cobra.Command{
		listRecoverableCmd,
		recoverToAddressCmd,
	})
}

func cmdListRecoverable(cmd *cobra.Command, args []string) {
	fs := cmd.Flags()
	out, closeOut := mustOpenOutput(fs)
	defer closeOut()
	tpl := mustOutputTemplate(fs, tradeListTemplates, nil)
	err := eachTrade(tradesDir(cmd), func(name string, tr trade.Trade) error {
		for _, i := range tr.Stager().Stages() {
			if i == stages.LockFunds {
				return nil
			}
		}
		if _, ok := tr.RecoverableFunds().Funds().([]*trade.Output); !ok {
			return nil
		}
		return tpl.Execute(out, newTradeInfo(name, tr))
	})
	if err != nil {
		errorExit(ecCantListTrades)
	}
}

func recoverFunds(tr trade.Trade, cl cryptocore.Client, cryptoInfo *trade.TraderInfo, addr string, feeFixed bool, fee uint64, out io.Writer, verbose int) error {
	addrScript, err := networks.AllByName[cryptoInfo.Crypto.Name][mustFlagCryptoChain(cryptoInfo.Crypto)].
		AddressToScript(addr)
	if err != nil {
		return err
	}
	var recoveryFunc func([]byte, uint64) (tx.Tx, error)
	if feeFixed {
		recoveryFunc = tr.RecoveryTxFixedFee
	} else {
		recoveryFunc = tr.RecoveryTx
	}
	tx, err := recoveryFunc(addrScript, fee)
	if err != nil {
		return err
	}
	b, err := tx.Serialize()
	if err != nil {
		return err
	}
	if verbose > 0 {
		fmt.Fprintf(out, "raw transaction: %s\n", hex.EncodeToString(b))
	}
	txID, err := cl.SendRawTransaction(b)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "funds recovered (tx id): %s\n", txID.Hex())
	return nil
}

func cmdRecoverToAddress(cmd *cobra.Command, args []string) {
	tr := openTrade(cmd, args[0])
	out, closeOut := mustOpenOutput(cmd.Flags())
	defer closeOut()
	fs := cmd.Flags()
	err := recoverFunds(
		tr,
		newClient(fs, tr.OwnInfo().Crypto),
		tr.OwnInfo(),
		args[1],
		flagFeeFixed(fs),
		flagFee(fs),
		out,
		mustVerboseLevel(fs, 1),
	)
	if err != nil {
		errorExit(ecCantRecover, err)
	}
}
