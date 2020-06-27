package testutil

import (
	"crypto/rand"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/networks"
	"transmutate.io/pkg/atomicswap/script"
)

func MustNewPrivateKey(t *testing.T, c *cryptos.Crypto) key.Private {
	r, err := key.NewPrivate(c)
	require.NoError(t, err, "can't create private key")
	return r
}

func ReadRandom(n int) ([]byte, error) {
	r := make([]byte, n)
	rn, err := rand.Read(r)
	if err != nil {
		return nil, err
	} else if rn != n {
		return nil, errors.New("not enough random bytes")
	}
	return r, nil
}

func MustReadRandom(t *testing.T, n int) []byte {
	r, err := ReadRandom(n)
	require.NoError(t, err, "can't read random bytes")
	return r
}

func MustNewEngine(t *testing.T, c *cryptos.Crypto) *script.Engine {
	r, err := script.NewEngine(c)
	require.NoError(t, err, "can't create scripting engine")
	return r
}

func MustNewGenerator(t *testing.T, c *cryptos.Crypto) script.Generator {
	r, err := script.NewGenerator(c)
	require.NoError(t, err, "can't create script generator")
	return r
}

func MustP2SHAddress(t *testing.T, tc *Crypto, s []byte) string {
	r, err := networks.AllByName[tc.Name][tc.Chain].P2SHFromScript(s)
	require.NoError(t, err, "can't generate p2sh address")
	return r
}

func MustP2PKHAddress(t *testing.T, tc *Crypto, kd []byte) string {
	r, err := networks.AllByName[tc.Name][tc.Chain].P2PKH(kd)
	require.NoError(t, err, "can't generate p2sh address")
	return r
}
