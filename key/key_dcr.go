package key

import (
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/decred/dcrd/chaincfg/chainec"
	"github.com/transmutate-io/atomicswap/hash"
)

type PrivateDCR struct{ chainec.PrivateKey }

func parsePrivateDCR(b []byte) *PrivateDCR {
	priv, _ := chainec.Secp256k1.PrivKeyFromBytes(b)
	return &PrivateDCR{PrivateKey: priv}
}

func ParsePrivateDCR(b []byte) (Private, error) { return parsePrivateDCR(b), nil }

func newPrivateDCR() (*PrivateDCR, error) {
	b, _, _, err := chainec.Secp256k1.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	priv, _ := chainec.Secp256k1.PrivKeyFromBytes(b)
	return &PrivateDCR{PrivateKey: priv}, nil
}

func NewPrivateDCR() (Private, error) { return newPrivateDCR() }

func (k *PrivateDCR) Sign(b []byte) ([]byte, error) {
	r, s, err := chainec.Secp256k1.Sign(k.PrivateKey, b)
	if err != nil {
		return nil, err
	}
	return chainec.Secp256k1.NewSignature(r, s).Serialize(), nil
}

func (k *PrivateDCR) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.Serialize()), nil
}

func (k *PrivateDCR) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	b, err := base64.RawStdEncoding.DecodeString(r)
	if err != nil {
		return err
	}
	k.PrivateKey = parsePrivateDCR(b).PrivateKey
	return nil
}

func (k *PrivateDCR) Public() Public {
	x, y := k.PrivateKey.Public()
	return &PublicDCR{chainec.Secp256k1.NewPublicKey(x, y)}
}

func (k *PrivateDCR) Key() interface{} { return k.PrivateKey }

type PublicDCR struct{ chainec.PublicKey }

func newPublicDCR(b []byte) (*PublicDCR, error) {
	pub, err := chainec.Secp256k1.ParsePubKey(b)
	if err != nil {
		return nil, err
	}
	return &PublicDCR{PublicKey: pub}, nil
}

func NewPublicDCR(b []byte) (Public, error) { return newPublicDCR(b) }

func (k *PublicDCR) Verify(sig, msg []byte) error {
	s, err := chainec.Secp256k1.ParseSignature(sig)
	if err != nil {
		return err
	}
	if !chainec.Secp256k1.Verify(k.PublicKey, msg, s.GetR(), s.GetS()) {
		return errors.New("signature mismatch")
	}
	return nil
}

func (k *PublicDCR) Key() interface{} { return k.PublicKey }

func (k *PublicDCR) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.SerializeCompressed()), nil
}

func (k *PublicDCR) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	b, err := base64.RawStdEncoding.DecodeString(r)
	if err != nil {
		return err
	}
	k.PublicKey, err = chainec.Secp256k1.ParsePubKey(b)
	return err
}

func (k *PublicDCR) Hash160() []byte { return hash.NewDCR().Hash160(k.SerializeCompressed()) }

func (k *PublicDCR) KeyData() KeyData { return k.Hash160() }
