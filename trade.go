package atomicswap

import (
	"crypto/rand"
	"errors"
	"time"

	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/hash"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/roles"
	"transmutate.io/pkg/atomicswap/stages"
	"transmutate.io/pkg/cryptocore/types"
	"transmutate.io/pkg/reflection"
)

type TraderInfo struct {
	Crypto *cryptos.Crypto `yaml:"crypto"`
	Amount types.Amount    `yaml:"amount"`
}

type Trade struct {
	Role             roles.Role     `yaml:"role"`
	Duration         Duration       `yaml:"duration"`
	Token            types.Bytes    `yaml:"token,omitempty"`
	TokenHash        types.Bytes    `yaml:"token_hash,omitempty"`
	OwnInfo          *TraderInfo    `yaml:"own,omitempty"`
	TraderInfo       *TraderInfo    `yaml:"trader,omitempty"`
	RedeemKey        key.Private    `yaml:"redeem_key,omitempty"`
	RecoveryKey      key.Private    `yaml:"recover_key,omitempty"`
	RedeemableFunds  Funds          `yaml:"redeemable_funds,omitempty"`
	RecoverableFunds Funds          `yaml:"recoverable_funds,omitempty"`
	Stages           *stages.Stager `yaml:"stages,omitempty"`
}

func NewTrade() *Trade { return &Trade{} }

func (t *Trade) WithOwnAmountCrypto(amt types.Amount, c *cryptos.Crypto) *Trade {
	t.OwnInfo = &TraderInfo{Crypto: c, Amount: amt}
	t.RecoverableFunds = newFunds(c)
	return t
}

func (t *Trade) WithTraderAmountCrypto(amt types.Amount, c *cryptos.Crypto) *Trade {
	t.TraderInfo = &TraderInfo{Crypto: c, Amount: amt}
	t.RedeemableFunds = newFunds(c)
	return t
}

func (t *Trade) WithStages(s ...stages.Stage) *Trade {
	t.Stages = stages.NewStager(s...)
	return t
}

func (t *Trade) WithRole(r roles.Role) *Trade { t.Role = r; return t }

func (t *Trade) WithDuration(d time.Duration) *Trade { t.Duration = Duration(d); return t }

func (t *Trade) WithToken(tk []byte) *Trade { t.SetToken(tk); return t }

func (t *Trade) WithTokenHash(th []byte) *Trade { t.TokenHash = th; return t }

// UnmarshalYAML implements yaml.Unmarshaler
func (t *Trade) UnmarshalYAML(unmarshal func(interface{}) error) error {
	tModel := &Trade{}
	tc := &struct {
		OwnInfo    *TraderInfo `yaml:"own,omitempty"`
		TraderInfo *TraderInfo `yaml:"trader,omitempty"`
	}{}
	if err := unmarshal(tc); err != nil {
		return err
	}
	redeemKey, err := key.NewPrivate(tc.TraderInfo.Crypto)
	if err != nil {
		return err
	}
	recoveryKey, err := key.NewPrivate(tc.OwnInfo.Crypto)
	if err != nil {
		return err
	}
	td := reflection.MustReplaceTypeFields(tModel, reflection.FieldReplacementMap{
		"RedeemKey":        interface{}(redeemKey),
		"RecoveryKey":      interface{}(recoveryKey),
		"RedeemableFunds":  newFunds(tc.TraderInfo.Crypto),
		"RecoverableFunds": newFunds(tc.OwnInfo.Crypto),
	})
	if err := unmarshal(td); err != nil {
		return err
	}
	if err := reflection.CopyFields(td, t); err != nil {
		return err
	}
	return nil
}

