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
		Short:   "list available cryptocurrencies",
		Aliases: []string{"c"},
		Args:    cobra.NoArgs,
		Run:     cmdListCryptos,
	}
)

func init() {
	addFlags(flagMap{
		ListCryptosCmd.Flags(): []flagFunc{
			addFlagVerbose,
			addFlagFormat,
			addFlagOutput,
		},
	})
}

func cmdListCryptos(cmd *cobra.Command, args []string) {
	tpl := outputTemplate(cmd.Flags(), cryptosListTemplates, nil)
	names := make([]string, 0, len(cryptos.Cryptos))
	for _, i := range cryptos.Cryptos {
		names = append(names, strings.ToLower(i.Name))
	}
	sort.Strings(names)
	out, closeOut := openOutput(cmd.Flags())
	defer closeOut()
	for _, i := range names {
		if err := tpl.Execute(out, cryptos.Cryptos[i]); err != nil {
			errorExit(ecBadTemplate, err)
		}
	}
}
