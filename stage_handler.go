package atomicswap

import (
	"fmt"

	"transmutate.io/pkg/atomicswap/stages"
)

type (
	StageHandlerFunc = func(trade *Trade) error
	StageHandlerMap  = map[stages.Stage]StageHandlerFunc
	StageHandler     struct{ handlers StageHandlerMap }
)

func NewStageHandler(hm StageHandlerMap) *StageHandler {
	return &StageHandler{handlers: hm}
}

func (sh *StageHandler) InstallStageHandlers(hm StageHandlerMap) {
	for k, v := range sh.handlers {
		sh.handlers[k] = v
	}
}

func (sh *StageHandler) InstallHandler(s stages.Stage, h StageHandlerFunc) {
	sh.handlers[s] = h
}

func (sh *StageHandler) Unhandled(s ...stages.Stage) []stages.Stage {
	r := make([]stages.Stage, 0, len(s))
	for _, i := range s {
		if _, ok := sh.handlers[i]; !ok {
			r = append(r, i)
		}
	}
	return r
}

func (sh *StageHandler) Handle(s stages.Stage, t *Trade) error {
	h, ok := sh.handlers[s]
	if !ok {
		return StageNotHandlerError(s)
	}
	return h(t)
}

type StageNotHandlerError stages.Stage

func (e StageNotHandlerError) Error() string {
	return fmt.Sprintf("stage not handled: %s", stages.Stage(e).String())
}

var (
	DefaultStageHandlers = StageHandlerMap{
		stages.Done:         func(_ *Trade) error { return nil },
		stages.GenerateKeys: func(tr *Trade) error { return tr.GenerateKeys() },
		stages.GenerateToken: func(tr *Trade) error {
			_, err := tr.GenerateToken()
			return err
		},
	}
)

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
