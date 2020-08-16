package key

import (
	"encoding/base64"
	"errors"

	"github.com/gcash/bchd/bchec"
	"github.com/transmutate-io/atomicswap/hash"
)

// PrivateBCH represents a private key for bitcoin-cash
type PrivateBCH struct{ *bchec.PrivateKey }

func parsePrivateBCH(b []byte) *PrivateBCH {
	priv, _ := bchec.PrivKeyFromBytes(bchec.S256(), b)
	return &PrivateBCH{PrivateKey: priv}
}

// ParsePrivateBCH parses a bitcoin-cash private key
func ParsePrivateBCH(b []byte) (Private, error) { return parsePrivateBCH(b), nil }

func newPrivateBCH() (*PrivateBCH, error) {
	k, err := bchec.NewPrivateKey(bchec.S256())
	if err != nil {
		return nil, err
	}
	return &PrivateBCH{PrivateKey: k}, nil
}

// NewPrivateBCH returns a new bitcoin-cash private key
func NewPrivateBCH() (Private, error) { return newPrivateBCH() }

// Sign implement Private
func (k *PrivateBCH) Sign(b []byte) ([]byte, error) {
	sig, err := k.PrivateKey.SignECDSA(b)
	if err != nil {
		return nil, err
	}
	return sig.Serialize(), nil
}

// MarshalYAML implement yaml.Marshaler
func (k *PrivateBCH) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.Serialize()), nil
}

// UnmarshalYAML implement yaml.Unmarshaler
func (k *PrivateBCH) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	b, err := base64.RawStdEncoding.DecodeString(r)
	if err != nil {
		return err
	}
	k.PrivateKey = parsePrivateBCH(b).PrivateKey
	return nil
}

// Public implement Private
func (k *PrivateBCH) Public() Public { return &PublicBCH{k.PubKey()} }

// Key implement Private
func (k *PrivateBCH) Key() interface{} { return k.PrivateKey }

// PublicBCH represents a bitcoin-cash public key
type PublicBCH struct{ *bchec.PublicKey }

func parsePublicBCH(b []byte) (*PublicBCH, error) {
	pub, err := bchec.ParsePubKey(b, bchec.S256())
	if err != nil {
		return nil, err
	}
	return &PublicBCH{PublicKey: pub}, nil
}

// ParsePublicBCH parses a bitcoin-cash public key
func ParsePublicBCH(b []byte) (Public, error) { return parsePublicBCH(b) }

// Verify implement Public
func (k *PublicBCH) Verify(sig, msg []byte) error {
	s, err := bchec.ParseDERSignature(sig, bchec.S256())
	if err != nil {
		return err
	}
	if !s.Verify(msg, k.PublicKey) {
		return errors.New("can't verify")
	}
	return nil
}

// Key implement Public
func (k *PublicBCH) Key() interface{} { return k.PublicKey }

// MarshalYAML implement yaml.Marshaler
func (k *PublicBCH) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.SerializeCompressed()), nil
}

// UnmarshalYAML implement yaml.Unmarshaler
func (k *PublicBCH) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	b, err := base64.RawStdEncoding.DecodeString(r)
	if err != nil {
		return err
	}
	k.PublicKey, err = bchec.ParsePubKey(b, bchec.S256())
	return err
}

// Hash160 returns the hash160 of the key
func (k *PublicBCH) Hash160() []byte { return hash.NewBCH().Hash160(k.SerializeCompressed()) }

// KeyData implement Public
func (k *PublicBCH) KeyData() KeyData { return k.Hash160() }
