package trade

import (
	"gopkg.in/yaml.v2"
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/duration"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/cryptocore/types"
	"transmutate.io/pkg/reflection"
)

// BuyProposalInfo represents a buy proposal trader info
type BuyProposalInfo struct {
	Crypto       *cryptos.Crypto   `yaml:"crypto"`
	Amount       types.Amount      `yaml:"amount"`
	LockDuration duration.Duration `yaml:"lock_duration"`
}

// BuyProposal represents a buy proposal
type BuyProposal struct {
	Buyer           *BuyProposalInfo `yaml:"buyer"`
	Seller          *BuyProposalInfo `yaml:"seller"`
	TokenHash       []byte           `yaml:"token_hash"`
	RedeemKeyData   key.KeyData      `yaml:"redeem_key_data"`
	RecoveryKeyData key.KeyData      `yaml:"recovery_key_data"`
}

// UnamrshalBuyProposal unmarshals a buy proposal
func UnamrshalBuyProposal(b []byte) (*BuyProposal, error) {
	// find which cryptos first
	type (
		bpci struct {
			C *cryptos.Crypto `yaml:"crypto"`
		}
		bpc struct {
			Buyer  *bpci `yaml:"buyer"`
			Seller *bpci `yaml:"seller"`
		}
	)
	tc := &bpc{}
	if err := yaml.Unmarshal(b, tc); err != nil {
		return nil, err
	}
	// replace fields
	r := &BuyProposal{}
	t, err := reflection.ReplaceFieldsType(r, reflection.FieldReplacementMap{
		"RedeemKeyData":   types.Bytes{},
		"RecoveryKeyData": types.Bytes{},
	})
	if err != nil {
		return nil, err
	}
	// unmarshal
	if err = yaml.Unmarshal(b, t); err != nil {
		return nil, err
	}
	// copy fields
	if err = reflection.CopyFields(t, r); err != nil {
		return nil, err
	}
	return r, nil
}

// BuyProposalResponse is returned when a buy proposal is accepted
type BuyProposalResponse struct {
	Buyer  Lock `yaml:"buyer"`
	Seller Lock `yaml:"seller"`
}

// UnamrshalBuyProposalResponse unmarshals a buy proposal response
func UnamrshalBuyProposalResponse(buyer, seller *cryptos.Crypto, b []byte) (*BuyProposalResponse, error) {
	r := &BuyProposalResponse{}
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
