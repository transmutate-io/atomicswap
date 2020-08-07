package cmds

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/roles"
	"github.com/transmutate-io/atomicswap/stages"
	"github.com/transmutate-io/atomicswap/trade"
	"github.com/transmutate-io/cryptocore/types"
)

var (
	WatchCmd = &cobra.Command{
		Use:     "watch <command>",
		Short:   "blockchain watch commands",
		Aliases: []string{"w"},
	}
	listWatchableCmd = &cobra.Command{
		Use:     "list",
		Short:   "list trades in watchable states to output",
		Aliases: []string{"ls", "l"},
		Run:     cmdListWatchable,
	}
	watchOwnDepositCmd = &cobra.Command{
		Use:     "own <trade_name>",
		Short:   "watch own deposit for a trade",
		Aliases: []string{"o"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdWatchOwnDeposit,
	}
	watchTraderDepositCmd = &cobra.Command{
		Use:     "trader <trade_name>",
		Short:   "watch trader deposit for a trade",
		Aliases: []string{"t"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdWatchTraderDeposit,
	}
	watchSecretTokenCmd = &cobra.Command{
		Use:     "secret <trade_name>",
		Short:   "watch for a deposit redeem and collect the secret token",
		Aliases: []string{"s"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdWatchSecretToken,
	}
)

func init() {
	fs := listWatchableCmd.Flags()
	addFlagFormat(fs)
	addFlagVerbose(fs)
	addFlagOutput(fs)
	_cmds := []*cobra.Command{
		listWatchableCmd,
		watchOwnDepositCmd,
		watchTraderDepositCmd,
		watchSecretTokenCmd,
	}
	for _, i := range _cmds[1:] {
		fs := i.Flags()
		addFlagsRPC(fs)
		addFlagFirstBlock(fs)
		addFlagFormat(fs)
		addFlagVerbose(fs)
		addFlagOutput(fs)
		addFlagCryptoChain(fs)
	}
	// for _, i := range _cmds[1:3] {
	// 	fs := i.Flags()
	// 	addFlagConfirmations(fs)
	// 	addFlagIgnoreTarget(fs)
	// }
	for _, i := range _cmds {
		WatchCmd.AddCommand(i)
	}
}

func cmdListWatchable(cmd *cobra.Command, args []string) {
	out, closeOut := openOutput(cmd)
	defer closeOut()
	tpl := outputTemplate(cmd, watchableTradesTemplates, nil)
	eachTrade(tradesDir(cmd), func(name string, tr trade.Trade) error {
		switch tr.Stager().Stage() {
		case stages.SendProposalResponse,
			stages.LockFunds,
			stages.WaitLockedFunds,
			stages.WaitFundsRedeemed,
			stages.RedeemFunds:
		default:
			return nil
		}
		return tpl.Execute(out, newTradeInfo(name, tr))
	})
}

var watchableTradesTemplates = []string{
	`{{ .name }}: {{ $s := .trade.Stager.Stage.String }}{{ if eq $s "lock-funds" -}}
	wait for own funds deposit
	{{- else if or (eq $s "send-proposal-response") (eq $s "wait-locked-funds") -}}
	wait for trader funds deposit
	{{- else -}}
	wait for secret token (trader redeem)
	{{- end }}
`,
}

func newOutputInfo(prefix string, id string, crypto *cryptos.Crypto, amount, total, target types.Amount) templateData {
	return templateData{
		"prefix": prefix,
		"id":     id,
		"crypto": crypto,
		"amount": amount,
		"total":  total,
		"target": target,
	}
}

func newBlockInfo(id string, height uint64, txCount int) templateData {
	return templateData{
		"id":      id,
		"height":  height,
		"txCount": txCount,
	}
}

var (
	blockInspectionTemplates = []string{
		"",
		"inspecting block {{ .id }} at height {{ .height }}\n",
		"inspecting block {{ .id }} at height {{ .height }} ({{ .txCount }} transactions)\n",
	}
	depositChunkLogTemplates = []string{
		`{{ if ne .prefix "" }}{{ .prefix }}: {{ end -}}
{{ .id }} {{ .amount }} {{ .crypto.Short }}
`,
		`{{ if ne .prefix "" }}{{ .prefix }}: {{ end -}}
{{ .id }} {{ .amount }} {{ .crypto.Short }} ({{ .total }} {{.crypto.Short }} total)
`,
		`{{ if ne .prefix "" }}{{ .prefix }}: {{ end -}}
{{ .id }} {{ .amount }} {{ .crypto.Short }} ({{ .total }} of {{ .target }} {{.crypto.Short }})
`,
	}
)

func containsString(s []string, v string) bool {
	for _, i := range s {
		if i == v {
			return true
		}
	}
	return false
}

func outputID(tx []byte, n uint64) string { return fmt.Sprintf("%s:%d", hex.EncodeToString(tx), n) }

func cmdWatchDeposit(
	cmd *cobra.Command,
	tradeName string,
	watchStage stages.Stage,
	selectCryptoInfo func(trade.Trade) *trade.TraderInfo,
	selectWatchData func(*watchData) *blockWatchData,
	selectFunds func(trade.Trade) trade.FundsData,
	selectInterruptStage func(trade.Trade) stages.Stage,
) {
	tr := openTrade(cmd, tradeName)
	th := trade.NewHandler(trade.StageHandlerMap{
		watchStage: func(t trade.Trade) error {
			sig := make(chan os.Signal, 0)
			signal.Notify(sig, os.Interrupt, os.Kill)
			cryptoInfo := selectCryptoInfo(tr)
			crypto := cryptoInfo.Crypto
			cl := newClient(cmd, crypto)
			wd := openWatchData(cmd, tradeName)
			fs := cmd.Flags()
			bwd := selectWatchData(wd)
			// minConfirmations := flagConfirmations(fs)
			funds := selectFunds(tr)
			outputs, ok := funds.Funds().([]*trade.Output)
			if !ok {
				return errors.New("not implemented")
			}
			out, closeOut := openOutput(cmd)
			defer closeOut()
			outputTpl := outputTemplate(cmd, depositChunkLogTemplates, nil)
			blockTpl := outputTemplate(cmd, blockInspectionTemplates, nil)
			outMap := make(map[string]uint64, len(outputs))
			totalAmount := uint64(0)
			targetAmount := cryptoInfo.Amount.UInt64(crypto.Decimals)
			depositAddr, err := funds.Lock().Address(flagCryptoChain(cmd, crypto))
			ignoreTarget := flagIgnoreTarget(fs)
			if err != nil {
				return err
			}
			for _, i := range outputs {
				txID := outputID(i.TxID, uint64(i.N))
				outMap[txID] = i.Amount
				totalAmount += i.Amount
				err := outputTpl.Execute(out, newOutputInfo(
					"known output",
					txID,
					crypto,
					types.NewAmount(i.Amount, uint64(crypto.Decimals)),
					types.NewAmount(totalAmount, uint64(crypto.Decimals)),
					cryptoInfo.Amount,
				))
				if err != nil {
					return err
				}
			}
			if !ignoreTarget && totalAmount >= targetAmount {
				return nil
			}
			bdc, errc, closeIter := iterateBlocks(cl, bwd, flagFirstBlock(fs))
			var tradeChanged bool
			for {
				select {
				case err := <-errc:
					return err
				case <-sig:
					closeIter()
					return trade.ErrInterruptTrade
				case bd := <-bdc:
					err = blockTpl.Execute(out, newBlockInfo(bd.hash.Hex(), bd.height, len(bd.txs)))
					if err != nil {
						return err
					}
					for _, i := range bd.txs {
						txUtxo, ok := i.UTXO()
						if !ok {
							return errors.New("not implemented")
						}
						for _, j := range txUtxo.Outputs() {
							if !containsString(j.LockScript().Addresses(), depositAddr) {
								continue
							}
							outID := outputID(i.ID(), uint64(j.N()))
							if _, ok := outMap[outID]; ok {
								continue
							}
							amount := j.Value().UInt64(crypto.Decimals)
							funds.AddFunds(&trade.Output{
								TxID:   i.ID(),
								N:      uint32(j.N()),
								Amount: amount,
							})
							outMap[outID] = amount
							totalAmount += amount
							err := outputTpl.Execute(out, newOutputInfo(
								"new output found",
								outID,
								crypto,
								j.Value(),
								types.NewAmount(totalAmount, uint64(crypto.Decimals)),
								cryptoInfo.Amount,
							))
							if err != nil {
								return err
							}
							tradeChanged = true
						}
					}
					if bd.height > bwd.Top {
						bwd.Top = bd.height
					}
					if bwd.Bottom == 0 || bd.height < bwd.Bottom {
						bwd.Bottom = bd.height
					}
					saveWatchData(cmd, tradeName, wd)
					if tradeChanged {
						saveTrade(cmd, tradeName, tr)
						tradeChanged = false
					}
					if !ignoreTarget && totalAmount >= targetAmount {
						return nil
					}
				}
			}
		},
		selectInterruptStage(tr): trade.InterruptHandler,
	})

	for _, i := range th.Unhandled(tr.Stager().Stages()...) {
		th.InstallStageHandler(i, trade.NoOpHandler)
	}
	if err := th.HandleTrade(tr); err != nil && err != trade.ErrInterruptTrade {
		errorExit(ecFailedToWatch, err)
	}
	saveTrade(cmd, tradeName, tr)
}

func cmdWatchOwnDeposit(cmd *cobra.Command, args []string) {
	cmdWatchDeposit(
		cmd,
		args[0],
		stages.LockFunds,
		func(tr trade.Trade) *trade.TraderInfo { return tr.OwnInfo() },
		func(wd *watchData) *blockWatchData { return wd.Own },
		func(tr trade.Trade) trade.FundsData { return tr.RecoverableFunds() },
		func(tr trade.Trade) stages.Stage {
			if tr.Role() == roles.Buyer {
				return stages.WaitLockedFunds
			}
			return stages.WaitFundsRedeemed
		},
	)
}

func cmdWatchTraderDeposit(cmd *cobra.Command, args []string) {
	cmdWatchDeposit(
		cmd,
		args[0],
		stages.WaitLockedFunds,
		func(tr trade.Trade) *trade.TraderInfo { return tr.TraderInfo() },
		func(wd *watchData) *blockWatchData { return wd.Trader },
		func(tr trade.Trade) trade.FundsData { return tr.RedeemableFunds() },
		func(tr trade.Trade) stages.Stage {
			if tr.Role() == roles.Buyer {
				return stages.RedeemFunds
			}
			return stages.LockFunds
		},
	)
}

func cmdWatchSecretToken(cmd *cobra.Command, args []string) {
	// tr := openTrade(cmd, args[0])
	// th := trade.NewHandler(nil)
	// th.InstallStageHandlers(trade.StageHandlerMap{
	// 	stages.WaitFundsRedeemed: func(tr trade.Trade) error {
	// 		// sig := make(chan os.Signal, 0)
	// 		// signal.Notify(sig, os.Interrupt, os.Kill)
	// 		return nil
	// 	},
	// })
	// for _, i := range th.Unhandled(tr.Stager().Stages()...) {
	// 	th.InstallStageHandler(i, trade.NoOpHandler)
	// }
	// if err := th.HandleTrade(tr); err != nil && err != trade.ErrInterruptTrade {
	// 	errorExit(ecFailedToWatch, err)
	// }
	// saveTrade(cmd, args[0], tr)
}
