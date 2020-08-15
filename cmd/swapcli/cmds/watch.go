package cmds

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/roles"
	"github.com/transmutate-io/atomicswap/stages"
	"github.com/transmutate-io/atomicswap/trade"
	"github.com/transmutate-io/cryptocore"
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
	addFlags(flagMap{
		listWatchableCmd.Flags(): []flagFunc{
			addFlagFormat,
			addFlagVerbose,
			addFlagOutput,
		},
		watchOwnDepositCmd.Flags(): []flagFunc{
			addFlagsRPC,
			addFlagFirstBlock,
			addFlagFormat,
			addFlagVerbose,
			addFlagOutput,
			addFlagCryptoChain,
			addFlagIgnoreTarget,
			addFlagConfirmations,
		},
		watchTraderDepositCmd.Flags(): []flagFunc{
			addFlagsRPC,
			addFlagFirstBlock,
			addFlagFormat,
			addFlagVerbose,
			addFlagOutput,
			addFlagCryptoChain,
			addFlagIgnoreTarget,
			addFlagConfirmations,
		},
		watchSecretTokenCmd.Flags(): []flagFunc{
			addFlagsRPC,
			addFlagFirstBlock,
			addFlagFormat,
			addFlagVerbose,
			addFlagOutput,
			addFlagCryptoChain,
			addFlagIgnoreTarget,
		},
	})
	addCommands(WatchCmd, []*cobra.Command{
		listWatchableCmd,
		watchOwnDepositCmd,
		watchTraderDepositCmd,
		watchSecretTokenCmd,
	})
}

