package {{ .All.Package | default "main" }}

import "transmutate.io/pkg/atomicswap/cryptotypes"

func newCrypto{{ .Self.short }}() *Crypto {
	return &Crypto{
		Name:     "{{ .Self.name }}",
		Short:    "{{ .Self.short }}",
		Decimals: {{ .Self.decimals }},
		Type:     cryptotypes.{{ .Self.type }},
	}
}
