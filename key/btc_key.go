package key

import (
	"encoding/base64"
	"errors"

	"github.com/btcsuite/btcd/btcec"
	"transmutate.io/pkg/atomicswap/hash"
)

type privateBTC struct{ *btcec.PrivateKey }

func ParsePrivateBTC(b []byte) Private {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), b)
	return &privateBTC{PrivateKey: priv}
}

func NewPrivateBTC() (Private, error) {
	k, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	return &privateBTC{PrivateKey: k}, nil
}

func (k *privateBTC) Sign(b []byte) ([]byte, error) {
	sig, err := k.PrivateKey.Sign(b)
	if err != nil {
		return nil, err
	}
	return sig.Serialize(), nil
}

func (k *privateBTC) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.Serialize()), nil
}

func (k *privateBTC) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

func (k *privateBTC) Public() Public { return &publicBTC{k.PubKey()} }

func (k *privateBTC) Key() interface{} { return k.PrivateKey }

type publicBTC struct{ *btcec.PublicKey }

func NewPublicBTC(b []byte) (Public, error) {
	pub, err := btcec.ParsePubKey(b, btcec.S256())
	if err != nil {
		return nil, err
	}
	return &publicBTC{PublicKey: pub}, nil
}

func (k *publicBTC) Verify(sig, msg []byte) error {
	s, err := btcec.ParseSignature(sig, btcec.S256())
	if err != nil {
		return err
	}
	if !s.Verify(msg, k.PublicKey) {
		return errors.New("can't verify")
	}
	return nil
}

func (k *publicBTC) Key() interface{} { return k.PublicKey }

func (k *publicBTC) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.SerializeCompressed()), nil
}

func (k *publicBTC) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

func (k *publicBTC) Hash160() []byte { return hash.Hash160(k.SerializeCompressed()) }
