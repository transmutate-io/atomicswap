package trade

import (
	"bytes"
	"crypto/rand"
	"errors"
	"time"

	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/duration"
	"github.com/transmutate-io/atomicswap/hash"
	"github.com/transmutate-io/atomicswap/key"
	"github.com/transmutate-io/atomicswap/roles"
	"github.com/transmutate-io/atomicswap/script"
	"github.com/transmutate-io/atomicswap/tx"
	"github.com/transmutate-io/cryptocore/types"
	"github.com/transmutate-io/reflection"
)

type (
	// TraderInfo represents a trader information
	TraderInfo struct {
		Crypto *cryptos.Crypto `yaml:"crypto"`
		Amount types.Amount    `yaml:"amount"`
	}

	// BuyerTrade represents a buyer trade
	BuyerTrade interface {
		// GenerateToken generates a new token
		GenerateToken() (types.Bytes, error)
		// GenerateBuyProposal generates a buy proposal
		GenerateBuyProposal() (*BuyProposal, error)
		// SetLocks sets the locks for the trade
		SetLocks(locks *Locks) error
	}

	// SellerTrade represents a seller trade
	SellerTrade interface {
		// AcceptBuyProposal accepts a buy proposal
		AcceptBuyProposal(prop *BuyProposal) error
		// Locks returns the locks for the trade
		Locks() *Locks
	}

	// Trade represents a trade
	Trade interface {
		// Role returns the user role in the trade
		Role() roles.Role
		// Duration returns the trade duration
		Duration() duration.Duration
		// Token returns the token
		Token() types.Bytes
		// TokenHash returns the token hash
		TokenHash() types.Bytes
		// OwnInfo returns the trader info for the user
		OwnInfo() *TraderInfo
		// TraderInfo returns the trader info for the trader
		TraderInfo() *TraderInfo
		// RedeemKey returns the redeem key
		RedeemKey() key.Private
		// RecoveryKey returns the recovery key
		RecoveryKey() key.Private
		// RedeemableFunds returns the FundsData for the redeemable funds
		RedeemableFunds() FundsData
		// RecoverableFunds returns the FundsData for the recoverable funds
		RecoverableFunds() FundsData
		// GenerateKeys generates both redeem and recovery keys
		GenerateKeys() error
		// SetToken sets the trade token (and token hash)
		SetToken(token types.Bytes)
		// RedeemTxFixedFee generates a redeem transaction with fixed fee
		RedeemTxFixedFee(lockScript []byte, fee uint64) (tx.Tx, error)
		// RedeemTx generates a redeem transaction with fee per byte
		RedeemTx(lockScript []byte, feePerByte uint64) (tx.Tx, error)
		// RecoveryTxFixedFee generates a recovery transaction with fixed fee
		RecoveryTxFixedFee(lockScript []byte, fee uint64) (tx.Tx, error)
		// RecoveryTx generates a recovery transaction with fee per byte
		RecoveryTx(lockScript []byte, feePerByte uint64) (tx.Tx, error)
		// Buyer returns a buyer trade
		Buyer() (BuyerTrade, error)
		// Seller returns a seller trade
		Seller() (SellerTrade, error)
	}
)

type baseTrade struct {
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
}

func newBuyerBaseTrade(dur time.Duration, ownAmount types.Amount, ownCrypto *cryptos.Crypto, traderAmount types.Amount, traderCrypto *cryptos.Crypto) (*baseTrade, error) {
	ownFundsData, err := newFundsData(ownCrypto)
	if err != nil {
		return nil, err
	}
	traderFundData, err := newFundsData(traderCrypto)
	if err != nil {
		return nil, err
	}
	r := &baseTrade{
		Role:     roles.Buyer,
		Duration: duration.Duration(dur),
		OwnInfo: &TraderInfo{
			Amount: ownAmount,
			Crypto: ownCrypto,
		},
		TraderInfo: &TraderInfo{
			Amount: traderAmount,
			Crypto: traderCrypto,
		},
		RecoverableFunds: ownFundsData,
		RedeemableFunds:  traderFundData,
	}
	if err = r.GenerateKeys(); err != nil {
		return nil, err
	}
	if _, err = r.GenerateToken(); err != nil {
		return nil, err
	}
	return r, nil
}

