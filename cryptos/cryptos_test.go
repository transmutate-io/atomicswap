package cryptos_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"transmutate.io/pkg/atomicswap/cryptos"
)

func TestParseCrypto(t *testing.T) {
	for i := range cryptos.Cryptos {
		t.Run(i, func(t *testing.T) {
			c, err := cryptos.Parse(i)
			require.NoError(t, err, "parsing error")
			_ = c
		})
	}
}
