package key

import "github.com/btcsuite/btcd/btcec"

type PrivateDOGE = PrivateBTC

func ParsePrivateDOGE(b []byte) Private {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), b)
	return &PrivateDOGE{PrivateKey: priv}
}

func NewPrivateDOGE() (Private, error) {
	k, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	return &PrivateDOGE{PrivateKey: k}, nil

}

type PublicDOGE = PublicBTC

func NewPublicDOGE(b []byte) (Public, error) {
	pub, err := btcec.ParsePubKey(b, btcec.S256())
	if err != nil {
		return nil, err
	}
	return &PublicDOGE{PublicKey: pub}, nil
}
