package {{ .All.Package | default "main" }}

import (
	"transmutate.io/pkg/atomicswap/types/key"
	"transmutate.io/pkg/atomicswap/types/transaction"
)

func newCrypto{{ .Self.short }}() Crypto {
	return &crypto{
		name:       "{{ .Self.name }}",
		short:      "{{ .Self.short }}",
		newPrivKey: key.NewPrivate{{ .Self.short | upper }},
		newTx:      transaction.New{{ .Self.short | upper }},
	}
}
