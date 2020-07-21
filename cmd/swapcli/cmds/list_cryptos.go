package cmds

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/cryptos"
)

var (
	ListCryptosCmd = &cobra.Command{
		Use:     "cryptos",
		Short:   "list cryptos",
		Aliases: []string{"c"},
		Run:     cmdListCryptos,
	}
)

func init() {
	fs := ListCryptosCmd.Flags()
	addFlagVerbose(fs)
	addFlagFormat(fs)
}

func cmdListCryptos(cmd *cobra.Command, args []string) {
	tpl := outputTemplate(cmd, cryptosListTemplates, nil)
	names := make([]string, 0, len(cryptos.Cryptos))
	for _, i := range cryptos.Cryptos {
		names = append(names, strings.ToLower(i.Name))
	}
	sort.Strings(names)
	out, closeOut := openOutput(cmd)
	defer closeOut()
	for _, i := range names {
		if err := tpl.Execute(out, cryptos.Cryptos[i]); err != nil {
			errorExit(ecBadTemplate, err)
		}
	}
}

var cryptosListTemplates = []string{
	"{{ .Name }}\n",
	"{{ .Name }}, {{ .Short }}\n",
	"{{ .Name }}, {{ .Short }}, {{ .Decimals }} decimal places\n",
	"{{ .Name }}, {{ .Short }}, {{ .Decimals }} decimal places, {{ .Type }}\n",
}
