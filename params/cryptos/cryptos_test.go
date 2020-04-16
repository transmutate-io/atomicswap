package cryptos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCrypto(t *testing.T) {
	for i := range _cryptos {
		t.Run(i, func(t *testing.T) {
			c, err := ParseCrypto(i)
			require.NoError(t, err, "parsing error")
			_ = c
		})
	}
}