// SetToken sets the token
func (t *Trade) SetToken(token types.Bytes) {
	t.Token = token
	t.TokenHash = hash.Hash160(token)
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

// GenerateToken generates and sets the token
func (t *Trade) GenerateToken() (types.Bytes, error) {
	rt, err := readRandomToken()
	if err != nil {
		return nil, err
	}
	t.Token = rt
	t.TokenHash = hash.Hash160(t.Token)
	return t.Token, nil
}

func (t *Trade) GenerateKeys() error {
	var err error
	if t.RecoveryKey, err = key.NewPrivate(t.OwnInfo.Crypto); err != nil {
		return err
	}
	t.RedeemKey, err = key.NewPrivate(t.TraderInfo.Crypto)
	return err
}

type BuyProposalInfo struct {
	Crypto       *cryptos.Crypto `yaml:"crypto"`
	Amount       types.Amount    `yaml:"amount"`
	LockDuration Duration        `yaml:"lock_duration"`
}

type BuyProposal struct {
	Buyer  *BuyProposalInfo `yaml:"buyer"`
	Seller *BuyProposalInfo `yaml:"seller"`
}

type InitialBuyProposal struct {
	Buyer           *BuyProposalInfo `yaml:"buyer"`
	Seller          *BuyProposalInfo `yaml:"seller"`
	TokenHash       []byte           `yaml:"token_hash"`
	RedeemKeyData   key.KeyData      `yaml:"redeem_key_data"`
	RecoveryKeyData key.KeyData      `yaml:"recovery_key_data"`
}

var (
	ErrNotABuyerTrade  = errors.New("not a buyer trade")
	ErrNotASellerTrade = errors.New("not a seller trade")
)

// func (t *Trade) GenerateBuyProposal() (*BuyProposal, error) {
// 	if t.Role != roles.Buyer {
// 		return nil, ErrNotABuyerTrade
// 	}
// 	return &BuyProposal{
// 		Buyer: &BuyProposalInfo{
// 			Crypto:       t.OwnInfo.Crypto,
// 			Amount:       t.OwnInfo.Amount,
// 			LockDuration: t.Duration,
// 		},
// 		Seller: &BuyProposalInfo{
// 			Crypto:       t.TraderInfo.Crypto,
// 			Amount:       t.TraderInfo.Amount,
// 			LockDuration: t.Duration / 2,
// 		},
// 		RecoveryKeyData: t.RecoveryKey.Public().KeyData(),
// 		RedeemKeyData:   t.RedeemKey.Public().KeyData(),
// 		TokenHash:       t.TokenHash,
// 	}, nil
// }

type BuyProposalResponse interface {
	Accepted() bool
	Locks() (Lock, Lock, error)
	Refused() bool
	CounterProposal() (*BuyProposal, error)
}

var (
	ErrNotAccepted = errors.New("not accepted")
	ErrNotACounter = errors.New("not a counter proposal")
)

type RefusedBuyResponse struct{}

func (r RefusedBuyResponse) Accepted() bool { return false }
func (r RefusedBuyResponse) Refused() bool  { return true }

func (r RefusedBuyResponse) Locks() (Lock, Lock, error) {
	return nil, nil, ErrNotAccepted
}

func (r RefusedBuyResponse) CounterProposal() (*BuyProposal, error) {
	return nil, ErrNotACounter
}

func (r RefusedBuyResponse) MarshalYAML() (interface{}, error) {
	return false, nil
}

func (r RefusedBuyResponse) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var b bool
	return unmarshal(&b)
}

type AcceptBuyResponse struct {
	Buyer  Lock `yaml:"buyer"`
	Seller Lock `yaml:"seller"`
}

func (r *AcceptBuyResponse) Accepted() bool { return true }
func (r *AcceptBuyResponse) Refused() bool  { return false }

func (r *AcceptBuyResponse) Locks() (Lock, Lock, error) {
	return r.Buyer, r.Seller, nil
}

func (r *AcceptBuyResponse) CounterProposal() (*BuyProposal, error) {
	return nil, ErrNotACounter
}

type CounterBuyResponse BuyProposal

func (r *CounterBuyResponse) Accepted() bool { return false }
func (r *CounterBuyResponse) Refused() bool  { return false }

func (r *CounterBuyResponse) Locks() (Lock, Lock, error) {
	return nil, nil, ErrNotAccepted
}

func (r *CounterBuyResponse) CounterProposal() (*BuyProposal, error) {
	return (*BuyProposal)(r), nil
}

// type BuyProposalResponse struct {
// 	BuyerCrypto  *cryptos.Crypto `yaml:"buyer_crypto"`
// 	BuyerAmount  types.Amount    `yaml:"buyer_amount"`
// 	BuyerLock    Lock            `yaml:"buyer_lock"`
// 	SellerCrypto *cryptos.Crypto `yaml:"seller_crypto"`
// 	SellerAmount types.Amount    `yaml:"seller_amount"`
// 	SellerLock   Lock            `yaml:"seller_lock"`
// }

// func (t *Trade) AcceptBuyProposal(prop *BuyProposal) (*BuyProposalResponse, error) {
// 	if t.Role != roles.Seller {
// 		return nil, ErrNotASellerTrade
// 	}
// 	// // generate seller lock
// 	// engine, err := script.NewEngine(prop.BuyerCrypto)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// _ = engine
// 	// // generate own lock with trader field data
// 	return nil, nil
// }

// func generateLock() {}

// ------------------------------------------------------------------------------------

// type (
// 	// Output represents an output
// 	Output struct {
// 		TxID   types.Bytes `yaml:"txid,omitempty"`
// 		N      uint32        `yaml:"n"`
// 		Amount uint64        `yaml:"amount"`
// 	}

// 	// Outputs represents the outputs involved
// 	Outputs struct {
// 		Redeemable  []*Output `yaml:"redeemable,omitempty"`
// 		Recoverable []*Output `yaml:"recoverable,omitempty"`
// 	}
// )

// // TraderTradeInfo represents the trader trade info
// type TraderTradeInfo struct {
// 	Crypto          cryptos.Crypto `yaml:"crypto"`
// 	Amount          types.Amount `yaml:"amount"`
// 	LastBlockHeight uint64         `yaml:"last_block_height"`
// 	RedeemKeyHash   types.Bytes  `yaml:"recover_key_hash,omitempty"`
// 	LockScript      types.Bytes  `yaml:"lock_script,omitempty"`
// }

// func (tti *TraderTradeInfo) UnmarshalYAML(unmarshal func(interface{}) error) error {
// 	ti := &ownTradeInfo{}
// 	err := unmarshal(ti)
// 	if err != nil {
// 		return err
// 	}
// 	if err = copyFieldsByName(ti, tti); err != nil {
// 		return err
// 	}
// 	tti.Crypto, err = cryptos.ParseCrypto(ti.Crypto)
// 	return err
// }

