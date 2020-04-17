package {{ .All.Package | default "main" }}

import (
	"transmutate.io/pkg/atomicswap/cryptotypes"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/transaction"
)

func newCrypto{{ .Self.short }}() *Crypto {
	return &Crypto{
		Name:       "{{ .Self.name }}",
		Short:      "{{ .Self.short }}",
		newPrivKey: key.NewPrivate{{ .Self.short | upper }},
		newTx:      transaction.New{{ .Self.short | upper }},
		Type:       cryptotypes.{{ .Self.type }},
	}
}
