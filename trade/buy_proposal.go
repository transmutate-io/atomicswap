package trade

import (
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/duration"
	"github.com/transmutate-io/atomicswap/key"
	"github.com/transmutate-io/cryptocore/types"
	"github.com/transmutate-io/reflection"
	"gopkg.in/yaml.v2"
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
	TokenHash       types.Bytes      `yaml:"token_hash"`
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
