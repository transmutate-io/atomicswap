package key

import (
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/cryptocore/types"
)

type (
	newPrivateFunc   = func() (Private, error)
	parsePublicFunc  = func(b []byte) (Public, error)
	parsePrivateFunc = func(b []byte) (Private, error)

	// KeyData represents the public key data
	KeyData = types.Bytes

	// Private is a crypto private key
	Private interface {
		// Public returns the public key
		Public() Public
		// Sign signs a message
		Sign(b []byte) ([]byte, error)
		// Serialize returns the serialized key
		Serialize() []byte
		// Key returns the underlying key
		Key() interface{}
	}

	// Public is a crypto private key
	Public interface {
		// KeyData returns the key data
		KeyData() KeyData
		// Verify verifies a signed message
		Verify(sig, msg []byte) error
		// SerializeCompressed returns the serialized and compressed key
		SerializeCompressed() []byte
		// Key returns the underlying key
		Key() interface{}
	}
)

type newFuncs struct {
	parsePriv parsePrivateFunc
	parsePub  parsePublicFunc
	newPriv   newPrivateFunc
}

func getCryptoFuncs(c *cryptos.Crypto) (*newFuncs, error) {
	cf, ok := cryptoFuncs[c.String()]
	if !ok {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return &cf, nil
}

// ParsePrivate parses a private key for the given crypto
func ParsePrivate(c *cryptos.Crypto, b []byte) (Private, error) {
	cf, err := getCryptoFuncs(c)
	if err != nil {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return cf.parsePriv(b)
}

// NewPrivate returns a new private key for the given crypto
func NewPrivate(c *cryptos.Crypto) (Private, error) {
	cf, err := getCryptoFuncs(c)
	if err != nil {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return cf.newPriv()
}

// ParsePublic parses a public key for a given crypto
func ParsePublic(c *cryptos.Crypto, b []byte) (Public, error) {
	cf, err := getCryptoFuncs(c)
	if err != nil {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return cf.parsePub(b)
}
