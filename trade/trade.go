package trade

import (
	"bytes"
	"crypto/rand"
	"errors"
	"time"

	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/duration"
	"transmutate.io/pkg/atomicswap/hash"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/params"
	"transmutate.io/pkg/atomicswap/roles"
	"transmutate.io/pkg/atomicswap/script"
	"transmutate.io/pkg/atomicswap/stages"
	"transmutate.io/pkg/atomicswap/tx"
	"transmutate.io/pkg/cryptocore/types"
	"transmutate.io/pkg/reflection"
)

type (
	Trade struct {
		Role             roles.Role        `yaml:"role"`
		Duration         duration.Duration `yaml:"duration,omitempty"`
		Token            types.Bytes       `yaml:"token,omitempty"`
		TokenHash        types.Bytes       `yaml:"token_hash,omitempty"`
		OwnInfo          *TraderInfo       `yaml:"own,omitempty"`
		TraderInfo       *TraderInfo       `yaml:"trader,omitempty"`
		RedeemKey        key.Private       `yaml:"redeem_key,omitempty"`
		RecoveryKey      key.Private       `yaml:"recover_key,omitempty"`
		RedeemableFunds  FundsData         `yaml:"redeemable_funds,omitempty"`
		RecoverableFunds FundsData         `yaml:"recoverable_funds,omitempty"`
		Stages           *stages.Stager    `yaml:"stages,omitempty"`
	}

	TraderInfo struct {
		Crypto *cryptos.Crypto `yaml:"crypto"`
		Amount types.Amount    `yaml:"amount"`
	}

	FundsData interface {
		AddFunds(funds interface{})
		Funds() interface{}
		SetLock(lock Lock)
		Lock() Lock
	}

	Lock interface {
		Bytes() types.Bytes
		LockData() (*LockData, error)
		Address(crypto *cryptos.Crypto, chain params.Chain) (string, error)
	}

	LockData struct {
		Locktime        time.Time
		TokenHash       []byte
		RedeemKeyData   key.KeyData
		RecoveryKeyData key.KeyData
	}
)

// type (
// 	Buy  interface{}
// 	Sell interface{}
// )

// NewBuy returns a new buyer trade
func NewBuy(
	ownAmount types.Amount,
	ownCrypto *cryptos.Crypto,
	traderAmount types.Amount,
	traderCrypto *cryptos.Crypto,
	dur time.Duration,
) *Trade {
	return &Trade{
		Role:     roles.Buyer,
		Duration: duration.Duration(dur),
		Stages:   stages.NewStager(tradeStages[roles.Buyer]...),
		OwnInfo: &TraderInfo{
			Amount: ownAmount,
			Crypto: ownCrypto,
		},
		TraderInfo: &TraderInfo{
			Amount: traderAmount,
			Crypto: traderCrypto,
		},
		RecoverableFunds: newFundsData(ownCrypto),
		RedeemableFunds:  newFundsData(traderCrypto),
	}
}

// NewSell returns a new seller trade
func NewSell() *Trade {
	return &Trade{
		Role:   roles.Seller,
		Stages: stages.NewStager(tradeStages[roles.Seller]...),
	}
}

// UnmarshalYAML implements yaml.Unmarshaler
func (t *Trade) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// find which cryptos first
	tc := &struct {
		OwnInfo    *TraderInfo `yaml:"own,omitempty"`
		TraderInfo *TraderInfo `yaml:"trader,omitempty"`
	}{}
	if err := unmarshal(tc); err != nil {
		return err
	}
	var td reflection.Any
	if tc.TraderInfo != nil {
		// replace types in a &Trade{}
		redeemKey, err := key.NewPrivate(tc.TraderInfo.Crypto)
		if err != nil {
			return err
		}
		recoveryKey, err := key.NewPrivate(tc.OwnInfo.Crypto)
		if err != nil {
			return err
		}
		td = reflection.MustReplaceTypeFields(&Trade{}, reflection.FieldReplacementMap{
			"RedeemKey":        interface{}(redeemKey),
			"RecoveryKey":      interface{}(recoveryKey),
			"RedeemableFunds":  newFundsData(tc.TraderInfo.Crypto),
			"RecoverableFunds": newFundsData(tc.OwnInfo.Crypto),
		})
	} else {
		td = reflection.NewStructBuilder().
			WithField("Role", roles.Role(0), `yaml:"role,omitempty"`).
			WithField("Stages", &stages.Stager{}, `yaml:"stages,omitempty"`).
			BuildPointer()
	}
	// unmarshal
	if err := unmarshal(td); err != nil {
		return err
	}
	// copy fields
	if err := reflection.CopyFields(td, t); err != nil {
		return err
	}
	return nil
}

