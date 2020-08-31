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

func sortedCryptos() []string {
	r := make([]string, 0, len(cryptos.Cryptos))
	for _, i := range cryptos.Cryptos {
		r = append(r, strings.ToLower(i.Name))
	}
	sort.Strings(r)
	return r
}

func listCryptos(out io.Writer, tpl *template.Template) error {
	for _, i := range sortedCryptos() {
		if err := tpl.Execute(out, cryptos.Cryptos[i]); err != nil {
			return err
		}
	}
	return nil
}

func cmdListCryptos(cmd *cobra.Command, args []string) {
	out, closeOut := mustOpenOutput(cmd.Flags())
	defer closeOut()
	tpl := mustOutputTemplate(cmd.Flags(), cryptosListTemplates, nil)
	if err := listCryptos(out, tpl); err != nil {
		errorExit(ecBadTemplate, err)
	}
}