func cmdListWatchable(cmd *cobra.Command, args []string) {
	out, closeOut := openOutput(cmd.Flags())
	defer closeOut()
	tpl := outputTemplate(cmd.Flags(), watchableTradesTemplates, nil)
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

func containsString(s []string, v string) bool {
	for _, i := range s {
		if i == v {
			return true
		}
	}
	return false
}

func outputID(tx []byte, n uint64) string { return fmt.Sprintf("%s:%d", hex.EncodeToString(tx), n) }

func newDepositWatcher(
	out io.Writer,
	outputTpl *template.Template,
	blockTpl *template.Template,
	cl cryptocore.Client,
	firstBlock uint64,
	cryptoInfo *trade.TraderInfo,
	funds trade.FundsData,
	ignoreTarget bool,
	minConfirmations uint64,
	wd *watchData,
	bwd *blockWatchData,
	tradeSave func(trade.Trade),
	wdSave func(*watchData),
) func(trade.Trade) error {
	return func(tr trade.Trade) error {
		sig := make(chan os.Signal, 0)
		signal.Notify(sig, os.Interrupt, os.Kill)
		targetAmount := cryptoInfo.Amount.UInt64(cryptoInfo.Crypto.Decimals)
		depositAddr, err := funds.Lock().Address(flagCryptoChain(cryptoInfo.Crypto))
		if err != nil {
			return err
		}
		outputs, ok := funds.Funds().([]*trade.Output)
		if !ok {
			return errors.New("not implemented")
		}
		outMap := make(map[string]uint64, len(outputs))
		totalAmount := uint64(0)
		for _, i := range outputs {
			txID := outputID(i.TxID, uint64(i.N))
			outMap[txID] = i.Amount
			totalAmount += i.Amount
			err := outputTpl.Execute(out, newOutputInfo(
				"known output",
				txID,
				cryptoInfo.Crypto,
				types.NewAmount(i.Amount, uint64(cryptoInfo.Crypto.Decimals)),
				types.NewAmount(totalAmount, uint64(cryptoInfo.Crypto.Decimals)),
				cryptoInfo.Amount,
			))
			if err != nil {
				return err
			}
		}
		if !ignoreTarget && totalAmount >= targetAmount {
			return nil
		}
		bdc, errc, closeIter := iterateBlocks(cl, bwd, firstBlock)
		defer closeIter()
		var tradeChanged bool
		for {
			select {
			case err := <-errc:
				return err
			case <-sig:
				return trade.ErrInterruptTrade
			case bd := <-bdc:
				if err := blockTpl.Execute(out, newBlockInfo(bd.height, len(bd.txs))); err != nil {
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
						amount := j.Value().UInt64(cryptoInfo.Crypto.Decimals)
						funds.AddFunds(&trade.Output{
							TxID:   i.ID(),
							N:      uint32(j.N()),
							Amount: amount,
						})
						outMap[outID] = amount
						totalAmount += amount
						err = outputTpl.Execute(out, newOutputInfo(
							"new output found",
							outID,
							cryptoInfo.Crypto,
							j.Value(),
							types.NewAmount(totalAmount, uint64(cryptoInfo.Crypto.Decimals)),
							cryptoInfo.Amount,
						))
						if err != nil {
							return err
						}
						tradeChanged = true
						if !ignoreTarget && totalAmount >= targetAmount {
							break
						}
					}
				}
				if bd.height > bwd.Top {
					bwd.Top = bd.height
				}
				if bwd.Bottom == 0 || bd.height < bwd.Bottom {
					bwd.Bottom = bd.height
				}
				wdSave(wd)
				if tradeChanged {
					tradeSave(tr)
					tradeChanged = false
				}
				if !ignoreTarget && totalAmount >= targetAmount {
					return nil
				}
			}
		}
	}
}

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
	fs := cmd.Flags()
	out, closeOut := openOutput(fs)
	defer closeOut()
	wd := openWatchData(cmd, tradeName)
	cryptoInfo := selectCryptoInfo(tr)
	th := trade.NewHandler(trade.StageHandlerMap{
		watchStage: newDepositWatcher(
			out,
			outputTemplate(fs, depositChunkLogTemplates, nil),
			outputTemplate(fs, blockInspectionTemplates, nil),
			newClient(fs, cryptoInfo.Crypto),
			flagFirstBlock(fs),
			cryptoInfo,
			selectFunds(tr),
			flagIgnoreTarget(fs),
			flagConfirmations(fs),
			wd,
			selectWatchData(wd),
			func(t trade.Trade) { saveTrade(cmd, tradeName, t) },
			func(nwd *watchData) { saveWatchData(cmd, tradeName, nwd) },
		),
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

func newSecretTokenWatcher(cl cryptocore.Client, firstBlock uint64, wd *watchData) func(trade.Trade) error {
	return func(tr trade.Trade) error {
		sig := make(chan os.Signal, 0)
		signal.Notify(sig, os.Interrupt, os.Kill)
		bdc, errc, closeIter := iterateBlocks(cl, wd.Own, firstBlock)
		defer closeIter()
		for {
			select {
			case <-sig:
				return nil
			case err := <-errc:
				return err
			case db := <-bdc:
				for _, i := range db.txs {
					token, err := extractToken(
						tr.OwnInfo().Crypto,
						i,
						tr.RecoverableFunds().Lock(),
					)
					if err != nil {
						return err
					}
					if token == nil {
						continue
					}
					tr.SetToken(token)
					return nil
				}
			}
		}
	}
}

func cmdWatchSecretToken(cmd *cobra.Command, args []string) {
	tr := openTrade(cmd, args[0])
	fs := cmd.Flags()
	cl := newClient(fs, tr.OwnInfo().Crypto)
	wd := openWatchData(cmd, args[0])
	th := trade.NewHandler(nil)
	th.InstallStageHandlers(trade.StageHandlerMap{
		stages.WaitFundsRedeemed: newSecretTokenWatcher(cl, flagFirstBlock(fs), wd),
		stages.RedeemFunds:       trade.InterruptHandler,
	})
	for _, i := range th.Unhandled(tr.Stager().Stages()...) {
		th.InstallStageHandler(i, trade.NoOpHandler)
	}
	if err := th.HandleTrade(tr); err != nil && err != trade.ErrInterruptTrade {
		errorExit(ecFailedToWatch, err)
	}
	saveTrade(cmd, args[0], tr)
}