// SetToken sets the token
func (t *Trade) SetToken(token types.Bytes) {
	// set token
	t.Token = token
	// set token hash
	t.TokenHash = hash.Hash160(token)
}

// ErrNotEnoughBytes is returned the is not possible to read enough random bytes
var ErrNotEnoughBytes = errors.New("not enough bytes")

// read random bytes
func readRandom(n int) ([]byte, error) {
	r := make([]byte, n)
	if sz, err := rand.Read(r); err != nil {
		return nil, err
	} else if sz != len(r) {
		return nil, ErrNotEnoughBytes
	}
	return r, nil
}

// token size
const tokenSize = 32

// read random token
func readRandomToken() ([]byte, error) { return readRandom(tokenSize) }

// GenerateToken generates and sets the token
func (t *Trade) GenerateToken() (types.Bytes, error) {
	// generate token
	rt, err := readRandomToken()
	if err != nil {
		return nil, err
	}
	// save token
	t.SetToken(rt)
	return rt, nil
}

// GenerateKeys generates and sets the redeem and recovery keys
func (t *Trade) GenerateKeys() error {
	var err error
	// generate recovery key
	if t.RecoveryKey, err = key.NewPrivate(t.OwnInfo.Crypto); err != nil {
		return err
	}
	// generate redeem key
	t.RedeemKey, err = key.NewPrivate(t.TraderInfo.Crypto)
	return err
}

var (
	ErrNotABuyerTrade  = errors.New("not a buyer trade")
	ErrNotASellerTrade = errors.New("not a seller trade")
)

// GenerateBuyProposal returns a buy proposal from the values set
func (t *Trade) GenerateBuyProposal() (*BuyProposal, error) {
	// only a buyer can generate a proposal
	if t.Role != roles.Buyer {
		return nil, ErrNotABuyerTrade
	}
	return &BuyProposal{
		Buyer: &BuyProposalInfo{
			Crypto:       t.OwnInfo.Crypto,
			Amount:       t.OwnInfo.Amount,
			LockDuration: t.Duration,
		},
		Seller: &BuyProposalInfo{
			Crypto:       t.TraderInfo.Crypto,
			Amount:       t.TraderInfo.Amount,
			LockDuration: t.Duration / 2,
		},
		RecoveryKeyData: t.RecoveryKey.Public().KeyData(),
		RedeemKeyData:   t.RedeemKey.Public().KeyData(),
		TokenHash:       t.TokenHash,
	}, nil
}

