package atomicswap

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/btcsuite/btcd/wire"
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
	Duration  types.Duration   // trade contract duration
	token     types.Bytes      // secret token
	tokenHash types.Bytes      // secret token hash
	Outputs   *Outputs         // output data
	Own       *OwnTradeInfo    // own trade data
	Trader    *TraderTradeInfo // trader trade data
}

func newTrade(role roles.Role, stage stages.Stage, ownCrypto, tradeCrypto params.Crypto) (*Trade, error) {
	r := &Trade{
		Role:   role,
		Stage:  stage,
		Own:    &OwnTradeInfo{Crypto: ownCrypto},
		Trader: &TraderTradeInfo{Crypto: tradeCrypto},
	}
	if err := r.generateKeys(); err != nil {
		return nil, err
	}
	return r, nil
}

func NewBuyerTrade(ownCrypto, tradeCrypto params.Crypto) (*Trade, error) {
	return newTrade(roles.Buyer, stages.SendPublicKeyHash, ownCrypto, tradeCrypto)
}

func NewSellerTrade(ownCrypto, tradeCrypto params.Crypto) (*Trade, error) {
	r, err := newTrade(roles.Seller, stages.ReceivePublicKeyHash, ownCrypto, tradeCrypto)
	if err != nil {
		return nil, err
	}
	if err = r.generateToken(); err != nil {
		return nil, err
	}
	return r, nil
}

type Output struct {
	TxID types.Bytes
	N    uint32
}

type Outputs struct {
	Redeemable  *Output
	Recoverable *Output
}

type OwnTradeInfo struct {
	Crypto          params.Crypto `yaml:"crypto"`
	LastBlockHeight int           `yaml:"last_block_height"`
	RedeemKey       *key.Private  `yaml:"redeem_key,omitempty"`
	RecoveryKey     *key.Private  `yaml:"recover_key,omitempty"`
	LockScript      types.Bytes   `yaml:"lock_script,omitempty"`
}

type TraderTradeInfo struct {
	Crypto          params.Crypto `yaml:"crypto"`
	LastBlockHeight int           `yaml:"last_block_height"`
	RedeemKeyHash   types.Bytes   `yaml:"recover_key_hash,omitempty"`
	LockScript      types.Bytes   `yaml:"lock_script,omitempty"`
}

func (t *Trade) TokenHash() types.Bytes {
	if t.token != nil {
		return hash.Hash160(t.token)
	}
	return t.tokenHash
}

func (t *Trade) Token() types.Bytes                 { return t.token }
func (t *Trade) SetTokenHash(tokenHash types.Bytes) { t.tokenHash = tokenHash }

func (t *Trade) SetToken(token types.Bytes) {
	t.token = token
	t.tokenHash = hash.Hash160(token)
}

var (
	sellerStages = map[stages.Stage]stages.Stage{
		stages.ReceivePublicKeyHash: stages.SendPublicKeyHash,
		stages.SendPublicKeyHash:    stages.SendTokenHash,
		stages.SendTokenHash:        stages.GenerateLockScript,
		stages.GenerateLockScript:   stages.SendLockScript,
		stages.SendLockScript:       stages.ReceiveLockScript,
		stages.ReceiveLockScript:    stages.LockFunds,
		stages.LockFunds:            stages.WaitLockTransaction,
		stages.WaitLockTransaction:  stages.RedeemFunds,
		stages.RedeemFunds:          stages.Done,
	}
	buyerStages = map[stages.Stage]stages.Stage{
		stages.SendPublicKeyHash:     stages.ReceivePublicKeyHash,
		stages.ReceivePublicKeyHash:  stages.ReceiveTokenHash,
		stages.ReceiveTokenHash:      stages.ReceiveLockScript,
		stages.ReceiveLockScript:     stages.GenerateLockScript,
		stages.GenerateLockScript:    stages.SendLockScript,
		stages.SendLockScript:        stages.WaitLockTransaction,
		stages.WaitLockTransaction:   stages.LockFunds,
		stages.LockFunds:             stages.WaitRedeemTransaction,
		stages.WaitRedeemTransaction: stages.RedeemFunds,
		stages.RedeemFunds:           stages.Done,
	}
)

func (t *Trade) NextStage() stages.Stage {
	var stageMap map[stages.Stage]stages.Stage
	if t.Role == roles.Seller {
		stageMap = sellerStages
	} else {
		stageMap = buyerStages
	}
	t.Stage = stageMap[t.Stage]
	return t.Stage
}

