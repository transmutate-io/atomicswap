package key

import "github.com/btcsuite/btcd/btcec"

type privateLTC = privateBTC

func NewPrivateLTC() (Private, error) {
	k, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	return &privateLTC{PrivateKey: k}, nil

}

type publicLTC = publicBTC

func NewPublicLTC(b []byte) (Public, error) {
	pub, err := btcec.ParsePubKey(b, btcec.S256())
	if err != nil {
		return nil, err
	}
	return &publicLTC{PublicKey: pub}, nil
}