// AcceptBuyProposal accepts a buy proposal
func (t *Trade) AcceptBuyProposal(prop *BuyProposal) error {
	// only the seller can accept a trade
	if t.Role != roles.Seller {
		return ErrNotASellerTrade
	}
	// set token hash
	t.TokenHash = prop.TokenHash
	// own info
	t.OwnInfo = &TraderInfo{
		Amount: prop.Seller.Amount,
		Crypto: prop.Seller.Crypto,
	}
	// trader info
	t.TraderInfo = &TraderInfo{
		Amount: prop.Buyer.Amount,
		Crypto: prop.Buyer.Crypto,
	}
	// generate keys
	if err := t.GenerateKeys(); err != nil {
		return err
	}
	// now
	timeNow := time.Now().UTC()
	// generate buyer lock
	lock, err := generateTimeLock(
		prop.Buyer.Crypto,
		timeNow.Add(time.Duration(prop.Buyer.LockDuration)),
		prop.TokenHash,
		t.RedeemKey.Public().KeyData(),
		prop.RecoveryKeyData,
	)
	if err != nil {
		return err
	}
	t.RedeemableFunds = newFundsData(prop.Buyer.Crypto)
	t.RedeemableFunds.SetLock(lock)
	// generate seller lock
	lock, err = generateTimeLock(
		prop.Seller.Crypto,
		timeNow.Add(time.Duration(prop.Seller.LockDuration)),
		prop.TokenHash,
		prop.RedeemKeyData,
		t.RecoveryKey.Public().KeyData(),
	)
	if err != nil {
		return err
	}
	t.RecoverableFunds = newFundsData(prop.Seller.Crypto)
	t.RecoverableFunds.SetLock(lock)
	return nil
}

// generates a lock
func generateTimeLock(c *cryptos.Crypto, lockTime time.Time, tokenHash []byte, redeem, recovery key.KeyData) (Lock, error) {
	switch c.Type {
	case cryptos.UTXO:
		eng, err := script.NewEngine(c)
		if err != nil {
			return nil, err
		}
		lock, err := eng.HTLC(
			eng.LockTimeBytes(lockTime.UTC().Unix()),
			tokenHash,
			eng.P2PKHHashBytes(recovery),
			eng.P2PKHHashBytes(redeem),
		).Validate()
		return fundsUTXOLock(lock), nil
	case cryptos.StateBased:
		panic("not implemented")
	default:
		panic("not implemented")
	}
}

var (
	ErrMismatchTokenHash  = errors.New("mismatching token hash")
	ErrInvalidLockInteval = errors.New("invalid lock interval")
	ErrMismatchKeyData    = errors.New("mismatching key data")
)

// SetLocks sets the buyer and seller locks
func (t *Trade) SetLocks(locks *BuyProposalResponse) error {
	bd, err := locks.Buyer.LockData()
	if err != nil {
		return err
	}
	sd, err := locks.Seller.LockData()
	if err != nil {
		return err
	}
	if bd.Locktime.Sub(sd.Locktime) != time.Duration(t.Duration)/2 {
		return ErrInvalidLockInteval
	}
	if !bytes.Equal(bd.TokenHash, sd.TokenHash) || !bytes.Equal(bd.TokenHash, t.TokenHash) {
		return ErrMismatchTokenHash
	}
	if !bytes.Equal(bd.RecoveryKeyData, t.RecoveryKey.Public().KeyData()) {
		return ErrMismatchKeyData
	}
	if !bytes.Equal(sd.RedeemKeyData, t.RedeemKey.Public().KeyData()) {
		return ErrMismatchKeyData
	}
	t.RecoverableFunds.SetLock(locks.Buyer)
	t.RedeemableFunds.SetLock(locks.Seller)
	return nil
}

func (t *Trade) newRedeemTxUTXO(dest key.KeyData, fee uint64) (tx.Tx, error) {
	r, err := tx.New(t.TraderInfo.Crypto)
	if err != nil {
		return nil, err
	}
	tx := r.TxUTXO()
	amount := uint64(0)
	redeemableOutputs := t.RedeemableFunds.Funds().([]*Output)
	for _, i := range redeemableOutputs {
		amount += i.Amount
		if err = tx.AddInput(i.TxID, i.N, t.RedeemableFunds.Lock().Bytes(), i.Amount); err != nil {
			return nil, err
		}
	}
	eng, err := script.NewEngine(t.TraderInfo.Crypto)
	if err != nil {
		return nil, err
	}
	b, err := eng.P2PKHHash(dest).Validate()
	if err != nil {
		return nil, err
	}
	tx.AddOutput(amount-fee, b)
	for i := range redeemableOutputs {
		sig, err := tx.InputSignature(i, 1, t.RedeemKey)
		if err != nil {
			return nil, err
		}
		b, err = eng.
			Reset().
			HTLCRedeem(
				sig,
				t.RedeemKey.Public().SerializeCompressed(),
				t.Token,
				t.RedeemableFunds.Lock().Bytes(),
			).
			Validate()
		if err != nil {
			return nil, err
		}
		tx.SetInputSignatureScript(i, b)
	}
	return r, nil
}

