package key

import (
	"encoding/base64"
	"errors"

	"github.com/btcsuite/btcd/btcec"
	"transmutate.io/pkg/atomicswap/hash"
)

type PrivateBTC struct{ *btcec.PrivateKey }

func parsePrivateBTC(b []byte) *PrivateBTC {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), b)
	return &PrivateBTC{PrivateKey: priv}
}

func ParsePrivateBTC(b []byte) (Private, error) { return parsePrivateBTC(b), nil }

func newPrivateBTC() (*PrivateBTC, error) {
	k, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	return &PrivateBTC{PrivateKey: k}, nil
}

func NewPrivateBTC() (Private, error) { return newPrivateBTC() }

func (k *PrivateBTC) Sign(b []byte) ([]byte, error) {
	sig, err := k.PrivateKey.Sign(b)
	if err != nil {
		return nil, err
	}
	return sig.Serialize(), nil
}

func (k *PrivateBTC) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.Serialize()), nil
}

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

func (k *PrivateBTC) Public() Public { return &PublicBTC{k.PubKey()} }

func (k *PrivateBTC) Key() interface{} { return k.PrivateKey }

type PublicBTC struct{ *btcec.PublicKey }

func newPublicBTC(b []byte) (*PublicBTC, error) {
	pub, err := btcec.ParsePubKey(b, btcec.S256())
	if err != nil {
		return nil, err
	}
	return &PublicBTC{PublicKey: pub}, nil
}

func NewPublicBTC(b []byte) (Public, error) { return newPublicBTC(b) }

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

func (k *PublicBTC) Key() interface{} { return k.PublicKey }

func (k *PublicBTC) MarshalYAML() (interface{}, error) {
	return base64.RawStdEncoding.EncodeToString(k.SerializeCompressed()), nil
}

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

func (k *PublicBTC) Hash160() []byte { return hash.Hash160(k.SerializeCompressed()) }

func (k *PublicBTC) KeyData() KeyData { return k.Hash160() }
