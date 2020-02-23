package key

import (
	"encoding/base64"

	"github.com/btcsuite/btcd/btcec"
	"transmutate.io/pkg/atomicswap/hash"
)

type Private struct{ *btcec.PrivateKey }

func NewPrivate() (*Private, error) {
	k, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	return &Private{PrivateKey: k}, nil
}

func (k *Private) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.Serialize()), nil
}

func (k *Private) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	b, err := base64.RawStdEncoding.DecodeString(r)
	if err != nil {
		return err
	}
	k.PrivateKey, _ = btcec.PrivKeyFromBytes(btcec.S256(), b)
	return nil
}

func (k *Private) Public() *Public { return &Public{k.PubKey()} }

type Public struct{ *btcec.PublicKey }

func NewPublic(b []byte) (*Public, error) {
	pub, err := btcec.ParsePubKey(b, btcec.S256())
	if err != nil {
		return nil, err
	}
	return &Public{PublicKey: pub}, nil
}

func (k *Public) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.SerializeCompressed()), nil
}

func (k *Public) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	b, err := base64.RawStdEncoding.DecodeString(r)
	if err != nil {
		return err
	}
	k.PublicKey, _ = btcec.ParsePubKey(b, btcec.S256())
	return nil
}

func (k *Public) Hash160() []byte { return hash.Hash160(k.SerializeCompressed()) }