func (t *Trade) newRedeemTx(dest key.KeyData, fee uint64) (tx.Tx, error) {
	switch t.TraderInfo.Crypto.Type {
	case cryptos.UTXO:
		return t.newRedeemTxUTXO(dest, fee)
	case cryptos.StateBased:
		panic("not implemented")
	default:
		panic("not implemented")
	}
}

// RedeemTxFixedFee returns the redeem transaction for the locked funds with a fixed fee
func (t *Trade) RedeemTxFixedFee(dest key.KeyData, fee uint64) (tx.Tx, error) {
	return t.newRedeemTx(dest, fee)
}

// RedeemTx returns the redeem transaction for the locked funds
func (t *Trade) RedeemTx(dest key.KeyData, feePerByte uint64) (tx.Tx, error) {
	tx, err := t.newRedeemTx(dest, 0)
	if err != nil {
		return nil, err
	}
	return t.newRedeemTx(dest, feePerByte*tx.SerializedSize())
}

func (t *Trade) newRecoveryTxUTXO(dest key.KeyData, fee uint64) (tx.Tx, error) {
	r, err := tx.New(t.OwnInfo.Crypto)
	if err != nil {
		return nil, err
	}
	tx := r.TxUTXO()
	amount := uint64(0)
	outputs := t.RecoverableFunds.Funds().([]*Output)
	for ni, i := range outputs {
		amount += i.Amount
		if err := tx.AddInput(i.TxID, i.N, t.RecoverableFunds.Lock().Bytes(), i.Amount); err != nil {
			return nil, err
		}
		tx.SetInputSequenceNumber(ni, 0xfffffffe)
	}
	eng, err := script.NewEngine(t.OwnInfo.Crypto)
	if err != nil {
		return nil, err
	}
	b, err := eng.P2PKHHash(dest).Validate()
	if err != nil {
		return nil, err
	}
	tx.AddOutput(amount-fee, b)
	lst, err := t.RecoverableFunds.Lock().LockData()
	if err != nil {
		return nil, err
	}
	tx.SetLockTime(lst.Locktime.UTC())
	for i := range outputs {
		sig, err := tx.InputSignature(i, 1, t.RecoveryKey)
		if err != nil {
			return nil, err
		}
		b, err := eng.Reset().HTLCRecover(
			sig,
			t.RecoveryKey.Public().SerializeCompressed(),
			t.RecoverableFunds.Lock().Bytes(),
		).Validate()
		if err != nil {
			return nil, err
		}
		tx.SetInputSignatureScript(i, b)
	}
	return r, nil
}

func (t *Trade) newRecoveryTx(dest key.KeyData, fee uint64) (tx.Tx, error) {
	switch txType := t.OwnInfo.Crypto.Type; txType {
	case cryptos.UTXO:
		return t.newRecoveryTxUTXO(dest, fee)
	case cryptos.StateBased:
		panic("not implemented")
	default:
		panic("not implemented")
	}
}

// RecoveryTxFixedFee returns the recovery transaction for the locked funds with a fixed fee
func (t *Trade) RecoveryTxFixedFee(dest key.KeyData, fee uint64) (tx.Tx, error) {
	return t.newRecoveryTx(dest, fee)
}

// RecoveryTx returns the recovery transaction for the locked funds
func (t *Trade) RecoveryTx(dest key.KeyData, feePerByte uint64) (tx.Tx, error) {
	tx, err := t.newRecoveryTx(dest, 0)
	if err != nil {
		return nil, err
	}
	return t.newRecoveryTx(dest, tx.SerializedSize()*feePerByte)
}
