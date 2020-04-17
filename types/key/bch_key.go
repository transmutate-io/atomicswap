package key

import (
	"encoding/base64"
	"errors"

	"github.com/gcash/bchd/bchec"
	"transmutate.io/pkg/atomicswap/hash"
)

type privateBCH struct{ *bchec.PrivateKey }

func ParsePrivateBCH(b []byte) Private {
	priv, _ := bchec.PrivKeyFromBytes(bchec.S256(), b)
	return &privateBCH{PrivateKey: priv}
}

func NewPrivateBCH() (Private, error) {
	k, err := bchec.NewPrivateKey(bchec.S256())
	if err != nil {
		return nil, err
	}
	return &privateBCH{PrivateKey: k}, nil
}

func (k *privateBCH) Sign(b []byte) ([]byte, error) {
	sig, err := k.PrivateKey.SignECDSA(b)
	if err != nil {
		return nil, err
	}
	return sig.Serialize(), nil
}

func (k *privateBCH) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.Serialize()), nil
}

func (k *privateBCH) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	b, err := base64.RawStdEncoding.DecodeString(r)
	if err != nil {
		return err
	}
	k.PrivateKey, _ = bchec.PrivKeyFromBytes(bchec.S256(), b)
	return nil
}

func (k *privateBCH) Public() Public   { return &publicBCH{k.PubKey()} }
func (k *privateBCH) Key() interface{} { return k.PrivateKey }

type publicBCH struct{ *bchec.PublicKey }

func NewPublicBCH(b []byte) (Public, error) {
	pub, err := bchec.ParsePubKey(b, bchec.S256())
	if err != nil {
		return nil, err
	}
	return &publicBCH{PublicKey: pub}, nil
}

func (k *publicBCH) Verify(sig, msg []byte) error {
	s, err := bchec.ParseDERSignature(sig, bchec.S256())
	if err != nil {
		return err
	}
	if !s.Verify(msg, k.PublicKey) {
		return errors.New("can't verify")
	}
	return nil
}

func (k *publicBCH) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.SerializeCompressed()), nil
}

func (k *publicBCH) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

func (k *publicBCH) Hash160() []byte { return hash.Hash160(k.SerializeCompressed()) }

func (k *publicBCH) Key() interface{} { return k.PublicKey }