// type traderTradeInfo struct {
// 	Crypto          string         `yaml:"crypto"`
// 	Amount          types.Amount `yaml:"amount"`
// 	LastBlockHeight uint64         `yaml:"last_block_height"`
// 	RedeemKeyHash   types.Bytes  `yaml:"recover_key_hash,omitempty"`
// 	LockScript      types.Bytes  `yaml:"lock_script,omitempty"`
// }

// // TokenHash returns the token hash if set, otherwise nil
// func (t *Trade) TokenHash() types.Bytes {
// 	if t.Token != nil {
// 		return hash.Hash160(t.Token)
// 	}
// 	return t.TokenHash
// }

// // Token returns the token if set, otherwise nil
// func (t *Trade) Token() types.Bytes { return t.Token }

// // SetTokenHash sets the token hash
// func (t *Trade) SetTokenHash(tokenHash types.Bytes) { t.TokenHash = tokenHash }

// // SetToken sets the token
// func (t *Trade) SetToken(token types.Bytes) {
// 	t.Token = token
// 	t.TokenHash = hash.Hash160(token)
// }

// var (
// 	sellerStages = map[stages.Stage]stages.Stage{
// 		stages.ReceivePublicKeyHash: stages.SharePublicKeyHash,
// 		stages.SharePublicKeyHash:   stages.ShareTokenHash,
// 		stages.ShareTokenHash:       stages.GenerateLockScript,
// 		stages.GenerateLockScript:   stages.ShareLockScript,
// 		stages.ShareLockScript:      stages.ReceiveLockScript,
// 		stages.ReceiveLockScript:    stages.LockFunds,
// 		stages.LockFunds:            stages.WaitLockTransaction,
// 		stages.WaitLockTransaction:  stages.RedeemFunds,
// 		stages.RedeemFunds:          stages.Done,
// 	}
// 	buyerStages = map[stages.Stage]stages.Stage{
// 		stages.SharePublicKeyHash:    stages.ReceivePublicKeyHash,
// 		stages.ReceivePublicKeyHash:  stages.ReceiveTokenHash,
// 		stages.ReceiveTokenHash:      stages.ReceiveLockScript,
// 		stages.ReceiveLockScript:     stages.GenerateLockScript,
// 		stages.GenerateLockScript:    stages.ShareLockScript,
// 		stages.ShareLockScript:       stages.WaitLockTransaction,
// 		stages.WaitLockTransaction:   stages.LockFunds,
// 		stages.LockFunds:             stages.WaitRedeemTransaction,
// 		stages.WaitRedeemTransaction: stages.RedeemFunds,
// 		stages.RedeemFunds:           stages.Done,
// 	}
// )

// // NextStage advance the trade to the next stage
// func (t *Trade) NextStage() stages.Stage {
// 	var stageMap map[stages.Stage]stages.Stage
// 	if t.Role == roles.Seller {
// 		stageMap = sellerStages
// 	} else {
// 		stageMap = buyerStages
// 	}
// 	t.Stage = stageMap[t.Stage]
// 	return t.Stage
// }

// func (t *Trade) generateKeys() error {
// 	var err error
// 	if t.Own.RecoveryKey, err = t.Own.Crypto.NewPrivateKey(); err != nil {
// 		return err
// 	}
// 	if t.Own.RedeemKey, err = t.Trader.Crypto.NewPrivateKey(); err != nil {
// 		return err
// 	}
// 	if t.Role == roles.Seller {
// 		if t.Token, err = readRandomToken(); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (t *Trade) generateToken() error {
// 	rt, err := readRandomToken()
// 	if err != nil {
// 		return err
// 	}
// 	t.Token = rt
// 	t.TokenHash = hash.Hash160(t.Token)
// 	return nil
// }

// // ErrNotEnoughBytes is returned the is not possible to read enough random bytes
// var ErrNotEnoughBytes = errors.New("not enough bytes")

// const tokenSize = 32

// func readRandom(n int) ([]byte, error) {
// 	r := make([]byte, n)
// 	if sz, err := rand.Read(r); err != nil {
// 		return nil, err
// 	} else if sz != len(r) {
// 		return nil, ErrNotEnoughBytes
// 	}
// 	return r, nil
// }

// func readRandomToken() ([]byte, error) { return readRandom(tokenSize) }

// // GenerateOwnLockScript generates the user own lock script
// func (t *Trade) GenerateOwnLockScript() error {
// 	var lockTime time.Time
// 	if t.Trader.LockScript == nil {
// 		lockTime = time.Now().UTC().Add(time.Duration(t.Duration))
// 	} else {
// 		lst, err := t.Trader.LockScriptTime()
// 		if err != nil {
// 			return err
// 		}
// 		lockTime = lst.Add(-(time.Duration(t.Duration) / 2))
// 	}
// 	r, err := script.Validate(script.HTLC(
// 		script.LockTimeTime(lockTime),
// 		t.TokenHash,
// 		script.P2PKHHash(t.Own.RecoveryKey.Public().Hash160()),
// 		script.P2PKHHash(t.Trader.RedeemKeyHash),
// 	))
// 	if err != nil {
// 		return err
// 	}
// 	t.Own.LockScript = r
// 	return nil
// }

