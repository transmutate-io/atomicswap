package cmds

import (
	"sort"
	"strings"
	"text/template"

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
	addVerboseFlag(fs)
	addFormatFlag(fs)
}

func cmdListCryptos(cmd *cobra.Command, args []string) {
	out, closeOut := openOutput(cmd)
	defer closeOut()
	vl := verboseLevel(cmd.Flags(), len(cryptosListTemplates)-1)
	tpl, err := template.New("main").Parse(cryptosListTemplates[vl])
	if err != nil {
		errorExit(ECBadTemplate, "bad template: %#v\n", err)
	}
	names := make([]string, 0, len(cryptos.Cryptos))
	for _, i := range cryptos.Cryptos {
		names = append(names, strings.ToLower(i.Name))
	}
	sort.Strings(names)
	for _, i := range names {
		if err = tpl.Execute(out, cryptos.Cryptos[i]); err != nil {
			errorExit(ECBadTemplate, "bad template: %#v\n", err)
		}
	}
}

var cryptosListTemplates = []string{
	"{{ .Name }}\n",
	"{{ .Name }}, {{ .Short }}\n",
	"{{ .Name }}, {{ .Short }}, {{ .Decimals }} decimal places\n",
	"{{ .Name }}, {{ .Short }}, {{ .Decimals }} decimal places, {{ .Type }}\n",
}
