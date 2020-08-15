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
	out, closeOut := openOutput(fs)
	defer closeOut()
	tpl := outputTemplate(fs, tradeListTemplates, nil)
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

func cmdRecoverToAddress(cmd *cobra.Command, args []string) {
	tr := openTrade(cmd, args[0])
	fs := cmd.Flags()
	addrScript, err := networks.AllByName[tr.OwnInfo().Crypto.Name][flagCryptoChain(tr.OwnInfo().Crypto)].
		AddressToScript(args[1])
	if err != nil {
		errorExit(ecCantRecover, err)
	}
	var recoveryFunc func([]byte, uint64) (tx.Tx, error)
	if flagFeeFixed(fs) {
		recoveryFunc = tr.RecoveryTxFixedFee
	} else {
		recoveryFunc = tr.RecoveryTx
	}
	tx, err := recoveryFunc(addrScript, flagFee(fs))
	if err != nil {
		errorExit(ecCantRecover, err)
	}
	b, err := tx.Serialize()
	if err != nil {
		errorExit(ecCantRecover, err)
	}
	out, closeOut := openOutput(fs)
	defer closeOut()
	if verboseLevel(fs, 1) > 0 {
		fmt.Fprintf(out, "raw transaction: %s\n", hex.EncodeToString(b))
	}
	txID, err := newClient(fs, tr.OwnInfo().Crypto).SendRawTransaction(b)
	if err != nil {
		errorExit(ecCantRecover, err)
	}
	fmt.Fprintf(out, "funds recovered (tx id): %s\n", txID.Hex())
}
