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

func listWatchable(td string, out io.Writer, tpl *template.Template) error {
	return eachTrade(td, func(name string, tr trade.Trade) error {
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

func cmdListWatchable(cmd *cobra.Command, args []string) {
	out, closeOut := mustOpenOutput(cmd.Flags())
	defer closeOut()
	tpl := mustOutputTemplate(cmd.Flags(), watchableTradesTemplates, nil)
	if err := listWatchable(tradesDir(cmd), out, tpl); err != nil {
		errorExit(ecFailedToWatch, err)
	}
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
		depositAddr, err := funds.Lock().Address(mustFlagCryptoChain(cryptoInfo.Crypto))
		if err != nil {
			return err
		}
		fmt.Printf("watching deposit address: %s\n", depositAddr)
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

func watchDeposit(
	tr trade.Trade,
	wd *watchData,
	out io.Writer,
	depositTpl *template.Template,
	blockTpl *template.Template,
	cl cryptocore.Client,
	firstBlock uint64,
	ignoreTarget bool,
	confirmations uint64,
	watchStage stages.Stage,
	selectCryptoInfo func(trade.Trade) *trade.TraderInfo,
	selectWatchData func(*watchData) *blockWatchData,
	selectFunds func(trade.Trade) trade.FundsData,
	selectInterruptStage func(trade.Trade) stages.Stage,
	tradeSave func(trade.Trade),
	wdSave func(*watchData),
) error {
	cryptoInfo := selectCryptoInfo(tr)
	th := trade.NewHandler(trade.StageHandlerMap{
		watchStage: newDepositWatcher(
			out,
			depositTpl,
			blockTpl,
			cl,
			firstBlock,
			cryptoInfo,
			selectFunds(tr),
			ignoreTarget,
			confirmations,
			wd,
			selectWatchData(wd),
			tradeSave,
			wdSave,
		),
		selectInterruptStage(tr): trade.InterruptHandler,
	})
	for _, i := range th.Unhandled(tr.Stager().Stages()...) {
		th.InstallStageHandler(i, trade.NoOpHandler)
	}
	return th.HandleTrade(tr)
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
	tr := mustOpenTrade(cmd, tradeName)
	wd := mustOpenWatchData(cmd, tradeName)
	fs := cmd.Flags()
	out, closeOut := mustOpenOutput(fs)
	defer closeOut()
	err := watchDeposit(
		tr,
		wd,
		out,
		mustOutputTemplate(fs, depositChunkLogTemplates, nil),
		mustOutputTemplate(fs, blockInspectionTemplates, nil),
		mustNewclient(
			selectCryptoInfo(tr).Crypto,
			mustFlagRPCAddress(fs),
			mustFlagRPCUsername(fs),
			mustFlagRPCPassword(fs),
			mustFlagRPCTLSConfig(fs),
		),
		flagFirstBlock(fs),
		mustFlagIgnoreTarget(fs),
		mustFlagConfirmations(fs),
		watchStage,
		selectCryptoInfo,
		selectWatchData,
		selectFunds,
		selectInterruptStage,
		func(t trade.Trade) { mustSaveTrade(cmd, tradeName, t) },
		func(nwd *watchData) { saveWatchData(watchDataPath(cmd, tradeName), nwd) },
	)
	if err != nil {
		errorExit(ecFailedToWatch, err)
	}
	mustSaveTrade(cmd, tradeName, tr)
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

func newSecretTokenWatcher(cl cryptocore.Client, firstBlock uint64, wd *watchData, out io.Writer, blockTpl *template.Template, foundTpl *template.Template) func(trade.Trade) error {
	return func(tr trade.Trade) error {
		sig := make(chan os.Signal, 0)
		signal.Notify(sig, os.Interrupt, os.Kill)
		bdc, errc, closeIter := iterateBlocks(cl, wd.Own, firstBlock)
		defer closeIter()
		for {
			select {
			case <-sig:
				return trade.ErrInterruptTrade
			case err := <-errc:
				return err
			case db := <-bdc:
				if err := blockTpl.Execute(out, newBlockInfo(db.height, len(db.txs))); err != nil {
					return err
				}
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
					return foundTpl.Execute(out, token)
				}
			}
		}
	}
}

func watchSecretToken(tr trade.Trade, wd *watchData, cl cryptocore.Client, firstBlock uint64, out io.Writer, blockTpl *template.Template, foundTpl *template.Template) error {
	th := trade.NewHandler(nil)
	th.InstallStageHandlers(trade.StageHandlerMap{
		stages.WaitFundsRedeemed: newSecretTokenWatcher(cl, firstBlock, wd, out, blockTpl, foundTpl),
		stages.RedeemFunds:       trade.InterruptHandler,
	})
	for _, i := range th.Unhandled(tr.Stager().Stages()...) {
		th.InstallStageHandler(i, trade.NoOpHandler)
	}
	return th.HandleTrade(tr)
}

func cmdWatchSecretToken(cmd *cobra.Command, args []string) {
	tr := mustOpenTrade(cmd, args[0])
	fs := cmd.Flags()
	cl := mustNewclient(
		tr.OwnInfo().Crypto,
		mustFlagRPCAddress(fs),
		mustFlagRPCUsername(fs),
		mustFlagRPCPassword(fs),
		mustFlagRPCTLSConfig(fs),
	)
	out, outClose := mustOpenOutput(fs)
	defer outClose()
	blockTpl := mustOutputTemplate(fs, blockInspectionTemplates, nil)
	foundTpl, err := template.New("main").Parse("found token: {{ .Hex }}\n")
	if err != nil {
		errorExit(ecBadTemplate, err)
	}
	wd := mustOpenWatchData(cmd, args[0])
	if err := watchSecretToken(tr, wd, cl, flagFirstBlock(fs), out, blockTpl, foundTpl); err != nil {
		errorExit(ecFailedToWatch, err)
	}
	mustSaveTrade(cmd, args[0], tr)
}