func (t *Trade) generateKeys() error {
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

func (t *Trade) generateToken() error {
	rt, err := readRandomToken()
	if err != nil {
		return err
	}
	t.token = rt
	t.tokenHash = hash.Hash160(t.token)
	return nil
}

var ErrNotEnoughBytes = errors.New("not enough bytes")

const tokenSize = 32

func readRandom(n int) ([]byte, error) {
	r := make([]byte, n)
	if sz, err := rand.Read(r); err != nil {
		return nil, err
	} else if sz != len(r) {
		return nil, ErrNotEnoughBytes
	}
	return r, nil
}

func readRandomToken() ([]byte, error) { return readRandom(tokenSize) }

func (t *Trade) GenerateOwnLockScript() error {
	var lockTime time.Time
	if t.Trader.LockScript == nil {
		lockTime = time.Now().UTC().Add(time.Duration(t.Duration))
	} else {
		lst, err := t.Trader.LockScriptTime()
		if err != nil {
			return err
		}
		lockTime = lst.Add(-(time.Duration(t.Duration) / 2))
	}
	r, err := script.Validate(script.HTLC(
		script.LockTimeTime(lockTime),
		t.tokenHash,
		script.P2PKHHash(t.Own.RecoveryKey.Public().Hash160()),
		script.P2PKHHash(t.Trader.RedeemKeyHash),
	))
	if err != nil {
		return err
	}
	t.Own.LockScript = r
	return nil
}

func (tti *TraderTradeInfo) LockScriptTime() (time.Time, error) {
	lsd, err := parseLockScript(tti.LockScript)
	if err != nil {
		return time.Time{}, err
	}
	return lsd.timeLock, nil
}

var ErrInvalidLockScript = errors.New("invalid lock script")

var expHTLC = []string{
	"OP_IF",
	"", "OP_CHECKLOCKTIMEVERIFY", "OP_DROP",
	"OP_DUP", "OP_HASH160", "", "OP_EQUALVERIFY", "OP_CHECKSIG",
	"OP_ELSE",
	"OP_HASH160", "", "OP_EQUALVERIFY",
	"OP_DUP", "OP_HASH160", "", "OP_EQUALVERIFY", "OP_CHECKSIG", "OP_ENDIF",
}

type lockScriptData struct {
	timeLock        time.Time
	tokenHash       []byte
	redeemKeyHash   []byte
	recoveryKeyHash []byte
}

func parseLockScript(ls []byte) (*lockScriptData, error) {
	r := &lockScriptData{}
	// check contract format
	inst, err := script.DisassembleStrings(ls)
	if err != nil {
		return nil, err
	}
	if len(inst) != len(expHTLC) {
		return nil, ErrInvalidLockScript
	}
	for i, op := range inst {
		if expHTLC[i] == "" {
			continue
		}
		if op != expHTLC[i] {
			return nil, ErrInvalidLockScript
		}
	}
	// time lock
	b, err := hex.DecodeString(inst[1])
	if err != nil {
		return nil, err
	}
	n, err := script.ParseInt64(b)
	if err != nil {
		return nil, err
	}
	r.timeLock = time.Unix(n, 0)
	// token hash
	if r.tokenHash, err = hex.DecodeString(inst[11]); err != nil {
		return nil, err
	}
	// redeem key hash
	if r.redeemKeyHash, err = hex.DecodeString(inst[15]); err != nil {
		return nil, err
	}
	return r, nil
}

func (t *Trade) CheckTraderLockScript(tradeLockScript []byte) error {
	lsd, err := parseLockScript(tradeLockScript)
	if err != nil {
		return err
	}
	if t.Duration != 0 && time.Now().UTC().Add(time.Duration(t.Duration)).After(lsd.timeLock) {
		return ErrInvalidLockScript
	}
	if !bytes.Equal(lsd.tokenHash, t.tokenHash) {
		return ErrInvalidLockScript
	}
	if !bytes.Equal(lsd.redeemKeyHash, t.Own.RedeemKey.Public().Hash160()) {
		return ErrInvalidLockScript
	}
	return nil
}

func bytesJoin(b ...[]byte) []byte { return bytes.Join(b, []byte{}) }

func (t *Trade) RedeemTransaction(amount int64) (*types.Tx, error) {
	r := types.NewTx()
	redeemScript, err := script.Validate(bytesJoin(
		script.Data(t.token),
		script.Int64(0),
		script.Data(t.Trader.LockScript),
	))
	if err != nil {
		return nil, err
	}
	r.AddOutput(amount, script.P2PKHHash(t.Own.RecoveryKey.Public().Hash160()))
	r.AddInput(t.Outputs.Redeemable.TxID, t.Outputs.Redeemable.N, t.Trader.LockScript)
	sig, err := r.InputSignature(0, 1, t.Own.RedeemKey.PrivateKey)
	if err != nil {
		return nil, err
	}
	r.SetP2SHInputSignatureScript(0, bytesJoin(script.Data(sig), script.Data(t.Own.RedeemKey.Public().SerializeCompressed()), redeemScript))
	return r, nil
}

func (t *Trade) RecoveryTransaction(amount int64) (*types.Tx, error) {
	r := types.NewTx()
	r.AddOutput(amount, script.P2PKHHash(t.Own.RecoveryKey.Public().Hash160()))
	r.AddInput(t.Outputs.Recoverable.TxID, t.Outputs.Recoverable.N, t.Own.LockScript)
	lst, err := t.Trader.LockScriptTime()
	if err != nil {
		return nil, err
	}
	r.Tx().LockTime = uint32(lst.UTC().Unix())
	r.Tx().TxIn[0].Sequence = wire.MaxTxInSequenceNum - 1
	sig, err := r.InputSignature(0, 1, t.Own.RecoveryKey.PrivateKey)
	if err != nil {
		return nil, err
	}
	r.SetP2SHInputPrefixes(0,
		sig,
		t.Own.RecoveryKey.Public().SerializeCompressed(),
		[]byte{1},
	)
	return r, nil
}
