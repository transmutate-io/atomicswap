package cmds

import (
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/params"
	"github.com/transmutate-io/atomicswap/stages"
	"github.com/transmutate-io/atomicswap/trade"
	"gopkg.in/yaml.v2"
)

var (
	LockSetCmd = &cobra.Command{
		Use:     "lockset <command>",
		Short:   "lockset commands",
		Aliases: []string{"lock", "l"},
	}
	listLockSetsCmd = &cobra.Command{
		Use:     "list",
		Short:   "list trade names with exportable locksets to output",
		Aliases: []string{"ls", "l"},
		Args:    cobra.NoArgs,
		Run:     cmdListLockSets,
	}
	exportLockSetCmd = &cobra.Command{
		Use:     "export <trade_name>",
		Short:   "export a lockset to output",
		Aliases: []string{"exp", "e"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdExportLockSet,
	}
	acceptLockSetCmd = &cobra.Command{
		Use:     "accept <trade_name>",
		Short:   "accept a lockset for a trade from input",
		Aliases: []string{"a"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdAcceptLockSet,
	}
	showLockSetInfoCmd = &cobra.Command{
		Use:     "info <trade_name>",
		Short:   "show a lockset info against a trade to output",
		Aliases: []string{"i"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdShowLockSetInfo,
	}
)

func init() {
	fs := listLockSetsCmd.Flags()
	addFlagVerbose(fs)
	addFlagFormat(fs)
	addFlagOutput(fs)
	fs = showLockSetInfoCmd.Flags()
	addFlagVerbose(fs)
	addFlagFormat(fs)
	addFlagOutput(fs)
	addFlagInput(showLockSetInfoCmd.Flags())
	addFlagInput(acceptLockSetCmd.Flags())
	addFlagOutput(exportLockSetCmd.Flags())
	addFlagCryptoChain(fs)
	for _, i := range []*cobra.Command{
		listLockSetsCmd,
		exportLockSetCmd,
		acceptLockSetCmd,
		showLockSetInfoCmd,
	} {
		LockSetCmd.AddCommand(i)
	}
}

func cmdListLockSets(cmd *cobra.Command, args []string) {
	tpl := outputTemplate(cmd, tradeListTemplates, nil)
	out, closeOut := openOutput(cmd)
	defer closeOut()
	err := eachLockSet(tradesDir(cmd), func(name string, tr trade.Trade) error {
		return tpl.Execute(out, newTradeInfo(name, tr))
	})
	if err != nil {
		errorExit(ecCantListLockSets, err)
	}
}

func cmdExportLockSet(cmd *cobra.Command, args []string) {
	tr := openTrade(cmd, args[0])
	str, err := tr.Seller()
	if err != nil {
		errorExit(ecCantExportLockSet, err)
	}
	ls := str.Locks()
	out, outClose := openOutput(cmd)
	defer outClose()
	if err := yaml.NewEncoder(out).Encode(ls); err != nil {
		errorExit(ecCantExportLockSet, err)
	}
}

func cmdAcceptLockSet(cmd *cobra.Command, args []string) {
	tr := openTrade(cmd, args[0])
	in, inClose := openInput(cmd)
	defer inClose()
	btr, err := tr.Buyer()
	if err != nil {
		errorExit(ecCantAcceptLockSet, err)
	}
	th := trade.NewHandler(trade.DefaultStageHandlers)
	th.InstallStageHandlers(trade.StageHandlerMap{
		stages.ReceiveProposalResponse: func(t trade.Trade) error {
			return btr.SetLocks(openLockSet(in, tr.OwnInfo().Crypto, tr.TraderInfo().Crypto))
		},
		stages.LockFunds: trade.InterruptHandler,
	})
	for _, i := range th.Unhandled(tr.Stager().Stages()...) {
		th.InstallStageHandler(i, trade.NoOpHandler)
	}
	if err = th.HandleTrade(tr); err != nil && err != trade.ErrInterruptTrade {
		errorExit(ecCantAcceptLockSet, err)
	}
	saveTrade(cmd, args[0], tr)
}

func cmdShowLockSetInfo(cmd *cobra.Command, args []string) {
	tr := openTrade(cmd, args[0])
	if _, err := tr.Buyer(); err != nil {
		errorExit(ecCantOpenTrade, err)
	}
	in, inClose := openInput(cmd)
	defer inClose()
	ls := openLockSet(in, tr.OwnInfo().Crypto, tr.TraderInfo().Crypto)
	out, outClose := openOutput(cmd)
	defer outClose()
	tpl := outputTemplate(cmd, lockSetInfoTemplates, template.FuncMap{"now": time.Now})
	err := tpl.Execute(out, newLockSetInfo(
		tr,
		newLockInfo(cmd, ls.Buyer, tr.OwnInfo().Crypto),
		newLockInfo(cmd, ls.Seller, tr.TraderInfo().Crypto),
	))
	if err != nil {
		errorExit(ecBadTemplate, err)
	}
}

var lockSetInfoTemplates = []string{
	`hash: {{ if ne .buyer.LockData.TokenHash.Hex .seller.LockData.TokenHash.Hex }}mis{{ end }}match
buyer:
  recovery key data: {{ if ne .buyer.LockData.RecoveryKeyData.Hex .trade.RecoveryKey.Public.KeyData.Hex }}mis{{ end }}match
  time lock expiry: {{ .buyer.LockData.Locktime.UTC }} (in {{ .buyer.LockData.Locktime.UTC.Sub now.UTC  }})
  time lock expiry: {{ .buyer.LockData.Locktime.UTC }} (in {{ .buyer.LockData.Locktime.UTC.Sub now.UTC  }})
seller:
  redeem key data: {{ if ne .seller.LockData.RedeemKeyData.Hex .trade.RedeemKey.Public.KeyData.Hex }}mis{{ end }}match
  time lock expiry: {{ .seller.LockData.Locktime.UTC }} (in {{ .seller.LockData.Locktime.UTC.Sub now.UTC  }}, {{ .buyer.LockData.Locktime.UTC.Sub .seller.LockData.Locktime.UTC }} before buyer)
`,
	`hash: {{ if ne .buyer.LockData.TokenHash.Hex .seller.LockData.TokenHash.Hex }}mis{{ end }}match
buyer:
  deposit address: {{ .buyer.DepositAddress}}
  redeem key data: {{ .buyer.LockData.RedeemKeyData.Hex }}
  recovery key data: {{ if ne .buyer.LockData.RecoveryKeyData.Hex .trade.RecoveryKey.Public.KeyData.Hex }}mis{{ end }}match
  time lock expiry: {{ .buyer.LockData.Locktime.UTC }} (in {{ .buyer.LockData.Locktime.UTC.Sub now.UTC  }})
seller:
  deposit address: {{ .seller.DepositAddress }}
  redeem key data: {{ if ne .seller.LockData.RedeemKeyData.Hex .trade.RedeemKey.Public.KeyData.Hex }}mis{{ end }}match
  recovery key data: {{ .seller.LockData.RecoveryKeyData.Hex }}
  time lock expiry: {{ .seller.LockData.Locktime.UTC }} (in {{ .seller.LockData.Locktime.UTC.Sub now.UTC  }}, {{ .buyer.LockData.Locktime.UTC.Sub .seller.LockData.Locktime.UTC }} before buyer)
`,
	`hash: {{ if ne .buyer.LockData.TokenHash.Hex .seller.LockData.TokenHash.Hex }}mis{{ end }}match
buyer:
  deposit address: {{ .buyer.DepositAddress}} ({{ .buyer.Chain }})
  redeem key data: {{ .buyer.LockData.RedeemKeyData.Hex }}
  recovery key data: {{ if ne .buyer.LockData.RecoveryKeyData.Hex .trade.RecoveryKey.Public.KeyData.Hex }}mis{{ end }}match ({{ .buyer.LockData.RecoveryKeyData.Hex }}, {{ .trade.RecoveryKey.Public.KeyData.Hex }})
  time lock expiry: {{ .buyer.LockData.Locktime.UTC }} (in {{ .buyer.LockData.Locktime.UTC.Sub now.UTC  }})
seller:
  deposit address: {{ .seller.DepositAddress }} ({{ .seller.Chain }})
  redeem key data: {{ if ne .seller.LockData.RedeemKeyData.Hex .trade.RedeemKey.Public.KeyData.Hex }}mis{{ end }}match ({{ .seller.LockData.RedeemKeyData.Hex }}, {{ .trade.RedeemKey.Public.KeyData.Hex }})
  recovery key data: {{ .seller.LockData.RecoveryKeyData.Hex }}
  time lock expiry: {{ .seller.LockData.Locktime.UTC }} (in {{ .seller.LockData.Locktime.UTC.Sub now.UTC  }}, {{ .buyer.LockData.Locktime.UTC.Sub .seller.LockData.Locktime.UTC }} before buyer)
`,
}

func newLockInfo(cmd *cobra.Command, l trade.Lock, c *cryptos.Crypto) *lockInfo {
	chain := flagCryptoChain(cmd, c)
	addr, err := l.Address(chain)
	if err != nil {
		errorExit(ecCantCalculateAddress, err)
	}
	ld, err := l.LockData()
	if err != nil {
		errorExit(ecInvalidLockData, err)
	}
	return &lockInfo{
		DepositAddress: addr,
		Chain:          chain,
		LockData:       ld,
	}
}

type lockInfo struct {
	DepositAddress string
	Chain          params.Chain
	LockData       *trade.LockData
}

func newLockSetInfo(trade trade.Trade, buyer *lockInfo, seller *lockInfo) templateData {
	return templateData{
		"trade":  trade,
		"buyer":  buyer,
		"seller": seller,
	}
}
