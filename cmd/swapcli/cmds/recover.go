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

var _fee = &flagutil.FeeFlag{}

func init() {
	network := &_network
	flagutil.AddFlags(flagutil.FlagFuncMap{
		listRecoverableCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddVerbose,
			flagutil.AddFormat,
			flagutil.AddOutput,
		},
		recoverToAddressCmd.Flags(): []flagutil.FlagFunc{
			network.AddFlag,
			_fee.AddFlag,
			flagutil.AddRPC,
			flagutil.AddOutput,
			flagutil.AddVerbose,
		},
	})
	cmdutil.AddCommands(RecoverCmd, []*cobra.Command{
		listRecoverableCmd,
		recoverToAddressCmd,
	})
}

func listRecoverable(td string, out io.Writer, tpl *template.Template) error {
	return eachTrade(td, func(name string, tr trade.Trade) error {
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
}

func cmdListRecoverable(cmd *cobra.Command, args []string) {
	fs := cmd.Flags()
	out, closeOut := flagutil.MustOpenOutput(fs)
	defer closeOut()
	tpl := tplutil.MustOpenTemplate(fs, tradeListTemplates, nil)
	if err := listRecoverable(tradesDir(cmd), out, tpl); err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
}

func recoverFunds(tr trade.Trade, cl cryptocore.Client, cryptoInfo *trade.TraderInfo, addr string, feeFixed bool, fee uint64, out io.Writer, verboseRaw bool) error {
	addrScript, err := networks.
		AllByName[cryptoInfo.Crypto.Name][_network.MustNetwork(cryptoInfo.Crypto.Name)].
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
	if verboseRaw {
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
	tr := mustOpenTrade(cmd, args[0])
	out, closeOut := flagutil.MustOpenOutput(cmd.Flags())
	defer closeOut()
	fs := cmd.Flags()
	var verboseRaw bool
	if flagutil.MustVerboseLevel(fs, 1) > 0 {
		verboseRaw = true
	}
	err := recoverFunds(
		tr,
		mustNewClient(
			tr.OwnInfo().Crypto,
			flagutil.MustRPCAddress(fs),
			flagutil.MustRPCUsername(fs),
			flagutil.MustRPCPassword(fs),
			flagutil.MustRPCTLSConfig(fs),
		),
		tr.OwnInfo(),
		args[1],
		_fee.Fixed,
		_fee.Value,
		out,
		verboseRaw,
	)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
}
