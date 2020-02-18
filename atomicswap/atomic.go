package atomicswap

import (
	"crypto/rand"
	"errors"

	"transmutate.io/pkg/swapper/params"
	"transmutate.io/pkg/swapper/types"
	"transmutate.io/pkg/swapper/types/key"
	"transmutate.io/pkg/swapper/types/roles"
	"transmutate.io/pkg/swapper/types/stages"
)

type Trade struct {
	// stage of the trade
	Stage stages.Stage `yaml:"stage"`
	// role
	Role roles.Role `yaml:"role"`
	// owned crypto
	OwnCrypto params.Crypto `yaml:"own_crypto"`
	// crypto to acquire
	TraderCrypto params.Crypto `yaml:"trader_crypto"`
	// own redeem private key
	RedeemKey *key.Private `yaml:"redeem_key"`
	// own recovery key
	RecoverKey *key.Private `yaml:"recover_key"`
	// trade redeem key
	TraderRedeemKey types.Bytes `yaml:"trader_recover_key"`
	// trade recover key
	TraderRecoverKey types.Bytes `yaml:"trader_recover_key"`
	// secret token
	Token types.Bytes `yaml:"token,omitempty"`
	// secret token hash
	TokenHash types.Bytes `yaml:"token_hash,omitempty"`
	// own lock script
	OwnLockScript types.Bytes `yaml:"own_lock_script,omitempty"`
	// trade lock script
	TraderLockScript types.Bytes `yaml:"trader_lock_script,omitempty"`
}

func (t *Trade) NextStage() stages.Stage {
	var stageMap map[stages.Stage]stages.Stage
	if t.Role == roles.Seller {
		stageMap = stages.SellerStages
	} else {
		stageMap = stages.BuyerStages
	}
	t.Stage = stageMap[t.Stage]
	return t.Stage
}

func (t *Trade) GenerateSecrets() error {
	var err error
	if t.RecoverKey, err = key.NewPrivate(); err != nil {
		return err
	}
	if t.RedeemKey, err = key.NewPrivate(); err != nil {
		return err
	}
	if t.Role == roles.Seller {
		if t.Token, err = readRandomToken(); err != nil {
			return err
		}
	}
	return nil
}

var ErrNotEnoughBytes = errors.New("not enough bytes")

const TokenSize = 16

func readRandom(n int) ([]byte, error) {
	r := make([]byte, n)
	if sz, err := rand.Read(r); err != nil {
		return nil, err
	} else if sz != len(r) {
		return nil, ErrNotEnoughBytes
	}
	return r, nil
}

func readRandomToken() ([]byte, error) { return readRandom(TokenSize) }
