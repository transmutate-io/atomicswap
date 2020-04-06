package atomicswap

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/btcsuite/btcd/wire"
	"transmutate.io/pkg/atomicswap/hash"
	"transmutate.io/pkg/atomicswap/params"
	"transmutate.io/pkg/atomicswap/roles"
	"transmutate.io/pkg/atomicswap/script"
	"transmutate.io/pkg/atomicswap/stages"
	"transmutate.io/pkg/atomicswap/types"
	"transmutate.io/pkg/atomicswap/types/key"
	bctypes "transmutate.io/pkg/btccore/types"
)

// Trade represents an atomic swap trade
type Trade struct {
	// Stage of the trade
	Stage stages.Stage
	// Role on the trade
	Role roles.Role
	// Duration represents the trade lock time
	Duration  types.Duration
	token     bctypes.Bytes
	tokenHash bctypes.Bytes
	// Outputs contains the outputs involved
	Outputs *Outputs
	// Own contains own user data and keys
	Own *OwnTradeInfo
	// Trader contrains the trader data
	Trader *TraderTradeInfo
}

type tradeData struct {
	Stage     stages.Stage     `yaml:"stage"`
	Role      roles.Role       `yaml:"role"`
	Duration  types.Duration   `yaml:"duration"`
	Outputs   *Outputs         `yaml:"outputs,omitempty"`
	Own       *OwnTradeInfo    `yaml:"own,omitempty"`
	Trader    *TraderTradeInfo `yaml:"trader,omitempty"`
	Token     bctypes.Bytes    `yaml:"token,omitempty"`
	TokenHash bctypes.Bytes    `yaml:"token_hash,omitempty"`
}

func newTrade(role roles.Role, stage stages.Stage, ownAmount bctypes.Amount, ownCrypto params.Crypto, tradeAmount bctypes.Amount, tradeCrypto params.Crypto) (*Trade, error) {
	r := &Trade{
		Role:  role,
		Stage: stage,
		Own: &OwnTradeInfo{
			Crypto:          ownCrypto,
			Amount:          ownAmount,
			LastBlockHeight: 1,
		},
		Trader: &TraderTradeInfo{
			Crypto:          tradeCrypto,
			Amount:          tradeAmount,
			LastBlockHeight: 1,
		},
	}
	if err := r.generateKeys(); err != nil {
		return nil, err
	}
	return r, nil
}

// NewBuyerTrade starts a trade as a buyer
func NewBuyerTrade(ownAmount bctypes.Amount, ownCrypto params.Crypto, tradeAmount bctypes.Amount, tradeCrypto params.Crypto) (*Trade, error) {
	return newTrade(
		roles.Buyer,
		stages.SharePublicKeyHash,
		ownAmount,
		ownCrypto,
		tradeAmount,
		tradeCrypto,
	)
}

// NewSellerTrade starts a trade as a seller
func NewSellerTrade(ownAmount bctypes.Amount, ownCrypto params.Crypto, tradeAmount bctypes.Amount, tradeCrypto params.Crypto) (*Trade, error) {
	r, err := newTrade(
		roles.Seller,
		stages.ReceivePublicKeyHash,
		ownAmount,
		ownCrypto,
		tradeAmount,
		tradeCrypto,
	)
	if err != nil {
		return nil, err
	}
	if err = r.generateToken(); err != nil {
		return nil, err
	}
	return r, nil
}

type (
	// Output represents an output
	Output struct {
		TxID bctypes.Bytes `yaml:"txid,omitempty"`
		N    uint32        `yaml:"n"`
	}

	// Outputs represents the outputs involved
	Outputs struct {
		Redeemable  []*Output `yaml:"redeemable,omitempty"`
		Recoverable *Output   `yaml:"recoverable,omitempty"`
	}

	// OwnTradeInfo represents the own user trade info
	OwnTradeInfo struct {
		Crypto          params.Crypto  `yaml:"crypto"`
		Amount          bctypes.Amount `yaml:"amount"`
		LastBlockHeight uint64         `yaml:"last_block_height"`
		RedeemKey       *key.Private   `yaml:"redeem_key,omitempty"`
		RecoveryKey     *key.Private   `yaml:"recover_key,omitempty"`
		LockScript      bctypes.Bytes  `yaml:"lock_script,omitempty"`
	}

	// TraderTradeInfo represents the trader trade info
	TraderTradeInfo struct {
		Crypto          params.Crypto  `yaml:"crypto"`
		Amount          bctypes.Amount `yaml:"amount"`
		LastBlockHeight uint64         `yaml:"last_block_height"`
		RedeemKeyHash   bctypes.Bytes  `yaml:"recover_key_hash,omitempty"`
		LockScript      bctypes.Bytes  `yaml:"lock_script,omitempty"`
	}
)

// TokenHash returns the token hash if set, otherwise nil
func (t *Trade) TokenHash() bctypes.Bytes {
	if t.token != nil {
		return hash.Hash160(t.token)
	}
	return t.tokenHash
}

// Token returns the token if set, otherwise nil
func (t *Trade) Token() bctypes.Bytes { return t.token }

// SetTokenHash sets the token hash
func (t *Trade) SetTokenHash(tokenHash bctypes.Bytes) { t.tokenHash = tokenHash }

// SetToken sets the token
func (t *Trade) SetToken(token bctypes.Bytes) {
	t.token = token
	t.tokenHash = hash.Hash160(token)
}

