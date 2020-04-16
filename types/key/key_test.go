package key

import (
	"testing"

	"transmutate.io/pkg/atomicswap/params"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type testKey struct {
	newPriv NewPrivateFunc
	newPub  NewPublicFunc
}

var testKeys = map[string]*testKey{
	params.Bitcoin.String(): &testKey{
		newPriv: NewPrivateBTC,
		newPub:  NewPublicBTC,
	},
	params.Litecoin.String(): &testKey{
		newPriv: NewPrivateLTC,
		newPub:  NewPublicLTC,
	},
	params.Dogecoin.String(): &testKey{
		newPriv: NewPrivateBTC,
		newPub:  NewPublicBTC,
	},
	params.Dogecoin.String(): &testKey{
		newPriv: NewPrivateDOGE,
		newPub:  NewPublicDOGE,
	},
	params.BitcoinCash.String(): &testKey{
		newPriv: NewPrivateBTCCash,
		newPub:  NewPublicBTCCash,
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
