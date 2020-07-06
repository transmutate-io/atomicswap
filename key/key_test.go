package key

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/transmutate-io/atomicswap/cryptos"
	"gopkg.in/yaml.v2"
)

func TestKeys(t *testing.T) {
	for _, name := range testCryptos {
		t.Run(name, func(t *testing.T) {
			crypto, err := cryptos.Parse(name)
			require.NoError(t, err, "can't parse crypto")
			// generate private key
			k1, err := NewPrivate(crypto)
			require.NoError(t, err, "can't create private key")
			// marshal
			b, err := yaml.Marshal(k1)
			require.NoError(t, err, "can't private marshal")
			// generate private key
			k2, err := NewPrivate(crypto)
			require.NoError(t, err, "can't create new key")
			// unmarshal
			err = yaml.Unmarshal(b, k2)
			require.NoError(t, err, "can't unmarshal private key")
			// sign
			msg := []byte("hello world")
			sig, err := k1.Sign(msg)
			require.NoError(t, err, "can't sign")
			// verify
			err = k2.Public().Verify(sig, msg)
			require.NoError(t, err, "can't verify signature")
			// marshal public key
			b, err = yaml.Marshal(k1.Public())
			require.NoError(t, err, "can't marshal public key")
			// unmarshal public key
			err = yaml.Unmarshal(b, k2.Public())
			require.NoError(t, err, "can't unmarshal public key")
		})
	}
}
