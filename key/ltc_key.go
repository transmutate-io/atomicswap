package key

import "github.com/btcsuite/btcd/btcec"

type PrivateLTC = PrivateBTC

func ParsePrivateLTC(b []byte) Private {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), b)
	return &PrivateLTC{PrivateKey: priv}
}

func NewPrivateLTC() (Private, error) {
	k, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	return &PrivateLTC{PrivateKey: k}, nil

}

type PublicLTC = PublicBTC

func NewPublicLTC(b []byte) (Public, error) {
	pub, err := btcec.ParsePubKey(b, btcec.S256())
	if err != nil {
		return nil, err
	}
	return &PublicLTC{PublicKey: pub}, nil
}
