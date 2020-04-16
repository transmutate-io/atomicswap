package key

import "github.com/btcsuite/btcd/btcec"

type privateDOGE = privateBTC

func NewPrivateDOGE() (Private, error) {
	k, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	return &privateDOGE{PrivateKey: k}, nil

}

type publicDOGE = publicBTC

func NewPublicDOGE(b []byte) (Public, error) {
	pub, err := btcec.ParsePubKey(b, btcec.S256())
	if err != nil {
		return nil, err
	}
	return &publicDOGE{PublicKey: pub}, nil
}