// UnmarshalYAML implements yaml.Unmarshaler
func (bt *baseTrade) UnmarshalYAML(unmarshal func(interface{}) error) error {
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
		ownFundsdata, err := newFundsData(tc.OwnInfo.Crypto)
		if err != nil {
			return err
		}
		traderFundsData, err := newFundsData(tc.TraderInfo.Crypto)
		if err != nil {
			return err
		}
		td = reflection.MustReplaceTypeFields(&baseTrade{}, reflection.FieldReplacementMap{
			"RedeemKey":        interface{}(redeemKey),
			"RecoveryKey":      interface{}(recoveryKey),
			"RedeemableFunds":  traderFundsData,
			"RecoverableFunds": ownFundsdata,
		})
	} else {
		td = reflection.NewStructBuilder().
			WithField("Role", roles.Role(0), `yaml:"role,omitempty"`).
			BuildPointer()
	}
	// unmarshal
	if err := unmarshal(td); err != nil {
		return err
	}
	// copy fields
	if err := reflection.CopyFields(td, bt); err != nil {
		return err
	}
	return nil
}

// TokenHash returns the hash for the given token
func TokenHash(t []byte) []byte { return hash.Ripemd160Sum(hash.Sha256Sum(t)) }

// SetToken sets the token
func (bt *baseTrade) SetToken(token types.Bytes) {
	// set token
	bt.Token = token
	// set token hash
	bt.TokenHash = TokenHash(token)
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

// GenerateKeys implement BuyerTrade
func (bt *baseTrade) GenerateToken() (types.Bytes, error) {
	// generate token
	rt, err := readRandomToken()
	if err != nil {
		return nil, err
	}
	// save token
	bt.SetToken(rt)
	return rt, nil
}

// GenerateKeys implement Trade
func (bt *baseTrade) GenerateKeys() error {
	var err error
	// generate recovery key
	if bt.RecoveryKey, err = key.NewPrivate(bt.OwnInfo.Crypto); err != nil {
		return err
	}
	// generate redeem key
	bt.RedeemKey, err = key.NewPrivate(bt.TraderInfo.Crypto)
	return err
}

var (
	// ErrNotABuyerTrade is returned if the trade is not a buy
	ErrNotABuyerTrade = errors.New("not a buyer trade")

	// ErrNotASellerTrade is returned if the trade is not a sell
	ErrNotASellerTrade = errors.New("not a seller trade")
)

// GenerateBuyProposal implement BuyerTrade
func (bt *baseTrade) GenerateBuyProposal() (*BuyProposal, error) {
	// only a buyer can generate a proposal
	return &BuyProposal{
		Buyer: &BuyProposalInfo{
			Crypto:       bt.OwnInfo.Crypto,
			Amount:       bt.OwnInfo.Amount,
			LockDuration: bt.Duration,
		},
		Seller: &BuyProposalInfo{
			Crypto:       bt.TraderInfo.Crypto,
			Amount:       bt.TraderInfo.Amount,
			LockDuration: bt.Duration / 2,
		},
		RecoveryKeyData: bt.RecoveryKey.Public().KeyData(),
		RedeemKeyData:   bt.RedeemKey.Public().KeyData(),
		TokenHash:       bt.TokenHash,
	}, nil
}

// generates a lock
func generateTimeLock(c *cryptos.Crypto, lockTime time.Time, tokenHash []byte, redeem, recovery key.KeyData) (Lock, error) {
	gen, err := script.NewGenerator(c)
	if err != nil {
		return nil, err
	}
	return newFundsLock(
		c,
		gen.HTLC(
			gen.LockTime(lockTime.UTC().Unix()),
			tokenHash,
			gen.P2PKHHash(recovery),
			gen.P2PKHHash(redeem),
		),
	)
}

// AcceptBuyProposal implement SellerTrade
func (bt *baseTrade) AcceptBuyProposal(prop *BuyProposal) error {
	// set duration
	bt.Duration = prop.Seller.LockDuration
	// set token hash
	bt.TokenHash = prop.TokenHash
	// own info
	bt.OwnInfo = &TraderInfo{
		Amount: prop.Seller.Amount,
		Crypto: prop.Seller.Crypto,
	}
	// trader info
	bt.TraderInfo = &TraderInfo{
		Amount: prop.Buyer.Amount,
		Crypto: prop.Buyer.Crypto,
	}
	// generate keys
	if err := bt.GenerateKeys(); err != nil {
		return err
	}
	// now
	timeNow := time.Now().UTC()
	// generate buyer lock
	lock, err := generateTimeLock(
		prop.Buyer.Crypto,
		timeNow.Add(time.Duration(prop.Buyer.LockDuration)),
		prop.TokenHash,
		bt.RedeemKey.Public().KeyData(),
		prop.RecoveryKeyData,
	)
	if err != nil {
		return err
	}
	if bt.RedeemableFunds, err = newFundsData(prop.Buyer.Crypto); err != nil {
		return err
	}
	bt.RedeemableFunds.SetLock(lock)
	// generate seller lock
	lock, err = generateTimeLock(
		prop.Seller.Crypto,
		timeNow.Add(time.Duration(prop.Seller.LockDuration)),
		prop.TokenHash,
		prop.RedeemKeyData,
		bt.RecoveryKey.Public().KeyData(),
	)
	if err != nil {
		return err
	}
	if bt.RecoverableFunds, err = newFundsData(prop.Seller.Crypto); err != nil {
		return err
	}
	bt.RecoverableFunds.SetLock(lock)
	return nil
}

var (
	// ErrMismatchTokenHash is returned when there is a token hash mismatch
	ErrMismatchTokenHash = errors.New("mismatching token hash")

	// ErrInvalidLockInterval is returned when the time lock interval is invalid
	ErrInvalidLockInterval = errors.New("invalid lock interval")

	// ErrMismatchKeyData is returned when there is a mismatch in the key data
	ErrMismatchKeyData = errors.New("mismatching key data")
)

// SetLocks implement BuyerTrade
func (bt *baseTrade) SetLocks(locks *Locks) error {
	bd, err := locks.Buyer.LockData()
	if err != nil {
		return err
	}
	sd, err := locks.Seller.LockData()
	if err != nil {
		return err
	}
	if bd.LockTime.Sub(sd.LockTime) != time.Duration(bt.Duration)/2 {
		return ErrInvalidLockInterval
	}
	if !bytes.Equal(bd.TokenHash, sd.TokenHash) || !bytes.Equal(bd.TokenHash, bt.TokenHash) {
		return ErrMismatchTokenHash
	}
	if !bytes.Equal(bd.RecoveryKeyData, bt.RecoveryKey.Public().KeyData()) {
		return ErrMismatchKeyData
	}
	if !bytes.Equal(sd.RedeemKeyData, bt.RedeemKey.Public().KeyData()) {
		return ErrMismatchKeyData
	}
	bt.RecoverableFunds.SetLock(locks.Buyer)
	bt.RedeemableFunds.SetLock(locks.Seller)
	return nil
}

// Locks implement SellerTrade
func (bt *baseTrade) Locks() *Locks {
	return &Locks{
		Buyer:  bt.RedeemableFunds.Lock(),
		Seller: bt.RecoverableFunds.Lock(),
	}
}

// ErrNotUTXO is returned in the case the crypto is not a utxo crypto
var ErrNotUTXO = errors.New("not a utxo crypto")

func (bt *baseTrade) newRedeemTxUTXO(lockScript []byte, fee uint64) (tx.Tx, error) {
	r, err := tx.New(bt.TraderInfo.Crypto)
	if err != nil {
		return nil, err
	}
	tx, ok := r.TxUTXO()
	if !ok {
		return nil, ErrNotUTXO
	}
	amount := uint64(0)
	redeemableOutputs := bt.RedeemableFunds.Funds().([]*Output)
	for _, i := range redeemableOutputs {
		amount += i.Amount
		if err = tx.AddInput(i.TxID, i.N, bt.RedeemableFunds.Lock().Bytes(), i.Amount); err != nil {
			return nil, err
		}
	}
	gen, err := script.NewGenerator(bt.TraderInfo.Crypto)
	if err != nil {
		return nil, err
	}
	tx.AddOutput(amount-fee, lockScript)
	for i := range redeemableOutputs {
		sig, err := tx.InputSignature(i, 1, bt.RedeemKey)
		if err != nil {
			return nil, err
		}
		tx.SetInputSignatureScript(i,
			gen.HTLCRedeem(
				sig,
				bt.RedeemKey.Public().SerializeCompressed(),
				bt.Token,
				bt.RedeemableFunds.Lock().Bytes(),
			),
		)
	}
	return r, nil
}

func (bt *baseTrade) newRedeemTx(lockScript []byte, fee uint64) (tx.Tx, error) {
	switch bt.TraderInfo.Crypto.Type {
	case cryptos.UTXO:
		return bt.newRedeemTxUTXO(lockScript, fee)
	case cryptos.StateBased:
		panic("not implemented")
	default:
		panic("not implemented")
	}
}

// RedeemTxFixedFee implement Trade
func (bt *baseTrade) RedeemTxFixedFee(lockScript []byte, fee uint64) (tx.Tx, error) {
	return bt.newRedeemTx(lockScript, fee)
}

// RedeemTx implement Trade
func (bt *baseTrade) RedeemTx(lockScript []byte, feePerByte uint64) (tx.Tx, error) {
	tx, err := bt.newRedeemTx(lockScript, 0)
	if err != nil {
		return nil, err
	}
	return bt.newRedeemTx(lockScript, feePerByte*tx.SerializedSize())
}

func (bt *baseTrade) newRecoveryTxUTXO(lockScript []byte, fee uint64) (tx.Tx, error) {
	r, err := tx.New(bt.OwnInfo.Crypto)
	if err != nil {
		return nil, err
	}
	tx, ok := r.TxUTXO()
	if !ok {
		return nil, ErrNotUTXO
	}
	amount := uint64(0)
	outputs := bt.RecoverableFunds.Funds().([]*Output)
	for ni, i := range outputs {
		amount += i.Amount
		if err := tx.AddInput(i.TxID, i.N, bt.RecoverableFunds.Lock().Bytes(), i.Amount); err != nil {
			return nil, err
		}
		tx.SetInputSequenceNumber(ni, 0xfffffffe)
	}
	gen, err := script.NewGenerator(bt.OwnInfo.Crypto)
	if err != nil {
		return nil, err
	}
	tx.AddOutput(amount-fee, lockScript)
	lst, err := bt.RecoverableFunds.Lock().LockData()
	if err != nil {
		return nil, err
	}
	tx.SetLockTime(lst.LockTime.UTC())
	for i := range outputs {
		sig, err := tx.InputSignature(i, 1, bt.RecoveryKey)
		if err != nil {
			return nil, err
		}
		tx.SetInputSignatureScript(i,
			gen.HTLCRecover(
				sig,
				bt.RecoveryKey.Public().SerializeCompressed(),
				bt.RecoverableFunds.Lock().Bytes(),
			),
		)
	}
	return r, nil
}

func (bt *baseTrade) newRecoveryTx(lockScript []byte, fee uint64) (tx.Tx, error) {
	switch txType := bt.OwnInfo.Crypto.Type; txType {
	case cryptos.UTXO:
		return bt.newRecoveryTxUTXO(lockScript, fee)
	case cryptos.StateBased:
		panic("not implemented")
	default:
		panic("not implemented")
	}
}

// RecoveryTxFixedFee implement Trade
func (bt *baseTrade) RecoveryTxFixedFee(lockScript []byte, fee uint64) (tx.Tx, error) {
	return bt.newRecoveryTx(lockScript, fee)
}

// RecoveryTx implement Trade
func (bt *baseTrade) RecoveryTx(lockScript []byte, feePerByte uint64) (tx.Tx, error) {
	tx, err := bt.newRecoveryTx(lockScript, 0)
	if err != nil {
		return nil, err
	}
	return bt.newRecoveryTx(lockScript, tx.SerializedSize()*feePerByte)
}
