package key

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestPrivateKey(t *testing.T) {
	k1, err := NewPrivate()
	require.NoError(t, err, "unexpected error")
	b, err := yaml.Marshal(k1)
	require.NoError(t, err, "unexpected error")
	k2 := &Private{}
	err = yaml.Unmarshal(b, k2)
	require.NoError(t, err, "unexpected error")
	msg := []byte("hello world")
	sig, err := k1.Sign(msg)
	require.NoError(t, err, "unexpected error")
	require.True(t, sig.Verify(msg, k2.PubKey()), "keys mismatch")
}

func TestPublicKey(t *testing.T) {
	k, err := NewPrivate()
	require.NoError(t, err, "unexpected error")
	msg := []byte("hello world")
	sig, err := k.Sign(msg)
	require.NoError(t, err, "unexpected error")
	pubBytes, err := yaml.Marshal(&Public{PublicKey: k.PubKey()})
	pub := &Public{}
	err = yaml.Unmarshal(pubBytes, pub)
	require.NoError(t, err, "unexpected error")
	require.True(t, sig.Verify(msg, pub.PublicKey), "keys mismatch")
}
