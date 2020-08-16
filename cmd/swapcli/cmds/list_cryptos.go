package cmds

import (
	"io"
	"sort"
	"strings"
	"text/template"

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

func listCryptos(out io.Writer, tpl *template.Template) error {
	names := make([]string, 0, len(cryptos.Cryptos))
	for _, i := range cryptos.Cryptos {
		names = append(names, strings.ToLower(i.Name))
	}
	sort.Strings(names)
	for _, i := range names {
		if err := tpl.Execute(out, cryptos.Cryptos[i]); err != nil {
			return err
		}
	}
	return nil
}

func cmdListCryptos(cmd *cobra.Command, args []string) {
	out, closeOut := openOutput(cmd.Flags())
	defer closeOut()
	tpl := outputTemplate(cmd.Flags(), cryptosListTemplates, nil)
	if err := listCryptos(out, tpl); err != nil {
		errorExit(ecBadTemplate, err)
	}
}
