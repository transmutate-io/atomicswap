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
		Short:   "list trade names with lockset",
		Aliases: []string{"ls", "l"},
		Run:     cmdListLockSets,
	}
	exportLockSetCmd = &cobra.Command{
		Use:     "export <name>",
		Short:   "export lock set",
		Aliases: []string{"exp", "e"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdExportLockSet,
	}
	acceptLockSetCmd = &cobra.Command{
		Use:     "accept <name>",
		Short:   "accept a lock set for a trade",
		Aliases: []string{"a"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdAcceptLockSet,
	}
	showLockSetInfoCmd = &cobra.Command{
		Use:     "info <name>",
		Short:   "show a lock set info against a trade",
		Aliases: []string{"i"},
		Args:    cobra.ExactArgs(1),
		Run:     cmdShowLockSetInfo,
	}
)

func init() {
	fs := listLockSetsCmd.Flags()
	addFlagVerbose(fs)
	addFlagFormat(fs)
	fs = showLockSetInfoCmd.Flags()
	addFlagVerbose(fs)
	addFlagFormat(fs)
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
	err := eachLockSet(tradesDir(dataDir(cmd)), func(name string, tr trade.Trade) error {
		return tpl.Execute(out, &tradeInfo{Name: name, Trade: tr})
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
	th.InstallStageHandler(stages.ReceiveProposalResponse, func(t trade.Trade) error {
		return btr.SetLocks(openLockSet(in, tr.OwnInfo().Crypto, tr.TraderInfo().Crypto))
	})
	th.InstallStageHandler(stages.LockFunds, func(t trade.Trade) error {
		return trade.ErrInterruptTrade
	})
	for _, i := range th.Unhandled(tr.Stager().Stages()...) {
		th.InstallStageHandler(i, trade.NoOpHandler)
	}
	if err = th.HandleTrade(tr); err != nil && err != trade.ErrInterruptTrade {
		errorExit(ecCantAcceptLockSet, err)
	}
	if err = saveTrade(cmd, args[0], tr); err != nil {
		errorExit(ecCantCreateTrade, err)
	}
}

func cmdShowLockSetInfo(cmd *cobra.Command, args []string) {
	tr := openTrade(cmd, args[0])
	if _, err := tr.Buyer(); err != nil {
		errorExit(ecCantOpenTrade, err)
	}
	in, inClose := openInput(cmd)
	defer inClose()
	ls := openLockSet(in, tr.OwnInfo().Crypto, tr.TraderInfo().Crypto)
	info := &lockSetInfo{
		Trade:  tr,
		Buyer:  newLockInfo(cmd, ls.Buyer, tr.OwnInfo().Crypto),
		Seller: newLockInfo(cmd, ls.Seller, tr.TraderInfo().Crypto),
	}
	out, outClose := openOutput(cmd)
	defer outClose()
	tpl := outputTemplate(cmd, lockSetInfoTemplates, template.FuncMap{"now": time.Now})
	if err := tpl.Execute(out, info); err != nil {
		errorExit(ecBadTemplate, err)
	}
}

var lockSetInfoTemplates = []string{
	`hash: {{ if ne .Buyer.LockData.TokenHash.Hex .Seller.LockData.TokenHash.Hex }}mis{{ end }}match
buyer:
  recovery key data: {{ if ne .Buyer.LockData.RecoveryKeyData.Hex .Trade.RecoveryKey.Public.KeyData.Hex }}mis{{ end }}match
  time lock expiry: {{ .Buyer.LockData.Locktime.UTC }} (in {{ .Buyer.LockData.Locktime.UTC.Sub now.UTC  }})
  time lock expiry: {{ .Buyer.LockData.Locktime.UTC }} (in {{ .Buyer.LockData.Locktime.UTC.Sub now.UTC  }})
seller:
  redeem key data: {{ if ne .Seller.LockData.RedeemKeyData.Hex .Trade.RedeemKey.Public.KeyData.Hex }}mis{{ end }}match
  time lock expiry: {{ .Seller.LockData.Locktime.UTC }} (in {{ .Seller.LockData.Locktime.UTC.Sub now.UTC  }}, {{ .Buyer.LockData.Locktime.UTC.Sub .Seller.LockData.Locktime.UTC }} before buyer)
`,
	`hash: {{ if ne .Buyer.LockData.TokenHash.Hex .Seller.LockData.TokenHash.Hex }}mis{{ end }}match
buyer:
  deposit address: {{ .Buyer.DepositAddress}}
  redeem key data: {{ .Buyer.LockData.RedeemKeyData.Hex }}
  recovery key data: {{ if ne .Buyer.LockData.RecoveryKeyData.Hex .Trade.RecoveryKey.Public.KeyData.Hex }}mis{{ end }}match
  time lock expiry: {{ .Buyer.LockData.Locktime.UTC }} (in {{ .Buyer.LockData.Locktime.UTC.Sub now.UTC  }})
seller:
  deposit address: {{ .Seller.DepositAddress }}
  redeem key data: {{ if ne .Seller.LockData.RedeemKeyData.Hex .Trade.RedeemKey.Public.KeyData.Hex }}mis{{ end }}match
  recovery key data: {{ .Seller.LockData.RecoveryKeyData.Hex }}
  time lock expiry: {{ .Seller.LockData.Locktime.UTC }} (in {{ .Seller.LockData.Locktime.UTC.Sub now.UTC  }}, {{ .Buyer.LockData.Locktime.UTC.Sub .Seller.LockData.Locktime.UTC }} before buyer)
`,
	`hash: {{ if ne .Buyer.LockData.TokenHash.Hex .Seller.LockData.TokenHash.Hex }}mis{{ end }}match
buyer:
  deposit address: {{ .Buyer.DepositAddress}} ({{ .Buyer.Chain }})
  redeem key data: {{ .Buyer.LockData.RedeemKeyData.Hex }}
  recovery key data: {{ if ne .Buyer.LockData.RecoveryKeyData.Hex .Trade.RecoveryKey.Public.KeyData.Hex }}mis{{ end }}match ({{ .Buyer.LockData.RecoveryKeyData.Hex }}, {{ .Trade.RecoveryKey.Public.KeyData.Hex }})
  time lock expiry: {{ .Buyer.LockData.Locktime.UTC }} (in {{ .Buyer.LockData.Locktime.UTC.Sub now.UTC  }})
seller:
  deposit address: {{ .Seller.DepositAddress }} ({{ .Seller.Chain }})
  redeem key data: {{ if ne .Seller.LockData.RedeemKeyData.Hex .Trade.RedeemKey.Public.KeyData.Hex }}mis{{ end }}match ({{ .Seller.LockData.RedeemKeyData.Hex }}, {{ .Trade.RedeemKey.Public.KeyData.Hex }})
  recovery key data: {{ .Seller.LockData.RecoveryKeyData.Hex }}
  time lock expiry: {{ .Seller.LockData.Locktime.UTC }} (in {{ .Seller.LockData.Locktime.UTC.Sub now.UTC  }}, {{ .Buyer.LockData.Locktime.UTC.Sub .Seller.LockData.Locktime.UTC }} before buyer)
`,
}

func newLockInfo(cmd *cobra.Command, l trade.Lock, c *cryptos.Crypto) *lockInfo {
	chain := cryptoChain(cmd, c)
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

type lockSetInfo struct {
	Trade  trade.Trade
	Buyer  *lockInfo
	Seller *lockInfo
}