// // LockScriptTime returns the lock time from the trader lock script
// func (tti *TraderTradeInfo) LockScriptTime() (time.Time, error) {
// 	lsd, err := parseLockScript(tti.LockScript)
// 	if err != nil {
// 		return time.Time{}, err
// 	}
// 	return lsd.timeLock, nil
// }

// // ErrInvalidLockScript is returns when the lock script is invalid
// var ErrInvalidLockScript = errors.New("invalid lock script")

// var expHTLC = []string{
// 	"OP_IF",
// 	"", "OP_CHECKLOCKTIMEVERIFY", "OP_DROP",
// 	"OP_DUP", "OP_HASH160", "", "OP_EQUALVERIFY", "OP_CHECKSIG",
// 	"OP_ELSE",
// 	"OP_HASH160", "", "OP_EQUALVERIFY",
// 	"OP_DUP", "OP_HASH160", "", "OP_EQUALVERIFY", "OP_CHECKSIG", "OP_ENDIF",
// }

// type lockScriptData struct {
// 	timeLock        time.Time
// 	tokenHash       []byte
// 	redeemKeyHash   []byte
// 	recoveryKeyHash []byte
// }

// func parseLockScript(ls []byte) (*lockScriptData, error) {
// 	r := &lockScriptData{}
// 	// check contract format
// 	inst, err := script.DisassembleStrings(ls)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(inst) != len(expHTLC) {
// 		return nil, ErrInvalidLockScript
// 	}
// 	for i, op := range inst {
// 		if expHTLC[i] == "" {
// 			continue
// 		}
// 		if op != expHTLC[i] {
// 			return nil, ErrInvalidLockScript
// 		}
// 	}
// 	// time lock
// 	b, err := hex.DecodeString(inst[1])
// 	if err != nil {
// 		return nil, err
// 	}
// 	n, err := script.ParseInt64(b)
// 	if err != nil {
// 		return nil, err
// 	}
// 	r.timeLock = time.Unix(n, 0)
// 	// token hash
// 	if r.TokenHash, err = hex.DecodeString(inst[11]); err != nil {
// 		return nil, err
// 	}
// 	// redeem key hash
// 	if r.redeemKeyHash, err = hex.DecodeString(inst[15]); err != nil {
// 		return nil, err
// 	}
// 	return r, nil
// }

// // CheckTraderLockScript verifies the trader lock script
// func (t *Trade) CheckTraderLockScript(tradeLockScript []byte) error {
// 	lsd, err := parseLockScript(tradeLockScript)
// 	if err != nil {
// 		return err
// 	}
// 	if t.Duration != 0 && time.Now().UTC().Add(time.Duration(t.Duration)).After(lsd.timeLock) {
// 		return ErrInvalidLockScript
// 	}
// 	if !bytes.Equal(lsd.TokenHash, t.TokenHash) {
// 		return ErrInvalidLockScript
// 	}
// 	if !bytes.Equal(lsd.redeemKeyHash, t.Own.RedeemKey.Public().Hash160()) {
// 		return ErrInvalidLockScript
// 	}
// 	return nil
// }

// func bytesJoin(b ...[]byte) []byte { return bytes.Join(b, []byte{}) }

// func (t *Trade) newRedeemTransactionUTXO(tx transaction.TxUTXO, fee uint64) error {
// 	redeemScript, err := script.Validate(bytesJoin(
// 		script.Data(t.Token),
// 		script.Int64(0),
// 		script.Data(t.Trader.LockScript),
// 	))
// 	amount := uint64(0)
// 	for _, i := range t.Outputs.Redeemable {
// 		amount += i.Amount
// 		if err = tx.AddInput(i.TxID, i.N, t.Trader.LockScript); err != nil {
// 			return err
// 		}
// 	}
// 	if err != nil {
// 		return err
// 	}
// 	tx.AddOutput(amount-fee, script.P2PKHHash(t.Own.RecoveryKey.Public().Hash160()))
// 	for i := range t.Outputs.Redeemable {
// 		sig, err := tx.InputSignature(i, 1, t.Own.RedeemKey)
// 		if err != nil {
// 			return err
// 		}
// 		tx.SetP2SHInputSignatureScript(i, bytesJoin(script.Data(sig), script.Data(t.Own.RedeemKey.Public().SerializeCompressed()), redeemScript))
// 	}
// 	return nil
// }

// func (t *Trade) newRedeemTransaction(fee uint64) (types.Tx, error) {
// 	r := t.Trader.Crypto.NewTx()
// 	switch txType := r.Type(); txType {
// 	case types.UTXO:
// 		if err := t.newRedeemTransactionUTXO(r.TxUTXO(), fee); err != nil {
// 			return nil, err
// 		}
// 	default:
// 		return nil, errors.New(fmt.Sprintf("unknown transaction type: %v", txType))
// 	}

// 	return r, nil
// }

// // RedeemTransaction returns the redeem transaction for the locked funds with a fixed fee
// func (t *Trade) RedeemTransactionFixedFee(fee uint64) (types.Tx, error) {
// 	return t.newRedeemTransaction(fee)
// }

