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
	"transmutate.io/pkg/atomicswap/roles"
	"transmutate.io/pkg/atomicswap/script"
	"transmutate.io/pkg/atomicswap/stages"
	"transmutate.io/pkg/atomicswap/tx"
	"transmutate.io/pkg/cryptocore/types"
	"transmutate.io/pkg/reflection"
)

type (
	TraderInfo struct {
		Crypto *cryptos.Crypto `yaml:"crypto"`
		Amount types.Amount    `yaml:"amount"`
	}

	Trade interface {
		Role() roles.Role
		Duration() duration.Duration
		Token() types.Bytes
		TokenHash() types.Bytes
		OwnInfo() *TraderInfo
		TraderInfo() *TraderInfo
		RedeemKey() key.Private
		RecoveryKey() key.Private
		RedeemableFunds() FundsData
		RecoverableFunds() FundsData
		Stager() *stages.Stager

		GenerateToken() (types.Bytes, error)
		GenerateKeys() error
		GenerateBuyProposal() (*BuyProposal, error)
		AcceptBuyProposal(prop *BuyProposal) error
		SetLocks(locks *BuyProposalResponse) error
		SetToken(token types.Bytes)
		RedeemTxFixedFee(dest key.KeyData, fee uint64) (tx.Tx, error)
		RedeemTx(dest key.KeyData, feePerByte uint64) (tx.Tx, error)
		RecoveryTxFixedFee(dest key.KeyData, fee uint64) (tx.Tx, error)
		RecoveryTx(dest key.KeyData, feePerByte uint64) (tx.Tx, error)

		// newRedeemTxUTXO(dest key.KeyData, fee uint64) (tx.Tx, error)
		// newRedeemTx(dest key.KeyData, fee uint64) (tx.Tx, error)
		// newRecoveryTxUTXO(dest key.KeyData, fee uint64) (tx.Tx, error)
		// newRecoveryTx(dest key.KeyData, fee uint64) (tx.Tx, error)
	}
)

var _ Trade = (*OnChainTrade)(nil)

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
	Stager           *stages.Stager    `yaml:"stages,omitempty"`
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
			WithField("Stager", &stages.Stager{}, `yaml:"stages,omitempty"`).
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
	ErrNotABuyerTrade  = errors.New("not a buyer trade")
	ErrNotASellerTrade = errors.New("not a seller trade")
)

func (bt *baseTrade) GenerateBuyProposal() (*BuyProposal, error) {
	// only a buyer can generate a proposal
	if bt.Role != roles.Buyer {
		return nil, ErrNotABuyerTrade
	}
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

func (bt *baseTrade) AcceptBuyProposal(prop *BuyProposal) error {
	// only the seller can accept a trade
	if bt.Role != roles.Seller {
		return ErrNotASellerTrade
	}
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
	ErrMismatchTokenHash   = errors.New("mismatching token hash")
	ErrInvalidLockInterval = errors.New("invalid lock interval")
	ErrMismatchKeyData     = errors.New("mismatching key data")
)

// SetLocks sets the buyer and seller locks
func (bt *baseTrade) SetLocks(locks *BuyProposalResponse) error {
	bd, err := locks.Buyer.LockData()
	if err != nil {
		return err
	}
	sd, err := locks.Seller.LockData()
	if err != nil {
		return err
	}
	if bd.Locktime.Sub(sd.Locktime) != time.Duration(bt.Duration)/2 {
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

var ErrNotUTXO = errors.New("not a utxo crypto")

func (bt *baseTrade) newRedeemTxUTXO(dest key.KeyData, fee uint64) (tx.Tx, error) {
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
	tx.AddOutput(amount-fee, gen.P2PKHHash(dest))
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

func (bt *baseTrade) newRedeemTx(dest key.KeyData, fee uint64) (tx.Tx, error) {
	switch bt.TraderInfo.Crypto.Type {
	case cryptos.UTXO:
		return bt.newRedeemTxUTXO(dest, fee)
	case cryptos.StateBased:
		panic("not implemented")
	default:
		panic("not implemented")
	}
}

// RedeemTxFixedFee returns the redeem transaction for the locked funds with a fixed fee
func (bt *baseTrade) RedeemTxFixedFee(dest key.KeyData, fee uint64) (tx.Tx, error) {
	return bt.newRedeemTx(dest, fee)
}

// RedeemTx returns the redeem transaction for the locked funds
func (bt *baseTrade) RedeemTx(dest key.KeyData, feePerByte uint64) (tx.Tx, error) {
	tx, err := bt.newRedeemTx(dest, 0)
	if err != nil {
		return nil, err
	}
	return bt.newRedeemTx(dest, feePerByte*tx.SerializedSize())
}

func (bt *baseTrade) newRecoveryTxUTXO(dest key.KeyData, fee uint64) (tx.Tx, error) {
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
	tx.AddOutput(amount-fee, gen.P2PKHHash(dest))
	lst, err := bt.RecoverableFunds.Lock().LockData()
	if err != nil {
		return nil, err
	}
	tx.SetLockTime(lst.Locktime.UTC())
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

func (bt *baseTrade) newRecoveryTx(dest key.KeyData, fee uint64) (tx.Tx, error) {
	switch txType := bt.OwnInfo.Crypto.Type; txType {
	case cryptos.UTXO:
		return bt.newRecoveryTxUTXO(dest, fee)
	case cryptos.StateBased:
		panic("not implemented")
	default:
		panic("not implemented")
	}
}

// RecoveryTxFixedFee returns the recovery transaction for the locked funds with a fixed fee
func (bt *baseTrade) RecoveryTxFixedFee(dest key.KeyData, fee uint64) (tx.Tx, error) {
	return bt.newRecoveryTx(dest, fee)
}

// RecoveryTx returns the recovery transaction for the locked funds
func (bt *baseTrade) RecoveryTx(dest key.KeyData, feePerByte uint64) (tx.Tx, error) {
	tx, err := bt.newRecoveryTx(dest, 0)
	if err != nil {
		return nil, err
	}
	return bt.newRecoveryTx(dest, tx.SerializedSize()*feePerByte)
}
