package trade

import (
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/reflection"
	"gopkg.in/yaml.v2"
)

// Locks is returned when a buy proposal is accepted
type Locks struct {
	Buyer  Lock `yaml:"buyer"`
	Seller Lock `yaml:"seller"`
}

// UnamrshalLocks unmarshals a buy proposal response
func UnamrshalLocks(buyer, seller *cryptos.Crypto, b []byte) (*Locks, error) {
	r := &Locks{}
	// replace fields
	replaceMap := make(map[string]reflection.Any, 2)
	fd, err := newFundsData(buyer)
	if err != nil {
		return nil, err
	}
	replaceMap["Buyer"] = fd.Lock()
	if fd, err = newFundsData(seller); err != nil {
		return nil, err
	}
	replaceMap["Seller"] = fd.Lock()
	v, err := reflection.ReplaceFieldsType(r, replaceMap)
	if err != nil {
		return nil, err
	}
	// unmarshal
	if err = yaml.Unmarshal(b, v); err != nil {
		return nil, err
	}
	// copy fields
	if err = reflection.CopyFields(v, r); err != nil {
		return nil, err
	}
	return r, nil
}