// // RedeemTransaction returns the redeem transaction for the locked funds
// func (t *Trade) RedeemTransaction(feePerByte uint64) (types.Tx, error) {
// 	tx, err := t.newRedeemTransaction(0)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return t.newRedeemTransaction(feePerByte * tx.SerializedSize())
// }

// // AddRedeemableOutput adds a redeemable output to the trade
// func (t *Trade) AddRedeemableOutput(out *Output) {
// 	if t.Outputs == nil {
// 		t.Outputs = &Outputs{}
// 	}
// 	if t.Outputs.Redeemable == nil {
// 		t.Outputs.Redeemable = make([]*Output, 0, 4)
// 	}
// 	t.Outputs.Redeemable = append(t.Outputs.Redeemable, out)
// }

// // AddRecoverableOutput adds a recoverable output
// func (t *Trade) AddRecoverableOutput(out *Output) {
// 	if t.Outputs == nil {
// 		t.Outputs = &Outputs{}
// 	}
// 	t.Outputs.Recoverable = append(t.Outputs.Recoverable, out)
// }

// func (t *Trade) newRecoveryTransactionUTXO(tx transaction.TxUTXO, fee uint64) error {
// 	amount := uint64(0)
// 	for ni, i := range t.Outputs.Recoverable {
// 		amount += i.Amount
// 		if err := tx.AddInput(i.TxID, i.N, t.Own.LockScript); err != nil {
// 			return err
// 		}
// 		tx.SetInputSequence(ni, 0xfffffffe)
// 	}
// 	tx.AddOutput(amount-fee, script.P2PKHHash(t.Own.RecoveryKey.Public().Hash160()))
// 	lst, err := t.Trader.LockScriptTime()
// 	if err != nil {
// 		return err
// 	}
// 	tx.SetLockTime(uint32(lst.UTC().Unix()))
// 	for i := range t.Outputs.Recoverable {
// 		sig, err := tx.InputSignature(i, 1, t.Own.RecoveryKey)
// 		if err != nil {
// 			return err
// 		}
// 		err = tx.SetP2SHInputPrefixes(i, sig, t.Own.RecoveryKey.Public().SerializeCompressed(), []byte{1})
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (t *Trade) newRecoveryTransaction(fee uint64) (types.Tx, error) {
// 	r := t.Own.Crypto.NewTx()
// 	switch txType := r.Type(); txType {
// 	case types.UTXO:
// 		if err := t.newRecoveryTransactionUTXO(r.TxUTXO(), fee); err != nil {
// 			return nil, err
// 		}
// 	}
// 	return r, nil
// }

// // RecoveryTransaction returns the recovery transaction for the locked funds with a fixed fee
// func (t *Trade) RecoveryTransactionFixedFee(fee uint64) (types.Tx, error) {
// 	return t.newRecoveryTransaction(fee)
// }

// // RecoveryTransaction returns the recovery transaction for the locked funds
// func (t *Trade) RecoveryTransaction(feePerByte uint64) (types.Tx, error) {
// 	tx, err := t.newRecoveryTransaction(0)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return t.newRecoveryTransaction(tx.SerializedSize() * feePerByte)
// }

// // MarshalYAML implements yaml.Marshaler
// func (t *Trade) MarshalYAML() (interface{}, error) {
// 	r := &tradeData{}
// 	if err := copyFieldsByName(t, r); err != nil {
// 		return nil, err
// 	}
// 	return r, nil
// }

// // UnmarshalYAML implements yaml.Unmarshaler
// func (t *Trade) UnmarshalYAML(unmarshal func(interface{}) error) error {
// 	td := &tradeData{}
// 	if err := unmarshal(&td); err != nil {
// 		return err
// 	}
// 	if err := copyFieldsByName(td, t); err != nil {
// 		return err
// 	}
// 	return nil
// }

// // Trade represents an atomic swap trade
// type Trade struct {
// 	// Stage of the trade
// 	Stage stages.Stage
// 	// Role on the trade
// 	Role roles.Role
// 	// Duration represents the trade lock time
// 	Duration  types.Duration
// 	token     cctypes.Bytes
// 	tokenHash cctypes.Bytes
// 	// Outputs contains the outputs involved
// 	Outputs *Outputs
// 	// Own contains own user data and keys
// 	Own *OwnTradeInfo
// 	// Trader contrains the trader data
// 	Trader *TraderTradeInfo
// 	// OnChainDataExchange whether to exchange data between traders manually or on-chain
// 	OnChainDataExchange bool
// }

// type tradeData struct {
// 	Stage     stages.Stage     `yaml:"stage"`
// 	Role      roles.Role       `yaml:"role"`
// 	Duration  types.Duration   `yaml:"duration"`
// 	Outputs   *Outputs         `yaml:"outputs,omitempty"`
// 	Own       *OwnTradeInfo    `yaml:"own,omitempty"`
// 	Trader    *TraderTradeInfo `yaml:"trader,omitempty"`
// 	Token     cctypes.Bytes    `yaml:"token,omitempty"`
// 	TokenHash cctypes.Bytes    `yaml:"token_hash,omitempty"`
// }

// func newTrade(role roles.Role, stage stages.Stage, ownAmount cctypes.Amount, ownCrypto cryptos.Crypto, tradeAmount cctypes.Amount, tradeCrypto cryptos.Crypto) (*Trade, error) {
// 	r := &Trade{
// 		Role:  role,
// 		Stage: stage,
// 		Own: &OwnTradeInfo{
// 			Crypto:          ownCrypto,
// 			Amount:          ownAmount,
// 			LastBlockHeight: 1,
// 		},
// 		Trader: &TraderTradeInfo{
// 			Crypto:          tradeCrypto,
// 			Amount:          tradeAmount,
// 			LastBlockHeight: 1,
// 		},
// 	}
// 	if err := r.generateKeys(); err != nil {
// 		return nil, err
// 	}
// 	return r, nil
// }

// // NewBuyerTrade starts a trade as a buyer
// func NewBuyerTrade(ownAmount cctypes.Amount, ownCrypto cryptos.Crypto, tradeAmount cctypes.Amount, tradeCrypto cryptos.Crypto) (*Trade, error) {
// 	return newTrade(
// 		roles.Buyer,
// 		stages.SharePublicKeyHash,
// 		ownAmount,
// 		ownCrypto,
// 		tradeAmount,
// 		tradeCrypto,
// 	)
// }

// // NewSellerTrade starts a trade as a seller
// func NewSellerTrade(ownAmount cctypes.Amount, ownCrypto cryptos.Crypto, tradeAmount cctypes.Amount, tradeCrypto cryptos.Crypto) (*Trade, error) {
// 	r, err := newTrade(
// 		roles.Seller,
// 		stages.ReceivePublicKeyHash,
// 		ownAmount,
// 		ownCrypto,
// 		tradeAmount,
// 		tradeCrypto,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err = r.generateToken(); err != nil {
// 		return nil, err
// 	}
// 	return r, nil
// }

// 	// Outputs represents the outputs involved
// 	Outputs struct {
// 		Redeemable  []*Output `yaml:"redeemable,omitempty"`
// 		Recoverable []*Output `yaml:"recoverable,omitempty"`
// 	}
// )

// // OwnTradeInfo represents the own user trade info
// type OwnTradeInfo struct {
// 	Crypto          cryptos.Crypto `yaml:"crypto"`
// 	Amount          cctypes.Amount `yaml:"amount"`
// 	LastBlockHeight uint64         `yaml:"last_block_height"`
// 	RedeemKey       key.Private    `yaml:"redeem_key,omitempty"`
// 	RecoveryKey     key.Private    `yaml:"recover_key,omitempty"`
// 	LockScript      cctypes.Bytes  `yaml:"lock_script,omitempty"`
// }

// func (oti *OwnTradeInfo) UnmarshalYAML(unmarshal func(interface{}) error) error {
// 	ti := &ownTradeInfo{}
// 	err := unmarshal(ti)
// 	if err != nil {
// 		return err
// 	}
// 	if err = copyFieldsByName(ti, oti); err != nil {
// 		return err
// 	}
// 	oti.Crypto, err = cryptos.ParseCrypto(ti.Crypto)
// 	if err != nil {
// 		return err
// 	}
// 	// oti.Crypto
// 	return nil
// }

// type ownTradeInfo struct {
// 	Crypto          string         `yaml:"crypto"`
// 	Amount          cctypes.Amount `yaml:"amount"`
// 	LastBlockHeight uint64         `yaml:"last_block_height"`
// 	RedeemKey       string         `yaml:"redeem_key,omitempty"`
// 	RecoveryKey     string         `yaml:"recover_key,omitempty"`
// 	LockScript      cctypes.Bytes  `yaml:"lock_script,omitempty"`
// }

// // TraderTradeInfo represents the trader trade info
// type TraderTradeInfo struct {
// 	Crypto          cryptos.Crypto `yaml:"crypto"`
// 	Amount          cctypes.Amount `yaml:"amount"`
// 	LastBlockHeight uint64         `yaml:"last_block_height"`
// 	RedeemKeyHash   cctypes.Bytes  `yaml:"recover_key_hash,omitempty"`
// 	LockScript      cctypes.Bytes  `yaml:"lock_script,omitempty"`
// }

// func (tti *TraderTradeInfo) UnmarshalYAML(unmarshal func(interface{}) error) error {
// 	ti := &ownTradeInfo{}
// 	err := unmarshal(ti)
// 	if err != nil {
// 		return err
// 	}
// 	if err = copyFieldsByName(ti, tti); err != nil {
// 		return err
// 	}
// 	tti.Crypto, err = cryptos.ParseCrypto(ti.Crypto)
// 	return err
// }

// type traderTradeInfo struct {
// 	Crypto          string         `yaml:"crypto"`
// 	Amount          cctypes.Amount `yaml:"amount"`
// 	LastBlockHeight uint64         `yaml:"last_block_height"`
// 	RedeemKeyHash   cctypes.Bytes  `yaml:"recover_key_hash,omitempty"`
// 	LockScript      cctypes.Bytes  `yaml:"lock_script,omitempty"`
// }

// // GenerateOwnLockScript generates the user own lock script
// func (t *TradeBTtradeBTC) GenerateOwnLockScript() error {
// 	var lockTime time.Time
// 	if t.Trader.LockScript == nil {
// 		lockTime = time.Now().UTC().Add(time.Duration(t.Duration))
// 	} else {
// 		lst, err := t.Trader.LockScriptTime()
// 		if err != nil {
// 			return err
// 		}
// 		lockTime = lst.Add(-(time.Duration(t.Duration) / 2))
// 	}
// 	r, err := script.Validate(script.HTLC(
// 		script.LockTimeTime(lockTime),
// 		t.tokenHash,
// 		script.P2PKHHash(t.Own.RecoveryKey.Public().Hash160()),
// 		script.P2PKHHash(t.Trader.RedeemKeyHash),
// 	))
// 	if err != nil {
// 		return err
// 	}
// 	t.Own.LockScript = r
// 	return nil
// }

// // LockScriptTime returns the lock time from the trader lock script
// func (tti *TraderTradeInfo) LockScriptTime() (time.Time, error) {
// 	lsd, err := parseLockScript(tti.LockScript)
// 	if err != nil {
// 		return time.Time{}, err
// 	}
// 	return lsd.timeLock, nil
// }

// // ErrInvalidLockScript is returns when the lock script is invalid
// var ErrInvalidLockScript = errors.New("invalid lock script")

// var expHTLC = []string{
// 	"OP_IF",
// 	"", "OP_CHECKLOCKTIMEVERIFY", "OP_DROP",
// 	"OP_DUP", "OP_HASH160", "", "OP_EQUALVERIFY", "OP_CHECKSIG",
// 	"OP_ELSE",
// 	"OP_HASH160", "", "OP_EQUALVERIFY",
// 	"OP_DUP", "OP_HASH160", "", "OP_EQUALVERIFY", "OP_CHECKSIG", "OP_ENDIF",
// }

// type lockScriptData struct {
// 	timeLock        time.Time
// 	tokenHash       []byte
// 	redeemKeyHash   []byte
// 	recoveryKeyHash []byte
// }

// func parseLockScript(ls []byte) (*lockScriptData, error) {
// 	r := &lockScriptData{}
// 	// check contract format
// 	inst, err := script.DisassembleStrings(ls)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(inst) != len(expHTLC) {
// 		return nil, ErrInvalidLockScript
// 	}
// 	for i, op := range inst {
// 		if expHTLC[i] == "" {
// 			continue
// 		}
// 		if op != expHTLC[i] {
// 			return nil, ErrInvalidLockScript
// 		}
// 	}
// 	// time lock
// 	b, err := hex.DecodeString(inst[1])
// 	if err != nil {
// 		return nil, err
// 	}
// 	n, err := script.ParseInt64(b)
// 	if err != nil {
// 		return nil, err
// 	}
// 	r.timeLock = time.Unix(n, 0)
// 	// token hash
// 	if r.tokenHash, err = hex.DecodeString(inst[11]); err != nil {
// 		return nil, err
// 	}
// 	// redeem key hash
// 	if r.redeemKeyHash, err = hex.DecodeString(inst[15]); err != nil {
// 		return nil, err
// 	}
// 	return r, nil
// }

// // CheckTraderLockScript verifies the trader lock script
// func (t *TradeBTtradeBTC) CheckTraderLockScript(tradeLockScript []byte) error {
// 	lsd, err := parseLockScript(tradeLockScript)
// 	if err != nil {
// 		return err
// 	}
// 	if t.Duration != 0 && time.Now().UTC().Add(time.Duration(t.Duration)).After(lsd.timeLock) {
// 		return ErrInvalidLockScript
// 	}
// 	if !bytes.Equal(lsd.tokenHash, t.tokenHash) {
// 		return ErrInvalidLockScript
// 	}
// 	if !bytes.Equal(lsd.redeemKeyHash, t.Own.RedeemKey.Public().Hash160()) {
// 		return ErrInvalidLockScript
// 	}
// 	return nil
// }

// func bytesJoin(b ...[]byte) []byte { return bytes.Join(b, []byte{}) }

// func (t *TradeBTtradeBTC) newRedeemTransactionUTXO(tx transaction.TxUTXO, fee uint64) error {
// 	redeemScript, err := script.Validate(bytesJoin(
// 		script.Data(t.token),
// 		script.Int64(0),
// 		script.Data(t.Trader.LockScript),
// 	))
// 	amount := uint64(0)
// 	for _, i := range t.Outputs.Redeemable {
// 		amount += i.Amount
// 		if err = tx.AddInput(i.TxID, i.N, t.Trader.LockScript); err != nil {
// 			return err
// 		}
// 	}
// 	if err != nil {
// 		return err
// 	}
// 	tx.AddOutput(amount-fee, script.P2PKHHash(t.Own.RecoveryKey.Public().Hash160()))
// 	for i := range t.Outputs.Redeemable {
// 		sig, err := tx.InputSignature(i, 1, t.Own.RedeemKey)
// 		if err != nil {
// 			return err
// 		}
// 		tx.SetP2SHInputSignatureScript(i, bytesJoin(script.Data(sig), script.Data(t.Own.RedeemKey.Public().SerializeCompressed()), redeemScript))
// 	}
// 	return nil
// }

// func (t *TradeBTtradeBTC) newRedeemTransaction(fee uint64) (types.Tx, error) {
// 	r := t.Trader.Crypto.NewTx()
// 	switch txType := r.Type(); txType {
// 	case types.UTXO:
// 		if err := t.newRedeemTransactionUTXO(r.TxUTXO(), fee); err != nil {
// 			return nil, err
// 		}
// 	default:
// 		return nil, errors.New(fmt.Sprintf("unknown transaction type: %v", txType))
// 	}

// 	return r, nil
// }

// // RedeemTransaction returns the redeem transaction for the locked funds with a fixed fee
// func (t *TradeBTtradeBTC) RedeemTransactionFixedFee(fee uint64) (types.Tx, error) {
// 	return t.newRedeemTransaction(fee)
// }

// // RedeemTransaction returns the redeem transaction for the locked funds
// func (t *TradeBTtradeBTC) RedeemTransaction(feePerByte uint64) (types.Tx, error) {
// 	tx, err := t.newRedeemTransaction(0)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return t.newRedeemTransaction(feePerByte * tx.SerializedSize())
// }

// // AddRedeemableOutput adds a redeemable output to the trade
// func (t *TradeBTtradeBTC) AddRedeemableOutput(out *Output) {
// 	if t.Outputs == nil {
// 		t.Outputs = &Outputs{}
// 	}
// 	if t.Outputs.Redeemable == nil {
// 		t.Outputs.Redeemable = make([]*Output, 0, 4)
// 	}
// 	t.Outputs.Redeemable = append(t.Outputs.Redeemable, out)
// }

// // AddRecoverableOutput adds a recoverable output
// func (t *TradeBTtradeBTC) AddRecoverableOutput(out *Output) {
// 	if t.Outputs == nil {
// 		t.Outputs = &Outputs{}
// 	}
// 	t.Outputs.Recoverable = append(t.Outputs.Recoverable, out)
// }

// func (t *TradeBTtradeBTC) newRecoveryTransactionUTXO(tx transaction.TxUTXO, fee uint64) error {
// 	amount := uint64(0)
// 	for ni, i := range t.Outputs.Recoverable {
// 		amount += i.Amount
// 		if err := tx.AddInput(i.TxID, i.N, t.Own.LockScript); err != nil {
// 			return err
// 		}
// 		tx.SetInputSequence(ni, 0xfffffffe)
// 	}
// 	tx.AddOutput(amount-fee, script.P2PKHHash(t.Own.RecoveryKey.Public().Hash160()))
// 	lst, err := t.Trader.LockScriptTime()
// 	if err != nil {
// 		return err
// 	}
// 	tx.SetLockTime(uint32(lst.UTC().Unix()))
// 	for i := range t.Outputs.Recoverable {
// 		sig, err := tx.InputSignature(i, 1, t.Own.RecoveryKey)
// 		if err != nil {
// 			return err
// 		}
// 		err = tx.SetP2SHInputPrefixes(i, sig, t.Own.RecoveryKey.Public().SerializeCompressed(), []byte{1})
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (t *TradeBTtradeBTC) newRecoveryTransaction(fee uint64) (types.Tx, error) {
// 	r := t.Own.Crypto.NewTx()
// 	switch txType := r.Type(); txType {
// 	case types.UTXO:
// 		if err := t.newRecoveryTransactionUTXO(r.TxUTXO(), fee); err != nil {
// 			return nil, err
// 		}
// 	}
// 	return r, nil
// }

// // RecoveryTransaction returns the recovery transaction for the locked funds with a fixed fee
// func (t *TradeBTtradeBTC) RecoveryTransactionFixedFee(fee uint64) (types.Tx, error) {
// 	return t.newRecoveryTransaction(fee)
// }

// // RecoveryTransaction returns the recovery transaction for the locked funds
// func (t *TradeBTtradeBTC) RecoveryTransaction(feePerByte uint64) (types.Tx, error) {
// 	tx, err := t.newRecoveryTransaction(0)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return t.newRecoveryTransaction(tx.SerializedSize() * feePerByte)
// }

// var (
// 	errNilPointer        = errors.New("nil pointer")
// 	errNotAStructPointer = errors.New("not a struct pointer")
// )

// // copy fields by name
// func copyFieldsByName(src, dst interface{}) error {
// 	if src == nil || dst == nil {
// 		return errNilPointer
// 	}
// 	vs := reflect.ValueOf(src)
// 	vd := reflect.ValueOf(dst)
// 	if vs.Kind() != reflect.Ptr || vd.Kind() != reflect.Ptr {
// 		return errNotAStructPointer
// 	}
// 	if vs.IsNil() || vd.IsNil() {
// 		return errNilPointer
// 	}
// 	vs = vs.Elem()
// 	vd = vd.Elem()
// 	if vs.Kind() != reflect.Struct || vd.Kind() != reflect.Struct {
// 		return errNotAStructPointer
// 	}
// 	vst := vs.Type()
// 	vdt := vd.Type()
// 	for i := 0; i < vst.NumField(); i++ {
// 		sfld := vst.Field(i)
// 		dfld, ok := vdt.FieldByName(sfld.Name)
// 		if !ok {
// 			continue
// 		}
// 		if sfld.Type != dfld.Type {
// 			continue
// 		}
// 		vd.FieldByName(sfld.Name).Set(vs.Field(i))
// 	}
// 	return nil
// }
