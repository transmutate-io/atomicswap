package cmds

import (
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/trade"
	"github.com/transmutate-io/cryptocore/types"
)

type templateData = map[string]interface{}

func outputTemplate(format string, verbose int, tpls []string, funcs template.FuncMap) (*template.Template, error) {
	var tplStr string
	if tplStr = format; tplStr == "" {
		tplStr = tpls[verbose]
	}
	r := template.New("main")
	if funcs != nil {
		r = r.Funcs(funcs)
	}
	return r.Parse(tplStr)
}

func mustOutputTemplate(fs *pflag.FlagSet, tpls []string, funcs template.FuncMap) *template.Template {
	r, err := outputTemplate(mustFlagFormat(fs), mustVerboseLevel(fs, len(tpls)-1), tpls, funcs)
	if err != nil {
		errorExit(ecBadTemplate, err)
	}
	return r
}

var (
	cryptosListTemplates = []string{
		"{{ .Name }}\n",
		"{{ .Name }}, {{ .Short }}\n",
		"{{ .Name }}, {{ .Short }}, {{ .Decimals }} decimal places\n",
		"{{ .Name }}, {{ .Short }}, {{ .Decimals }} decimal places, {{ .Type }}\n",
	}

	watchableTradesTemplates = []string{
		`{{ .name }}: {{ $s := .trade.Stager.Stage.String }}{{ if eq $s "lock-funds" -}}
	wait for own funds deposit
	{{- else if or (eq $s "send-proposal-response") (eq $s "wait-locked-funds") -}}
	wait for trader funds deposit
	{{- else -}}
	wait for secret token (trader redeem)
	{{- end }}
`,
	}

	blockInspectionTemplates = []string{
		"",
		"inspecting block at height {{ .height }}\n",
		"inspecting block at height {{ .height }} ({{ .txCount }} transactions)\n",
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

	tradeListTemplates = []string{
		"{{ .name }}\n",
		"{{ .name }} - {{ .trade.OwnInfo.Amount }} {{ .trade.OwnInfo.Crypto.Short }} for {{ .trade.TraderInfo.Amount }} {{ .trade.TraderInfo.Crypto.Short }}\n",
		"{{ .name }} - {{ .trade.OwnInfo.Amount }} {{ .trade.OwnInfo.Crypto.Short }} (locked for {{ .trade.Duration.String }}) for {{ .trade.TraderInfo.Amount }} {{ .trade.TraderInfo.Crypto.Short }}\n",
		"{{ .name }} - {{ .trade.OwnInfo.Amount }} {{ .trade.OwnInfo.Crypto.Short }} (locked for {{ .trade.Duration.String }}) for {{ .trade.TraderInfo.Amount }} {{ .trade.TraderInfo.Crypto.Short }} - {{ .trade.Stager.Stage }}\n",
	}

	lockSetInfoTemplates = []string{
		`hash: {{ if ne .buyer.lockData.TokenHash.Hex .seller.lockData.TokenHash.Hex }}mis{{ end }}match
buyer:
  recovery key data: {{ if ne .buyer.lockData.RecoveryKeyData.Hex .trade.RecoveryKey.Public.KeyData.Hex }}mis{{ end }}match
  time lock expiry: {{ .buyer.lockData.Locktime.UTC }} (in {{ .buyer.lockData.Locktime.UTC.Sub now.UTC  }})
  time lock expiry: {{ .buyer.lockData.Locktime.UTC }} (in {{ .buyer.lockData.Locktime.UTC.Sub now.UTC  }})
seller:
  redeem key data: {{ if ne .seller.lockData.RedeemKeyData.Hex .trade.RedeemKey.Public.KeyData.Hex }}mis{{ end }}match
  time lock expiry: {{ .seller.lockData.Locktime.UTC }} (in {{ .seller.lockData.Locktime.UTC.Sub now.UTC  }}, {{ .buyer.lockData.Locktime.UTC.Sub .seller.lockData.Locktime.UTC }} before buyer)
`,
		`hash: {{ if ne .buyer.lockData.TokenHash.Hex .seller.lockData.TokenHash.Hex }}mis{{ end }}match
buyer:
  deposit address: {{ .buyer.depositAddr }}
  redeem key data: {{ .buyer.lockData.RedeemKeyData.Hex }}
  recovery key data: {{ if ne .buyer.lockData.RecoveryKeyData.Hex .trade.RecoveryKey.Public.KeyData.Hex }}mis{{ end }}match
  time lock expiry: {{ .buyer.lockData.Locktime.UTC }} (in {{ .buyer.lockData.Locktime.UTC.Sub now.UTC  }})
seller:
  deposit address: {{ .seller.depositAddr }}
  redeem key data: {{ if ne .seller.lockData.RedeemKeyData.Hex .trade.RedeemKey.Public.KeyData.Hex }}mis{{ end }}match
  recovery key data: {{ .seller.lockData.RecoveryKeyData.Hex }}
  time lock expiry: {{ .seller.lockData.Locktime.UTC }} (in {{ .seller.lockData.Locktime.UTC.Sub now.UTC  }}, {{ .buyer.lockData.Locktime.UTC.Sub .seller.lockData.Locktime.UTC }} before buyer)
`,
		`hash: {{ if ne .buyer.lockData.TokenHash.Hex .seller.lockData.TokenHash.Hex }}mis{{ end }}match
buyer:
  deposit address: {{ .buyer.depositAddr}} ({{ .buyer.chain }})
  redeem key data: {{ .buyer.lockData.RedeemKeyData.Hex }}
  recovery key data: {{ if ne .buyer.lockData.RecoveryKeyData.Hex .trade.RecoveryKey.Public.KeyData.Hex }}mis{{ end }}match ({{ .buyer.lockData.RecoveryKeyData.Hex }}, {{ .trade.RecoveryKey.Public.KeyData.Hex }})
  time lock expiry: {{ .buyer.lockData.Locktime.UTC }} (in {{ .buyer.lockData.Locktime.UTC.Sub now.UTC  }})
seller:
  deposit address: {{ .seller.depositAddr }} ({{ .seller.chain }})
  redeem key data: {{ if ne .seller.lockData.RedeemKeyData.Hex .trade.RedeemKey.Public.KeyData.Hex }}mis{{ end }}match ({{ .seller.lockData.RedeemKeyData.Hex }}, {{ .trade.RedeemKey.Public.KeyData.Hex }})
  recovery key data: {{ .seller.lockData.RecoveryKeyData.Hex }}
  time lock expiry: {{ .seller.lockData.Locktime.UTC }} (in {{ .seller.lockData.Locktime.UTC.Sub now.UTC  }}, {{ .buyer.lockData.Locktime.UTC.Sub .seller.lockData.Locktime.UTC }} before buyer)
`,
	}
)

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

func newBlockInfo(height uint64, txCount int) templateData {
	return templateData{"height": height, "txCount": txCount}
}

func newTradeInfo(name string, trade trade.Trade) templateData {
	return templateData{"name": name, "trade": trade}
}

func mustNewLockInfo(cmd *cobra.Command, l trade.Lock, c *cryptos.Crypto) templateData {
	chain := mustFlagCryptoChain(c)
	addr, err := l.Address(chain)
	if err != nil {
		errorExit(ecCantCalculateAddress, err)
	}
	ld, err := l.LockData()
	if err != nil {
		errorExit(ecInvalidLockData, err)
	}
	return templateData{"depositAddr": addr, "chain": chain, "lockData": ld}
}

func newLockSetInfo(trade trade.Trade, buyer, seller templateData) templateData {
	return templateData{"trade": trade, "buyer": buyer, "seller": seller}
}
