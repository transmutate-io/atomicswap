package key

import (
	"encoding/base64"
	"errors"

	"github.com/gcash/bchd/bchec"
	"github.com/transmutate-io/atomicswap/hash"
)

type PrivateBCH struct{ *bchec.PrivateKey }

func parsePrivateBCH(b []byte) *PrivateBCH {
	priv, _ := bchec.PrivKeyFromBytes(bchec.S256(), b)
	return &PrivateBCH{PrivateKey: priv}
}

func ParsePrivateBCH(b []byte) (Private, error) { return parsePrivateBCH(b), nil }

func newPrivateBCH() (*PrivateBCH, error) {
	k, err := bchec.NewPrivateKey(bchec.S256())
	if err != nil {
		return nil, err
	}
	return &PrivateBCH{PrivateKey: k}, nil
}

func NewPrivateBCH() (Private, error) { return newPrivateBCH() }

func (k *PrivateBCH) Sign(b []byte) ([]byte, error) {
	sig, err := k.PrivateKey.SignECDSA(b)
	if err != nil {
		return nil, err
	}
	return sig.Serialize(), nil
}

func (k *PrivateBCH) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.Serialize()), nil
}

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

func (k *PrivateBCH) Public() Public { return &PublicBCH{k.PubKey()} }

func (k *PrivateBCH) Key() interface{} { return k.PrivateKey }

type PublicBCH struct{ *bchec.PublicKey }

func newPublicBCH(b []byte) (*PublicBCH, error) {
	pub, err := bchec.ParsePubKey(b, bchec.S256())
	if err != nil {
		return nil, err
	}
	return &PublicBCH{PublicKey: pub}, nil
}

func NewPublicBCH(b []byte) (Public, error) { return newPublicBCH(b) }

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

func (k *PublicBCH) Key() interface{} { return k.PublicKey }

func (k *PublicBCH) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.SerializeCompressed()), nil
}

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

func (k *PublicBCH) Hash160() []byte { return hash.NewBCH().Hash160(k.SerializeCompressed()) }

func (k *PublicBCH) KeyData() KeyData { return k.Hash160() }
