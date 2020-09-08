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
	"github.com/transmutate-io/atomicswap/internal/cmdutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil/exitcodes"
	"github.com/transmutate-io/atomicswap/internal/tplutil"
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
	network := &_network
	flagutil.AddFlags(flagutil.FlagFuncMap{
		listWatchableCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddFormat,
			flagutil.AddVerbose,
			flagutil.AddOutput,
		},
		watchOwnDepositCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddRPC,
			flagutil.AddFirstBlock,
			flagutil.AddFormat,
			flagutil.AddVerbose,
			flagutil.AddOutput,
			network.AddFlag,
			flagutil.AddIgnoreTarget,
			flagutil.AddConfirmations,
		},
		watchTraderDepositCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddRPC,
			flagutil.AddFirstBlock,
			flagutil.AddFormat,
			flagutil.AddVerbose,
			flagutil.AddOutput,
			network.AddFlag,
			flagutil.AddIgnoreTarget,
			flagutil.AddConfirmations,
		},
		watchSecretTokenCmd.Flags(): []flagutil.FlagFunc{
			flagutil.AddRPC,
			flagutil.AddFirstBlock,
			flagutil.AddFormat,
			flagutil.AddVerbose,
			flagutil.AddOutput,
			network.AddFlag,
			flagutil.AddIgnoreTarget,
		},
	})
	cmdutil.AddCommands(WatchCmd, []*cobra.Command{
		listWatchableCmd,
		watchOwnDepositCmd,
		watchTraderDepositCmd,
		watchSecretTokenCmd,
	})
}

func listWatchable(td string, out io.Writer, tpl *template.Template) error {
	return eachTrade(td, func(name string, tr trade.Trade) error {
		return tpl.Execute(out, newTradeInfo(name, tr))
	})
}

func cmdListWatchable(cmd *cobra.Command, args []string) {
	out, closeOut := flagutil.MustOpenOutput(cmd.Flags())
	defer closeOut()
	tpl := tplutil.MustOpenTemplate(cmd.Flags(), watchableTradesTemplates, nil)
	if err := listWatchable(tradesDir(cmd), out, tpl); err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
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
	cryptoInfo *trade.TraderInfo,
	bwd *blockWatchData,
	funds trade.FundsData,
	tradeSave func(trade.Trade),
	wdSave func(*watchData),
) error {
	sig := make(chan os.Signal, 0)
	signal.Notify(sig, os.Interrupt, os.Kill)
	targetAmount := cryptoInfo.Amount.UInt64(cryptoInfo.Crypto.Decimals)
	depositAddr, err := funds.Lock().Address(_network.MustNetwork(cryptoInfo.Crypto.Name))
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
		err := depositTpl.Execute(out, newOutputInfo(
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
			return nil
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
					err = depositTpl.Execute(out, newOutputInfo(
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

func cmdWatchDeposit(
	cmd *cobra.Command,
	tradeName string,
	selectCryptoInfo func(trade.Trade) *trade.TraderInfo,
	selectWatchData func(*watchData) *blockWatchData,
	selectFunds func(trade.Trade) trade.FundsData,
) {
	tr := mustOpenTrade(cmd, tradeName)
	wd := mustOpenWatchData(cmd, tradeName)
	fs := cmd.Flags()
	out, closeOut := flagutil.MustOpenOutput(fs)
	defer closeOut()
	cryptoInfo := selectCryptoInfo(tr)
	err := watchDeposit(
		tr,
		wd,
		out,
		tplutil.MustOpenTemplate(fs, depositChunkLogTemplates, nil),
		tplutil.MustOpenTemplate(fs, blockInspectionTemplates, nil),
		mustNewClient(
			cryptoInfo.Crypto,
			flagutil.MustRPCAddress(fs),
			flagutil.MustRPCUsername(fs),
			flagutil.MustRPCPassword(fs),
			flagutil.MustRPCTLSConfig(fs),
		),
		flagutil.MustFirstBlock(fs),
		flagutil.MustIgnoreTarget(fs),
		flagutil.MustConfirmations(fs),
		cryptoInfo,
		selectWatchData(wd),
		selectFunds(tr),
		func(t trade.Trade) { mustSaveTrade(cmd, tradeName, t) },
		func(nwd *watchData) { saveWatchData(watchDataPath(cmd, tradeName), nwd) },
	)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
	mustSaveTrade(cmd, tradeName, tr)
}

func cmdWatchOwnDeposit(cmd *cobra.Command, args []string) {
	cmdWatchDeposit(
		cmd,
		args[0],
		func(tr trade.Trade) *trade.TraderInfo { return tr.OwnInfo() },
		func(wd *watchData) *blockWatchData { return wd.Own },
		func(tr trade.Trade) trade.FundsData { return tr.RecoverableFunds() },
	)
}

func cmdWatchTraderDeposit(cmd *cobra.Command, args []string) {
	cmdWatchDeposit(
		cmd,
		args[0],
		func(tr trade.Trade) *trade.TraderInfo { return tr.TraderInfo() },
		func(wd *watchData) *blockWatchData { return wd.Trader },
		func(tr trade.Trade) trade.FundsData { return tr.RedeemableFunds() },
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
				return nil
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

func cmdWatchSecretToken(cmd *cobra.Command, args []string) {
	tr := mustOpenTrade(cmd, args[0])
	fs := cmd.Flags()
	cl := mustNewClient(
		tr.OwnInfo().Crypto,
		flagutil.MustRPCAddress(fs),
		flagutil.MustRPCUsername(fs),
		flagutil.MustRPCPassword(fs),
		flagutil.MustRPCTLSConfig(fs),
	)
	out, outClose := flagutil.MustOpenOutput(fs)
	defer outClose()
	blockTpl := tplutil.MustOpenTemplate(fs, blockInspectionTemplates, nil)
	foundTpl, err := template.New("main").Parse("found token: {{ .Hex }}\n")
	if err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
	wd := mustOpenWatchData(cmd, args[0])
	if err := watchSecretToken(tr, wd, cl, flagutil.MustFirstBlock(fs), out, blockTpl, foundTpl); err != nil {
		cmdutil.ErrorExit(exitcodes.ExecutionError, err)
	}
	mustSaveTrade(cmd, args[0], tr)
}
