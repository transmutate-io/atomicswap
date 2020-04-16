package key

import (
	"encoding/base64"
	"errors"

	"github.com/gcash/bchd/bchec"
	"transmutate.io/pkg/atomicswap/hash"
)

type privateBTCCash struct{ *bchec.PrivateKey }

func NewPrivateBTCCash() (Private, error) {
	k, err := bchec.NewPrivateKey(bchec.S256())
	if err != nil {
		return nil, err
	}
	return &privateBTCCash{PrivateKey: k}, nil
}

func (k *privateBTCCash) Sign(b []byte) ([]byte, error) {
	sig, err := k.PrivateKey.SignECDSA(b)
	if err != nil {
		return nil, err
	}
	return sig.Serialize(), nil
}

func (k *privateBTCCash) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.Serialize()), nil
}

func (k *privateBTCCash) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

func (k *privateBTCCash) Public() Public { return &publicBTCCash{k.PubKey()} }

type publicBTCCash struct{ *bchec.PublicKey }

func NewPublicBTCCash(b []byte) (Public, error) {
	pub, err := bchec.ParsePubKey(b, bchec.S256())
	if err != nil {
		return nil, err
	}
	return &publicBTCCash{PublicKey: pub}, nil
}

func (k *privateBTCCash) Key() interface{} { return k.PublicKey }

func (k *publicBTCCash) Verify(sig, msg []byte) error {
	s, err := bchec.ParseDERSignature(sig, bchec.S256())
	if err != nil {
		return err
	}
	if !s.Verify(msg, k.PublicKey) {
		return errors.New("can't verify")
	}
	return nil
}

func (k *publicBTCCash) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.SerializeCompressed()), nil
}

func (k *publicBTCCash) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	b, err := base64.RawStdEncoding.DecodeString(r)
	if err != nil {
		return err
	}
	k.PublicKey, _ = bchec.ParsePubKey(b, bchec.S256())
	return nil
}

func (k *publicBTCCash) Hash160() []byte { return hash.Hash160(k.SerializeCompressed()) }

func (k *publicBTCCash) Key() interface{} { return k.PublicKey }
