package atomicswap

import (
	"crypto/rand"
	"errors"
	"time"

	"transmutate.io/pkg/atomicswap/hash"
	"transmutate.io/pkg/atomicswap/params"
	"transmutate.io/pkg/atomicswap/script"
	"transmutate.io/pkg/atomicswap/types"
	"transmutate.io/pkg/atomicswap/types/key"
	"transmutate.io/pkg/atomicswap/types/roles"
	"transmutate.io/pkg/atomicswap/types/stages"
)

type Trade struct {
	Stage     stages.Stage     // stage of the trade
	Role      roles.Role       // role
	token     types.Bytes      // secret token
	tokenHash types.Bytes      // secret token hash
	Own       *OwnTradeInfo    // own trade data
	Trader    *TraderTradeInfo // trader trade data
}

type OwnTradeInfo struct {
	// owned crypto
	Crypto params.Crypto `yaml:"own_crypto"`
	// own redeem private key
	RedeemKey *key.Private `yaml:"redeem_key,omitempty"`
	// own recovery key
	RecoveryKey *key.Private `yaml:"recover_key,omitempty"`
	// own lock script
	LockScript types.Bytes `yaml:"own_lock_script,omitempty"`
}

type TraderTradeInfo struct {
	// crypto to acquire
	Crypto params.Crypto `yaml:"trader_crypto"`
	// trade redeem key hash
	RedeemKeyHash types.Bytes `yaml:"trader_recover_key_hash,omitempty"`
	// trade recover key hash
	RecoveryKeyHash types.Bytes `yaml:"trader_recover_key_hash,omitempty"`
	// trade lock script
	LockScript types.Bytes `yaml:"trader_lock_script,omitempty"`
}

func NewSellerTrade() *Trade { return &Trade{Role: roles.Seller} }
func NewBuyerTrade() *Trade  { return &Trade{Role: roles.Seller} }

func (t *Trade) TokenHash() types.Bytes {
	if t.token != nil {
		return hash.Hash160(t.token)
	}
	return t.tokenHash
}

func (t *Trade) Token() types.Bytes                 { return t.token }
func (t *Trade) SetTokenHash(tokenHash types.Bytes) { t.tokenHash = tokenHash }
func (t *Trade) SetToken(token types.Bytes)         { t.token = token }

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
	if t.Own.RecoveryKey, err = key.NewPrivate(); err != nil {
		return err
	}
	if t.Own.RedeemKey, err = key.NewPrivate(); err != nil {
		return err
	}
	if t.Role == roles.Seller {
		if t.token, err = readRandomToken(); err != nil {
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

func (t *Trade) GenerateOwnLockScript(lockTime time.Time) (types.Bytes, error) {
	return script.Validate(script.HTLC(
		script.LockTimeTime(lockTime),
		nil, // tokenHash,
		nil, // tlsc,
		nil, // hlsc,
	))
}

var ErrInvalidLockScript = errors.New("invalid lock script")

var expHTLC = []string{
	"OP_IF", "", "OP_CHECKLOCKTIMEVERIFY", "OP_DUP", "OP_HASH160", "",
	"OP_EQUALVERIFY", "OP_CHECKSIG", "OP_ELSE", "OP_HASH160", "", "OP_EQUALVERIFY",
	"OP_DUP", "OP_HASH160", "", "OP_EQUALVERIFY", "OP_CHECKSIG", "OP_ENDIF",
}

func (t *Trade) CheckTradeLockScript(tradeLockScript []byte) error {
	inst, err := script.DisassembleStrings(tradeLockScript)
	if err != nil {
		return err
	}
	if len(inst) != len(expHTLC) {
		return ErrInvalidLockScript
	}
	for i, op := range inst {
		if expHTLC[i] == "" {
			continue
		}
		if op != expHTLC[i] {
			return ErrInvalidLockScript
		}
	}
	return nil
}

func (t *Trade) GenerateRedeemScript() ([]byte, error) {
	s, err := script.DisassembleStrings(t.Trader.LockScript)
	if err != nil {
		return nil, err
	}
	_ = s
	return nil, nil
}
