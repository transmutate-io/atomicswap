package {{ .All.Package | default "main" }}

import "transmutate.io/pkg/atomicswap/types/key"

func newCrypto{{ .Self.short }}() Crypto {
	return &crypto{
		name:       "{{ .Self.name }}",
		short:      "{{ .Self.short }}",
		newPrivKey: key.NewPrivate{{ .Self.short | upper }},
	}
}