var (
	sellerStages = map[stages.Stage]stages.Stage{
		stages.ReceivePublicKeyHash: stages.SharePublicKeyHash,
		stages.SharePublicKeyHash:   stages.ShareTokenHash,
		stages.ShareTokenHash:       stages.GenerateLockScript,
		stages.GenerateLockScript:   stages.ShareLockScript,
		stages.ShareLockScript:      stages.ReceiveLockScript,
		stages.ReceiveLockScript:    stages.LockFunds,
		stages.LockFunds:            stages.WaitLockTransaction,
		stages.WaitLockTransaction:  stages.RedeemFunds,
		stages.RedeemFunds:          stages.Done,
	}
	buyerStages = map[stages.Stage]stages.Stage{
		stages.SharePublicKeyHash:    stages.ReceivePublicKeyHash,
		stages.ReceivePublicKeyHash:  stages.ReceiveTokenHash,
		stages.ReceiveTokenHash:      stages.ReceiveLockScript,
		stages.ReceiveLockScript:     stages.GenerateLockScript,
		stages.GenerateLockScript:    stages.ShareLockScript,
		stages.ShareLockScript:       stages.WaitLockTransaction,
		stages.WaitLockTransaction:   stages.LockFunds,
		stages.LockFunds:             stages.WaitRedeemTransaction,
		stages.WaitRedeemTransaction: stages.RedeemFunds,
		stages.RedeemFunds:           stages.Done,
	}
)

// NextStage advance the trade to the next stage
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

// ErrNotEnoughBytes is returned the is not possible to read enough random bytes
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

// GenerateOwnLockScript generates the user own lock script
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

// LockScriptTime returns the lock time from the trader lock script
func (tti *TraderTradeInfo) LockScriptTime() (time.Time, error) {
	lsd, err := parseLockScript(tti.LockScript)
	if err != nil {
		return time.Time{}, err
	}
	return lsd.timeLock, nil
}

// ErrInvalidLockScript is returns when the lock script is invalid
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

// CheckTraderLockScript verifies the trader lock script
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

// RedeemTransaction returns the redeem transaction for the locked funds
func (t *Trade) RedeemTransaction(amount uint64) (*types.Tx, error) {
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
	for ni, i := range t.Outputs.Redeemable {
		if err = r.AddInput(i.TxID, i.N, t.Trader.LockScript); err != nil {
			return nil, err
		}
		sig, err := r.InputSignature(ni, 1, t.Own.RedeemKey)
		if err != nil {
			return nil, err
		}
		r.SetP2SHInputSignatureScript(ni, bytesJoin(script.Data(sig), script.Data(t.Own.RedeemKey.Public().SerializeCompressed()), redeemScript))
	}
	return r, nil
}

// AddRedeemableOutput adds a redeemable output to the trade
func (t *Trade) AddRedeemableOutput(out *Output) {
	if t.Outputs == nil {
		t.Outputs = &Outputs{}
	}
	if t.Outputs.Redeemable == nil {
		t.Outputs.Redeemable = make([]*Output, 0, 4)
	}
	t.Outputs.Redeemable = append(t.Outputs.Redeemable, out)
}

// SetRecoverableOutput sets the recoverable output field
func (t *Trade) SetRecoverableOutput(out *Output) {
	if t.Outputs == nil {
		t.Outputs = &Outputs{}
	}
	t.Outputs.Recoverable = out
}

// RecoveryTransaction returns the recovery transaction for the locked funds
func (t *Trade) RecoveryTransaction(amount uint64) (*types.Tx, error) {
	r := types.NewTx()
	fmt.Printf(">>> %#v\n", t.Outputs)
	r.AddOutput(amount, script.P2PKHHash(t.Own.RecoveryKey.Public().Hash160()))
	if err := r.AddInput(t.Outputs.Recoverable.TxID, t.Outputs.Recoverable.N, t.Own.LockScript); err != nil {
		return nil, err
	}
	lst, err := t.Trader.LockScriptTime()
	if err != nil {
		return nil, err
	}
	r.Tx().LockTime = uint32(lst.UTC().Unix())
	r.Tx().TxIn[0].Sequence = wire.MaxTxInSequenceNum - 1
	sig, err := r.InputSignature(0, 1, t.Own.RecoveryKey)
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

var (
	errNilPointer        = errors.New("nil pointer")
	errNotAStructPointer = errors.New("not a struct pointer")
)

// copy fields by name
func copyFieldsByName(src, dst interface{}) error {
	if src == nil || dst == nil {
		return errNilPointer
	}
	vs := reflect.ValueOf(src)
	vd := reflect.ValueOf(dst)
	if vs.Kind() != reflect.Ptr || vd.Kind() != reflect.Ptr {
		return errNotAStructPointer
	}
	if vs.IsNil() || vd.IsNil() {
		return errNilPointer
	}
	vs = vs.Elem()
	vd = vd.Elem()
	if vs.Kind() != reflect.Struct || vd.Kind() != reflect.Struct {
		return errNotAStructPointer
	}
	vst := vs.Type()
	vdt := vd.Type()
	for i := 0; i < vst.NumField(); i++ {
		sfld := vst.Field(i)
		dfld, ok := vdt.FieldByName(sfld.Name)
		if !ok {
			continue
		}
		if sfld.Type != dfld.Type {
			continue
		}
		vd.FieldByName(sfld.Name).Set(vs.Field(i))
	}
	return nil
}

// MarshalYAML implements yaml.Marshaler
func (t *Trade) MarshalYAML() (interface{}, error) {
	r := &tradeData{}
	if err := copyFieldsByName(t, r); err != nil {
		return nil, err
	}
	return r, nil
}

// UnmarshalYAML implements yaml.Unmarshaler
func (t *Trade) UnmarshalYAML(unmarshal func(interface{}) error) error {
	td := &tradeData{}
	if err := unmarshal(&td); err != nil {
		return err
	}
	if err := copyFieldsByName(td, t); err != nil {
		return err
	}
	return nil
}
