package key

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type testKey struct {
	parsePriv ParsePrivateFunc
	newPriv   NewPrivateFunc
	newPub    NewPublicFunc
}

var testKeys = map[string]*testKey{
	"bitcoin": &testKey{
		parsePriv: ParsePrivateBTC,
		newPriv:   NewPrivateBTC,
		newPub:    NewPublicBTC,
	},
	"litecoin": &testKey{
		parsePriv: ParsePrivateLTC,
		newPriv:   NewPrivateLTC,
		newPub:    NewPublicLTC,
	},
	"dogecoin": &testKey{
		parsePriv: ParsePrivateDOGE,
		newPriv:   NewPrivateDOGE,
		newPub:    NewPublicDOGE,
	},
	"bitcoin-cash": &testKey{
		parsePriv: ParsePrivateBCH,
		newPriv:   NewPrivateBCH,
		newPub:    NewPublicBCH,
	},
}

func TestKeys(t *testing.T) {
	for name, n := range testKeys {
		t.Run(name, func(t *testing.T) {
			// generate private key
			k1, err := n.newPriv()
			require.NoError(t, err, "can't create private key")
			// marshal
			b, err := yaml.Marshal(k1)
			require.NoError(t, err, "can't private marshal")
			// generate private key
			k2, err := n.newPriv()
			require.NoError(t, err, "can't create new key")
			// unmarshal
			err = yaml.Unmarshal(b, k2)
			require.NoError(t, err, "can't unmarshal private key")
			// sign
			msg := []byte("hello world")
			sig, err := k1.Sign(msg)
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
