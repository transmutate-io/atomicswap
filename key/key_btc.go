package key

import (
	"encoding/base64"
	"errors"

	"github.com/btcsuite/btcd/btcec"
	"github.com/transmutate-io/atomicswap/hash"
)

// PrivateBTC represents a private key for bitcoin
type PrivateBTC struct{ *btcec.PrivateKey }

func parsePrivateBTC(b []byte) *PrivateBTC {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), b)
	return &PrivateBTC{PrivateKey: priv}
}

// ParsePrivateBTC parses a bitcoin private key
func ParsePrivateBTC(b []byte) (Private, error) { return parsePrivateBTC(b), nil }

func newPrivateBTC() (*PrivateBTC, error) {
	k, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	return &PrivateBTC{PrivateKey: k}, nil
}

// NewPrivateBTC returns a new bitcoin private key
func NewPrivateBTC() (Private, error) { return newPrivateBTC() }

// Sign implement Private
func (k *PrivateBTC) Sign(b []byte) ([]byte, error) {
	sig, err := k.PrivateKey.Sign(b)
	if err != nil {
		return nil, err
	}
	return sig.Serialize(), nil
}

// MarshalYAML implement yaml.Marshaler
func (k *PrivateBTC) MarshalYAML() (interface{}, error) {
	if k == nil {
		return nil, nil
	}
	return base64.RawStdEncoding.EncodeToString(k.Serialize()), nil
}

// UnmarshalYAML implement yaml.Unmarshaler
func (k *PrivateBTC) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	b, err := base64.RawStdEncoding.DecodeString(r)
	if err != nil {
		return err
	}
	k.PrivateKey = parsePrivateBTC(b).PrivateKey
	return nil
}

// Public implement Private
func (k *PrivateBTC) Public() Public { return &PublicBTC{k.PubKey()} }

// Key implement Private
func (k *PrivateBTC) Key() interface{} { return k.PrivateKey }

// PublicBTC represents a bitcoin public key
type PublicBTC struct{ *btcec.PublicKey }

func parsePublicBTC(b []byte) (*PublicBTC, error) {
	pub, err := btcec.ParsePubKey(b, btcec.S256())
	if err != nil {
		return nil, err
	}
	return &PublicBTC{PublicKey: pub}, nil
}

// ParsePublicBTC parses a bitcoin public key
func ParsePublicBTC(b []byte) (Public, error) { return parsePublicBTC(b) }

// Verify implement Public
func (k *PublicBTC) Verify(sig, msg []byte) error {
	s, err := btcec.ParseSignature(sig, btcec.S256())
	if err != nil {
		return err
	}
	if !s.Verify(msg, k.PublicKey) {
		return errors.New("can't verify")
	}
	return nil
}

// Key implement Public
func (k *PublicBTC) Key() interface{} { return k.PublicKey }

// MarshalYAML implement yaml.Marshaler
func (k *PublicBTC) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.SerializeCompressed()), nil
}

// UnmarshalYAML implement yaml.Unmarshaler
func (k *PublicBTC) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	b, err := base64.RawStdEncoding.DecodeString(r)
	if err != nil {
		return err
	}
	k.PublicKey, err = btcec.ParsePubKey(b, btcec.S256())
	return err
}

// Hash160 returns the hash160 of the key
func (k *PublicBTC) Hash160() []byte { return hash.NewBTC().Hash160(k.SerializeCompressed()) }

// KeyData implement Public
func (k *PublicBTC) KeyData() KeyData { return k.Hash160() }
