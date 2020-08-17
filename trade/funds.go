package trade

import (
	"encoding/hex"
	"errors"
	"time"

	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/key"
	"github.com/transmutate-io/atomicswap/params"
	"github.com/transmutate-io/atomicswap/script"
	"github.com/transmutate-io/cryptocore/types"
)

type (
	// FundsData holds information about funds
	FundsData interface {
		// AddFunds adds funds to the manager
		AddFunds(funds interface{})
		// Funds returns the fund on the manager
		Funds() interface{}
		// SetLock sets the lock for the funds
		SetLock(lock Lock)
		// Lock returns the lock for the funds
		Lock() Lock
	}

	// Lock represents a cryptocurrency lock
	Lock interface {
		// Bytes returns the lock bytes
		Bytes() types.Bytes
		// LockData returns the lock data
		LockData() (*LockData, error)
		// Address returns the address for the given chain
		Address(chain params.Chain) (string, error)
	}

	// LockData represents a lock
	LockData struct {
		LockTime        time.Time
		TokenHash       types.Bytes
		RedeemKeyData   key.KeyData
		RecoveryKeyData key.KeyData
	}

	// Output represents an output
	Output struct {
		TxID   types.Bytes `yaml:"txid"`
		N      uint32      `yaml:"n"`
		Amount uint64      `yaml:"amount"`
	}
)

// ErrInvalidLockScript is returns when the lock script is invalid
var ErrInvalidLockScript = errors.New("invalid lock script")

var expHTLC = []string{
	"OP_IF",
	"", "OP_CHECKLOCKTIMEVERIFY", "OP_DROP",
	"OP_DUP", "OP_HASH160", "", "OP_EQUALVERIFY", "OP_CHECKSIG",
	"OP_ELSE",
	"OP_SHA256", "OP_RIPEMD160", "", "OP_EQUALVERIFY",
	"OP_DUP", "OP_HASH160", "", "OP_EQUALVERIFY", "OP_CHECKSIG", "OP_ENDIF",
}

func parseLockScript(c *cryptos.Crypto, ls []byte) (*LockData, error) {
	inst, err := script.DisassembleStrings(c, ls)
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
	n, err := script.NewIntParserBTC().ParseInt64(b)
	if err != nil {
		return nil, err
	}
	r := &LockData{}
	r.LockTime = time.Unix(n, 0)
	// token hash
	if r.TokenHash, err = hex.DecodeString(inst[12]); err != nil {
		return nil, err
	}
	// recovery key hash
	if r.RecoveryKeyData, err = hex.DecodeString(inst[6]); err != nil {
		return nil, err
	}
	// redeem key hash
	if r.RedeemKeyData, err = hex.DecodeString(inst[16]); err != nil {
		return nil, err
	}
	return r, nil
}

func newFundsData(c *cryptos.Crypto) (FundsData, error) {
	nf, ok := newFundsDataFuncs[c.Name]
	if !ok {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return nf(), nil
}

func newFundsLock(c *cryptos.Crypto, b []byte) (Lock, error) {
	nf, ok := newFundsLockFuncs[c.Name]
	if !ok {
		return nil, cryptos.InvalidCryptoError(c.Name)
	}
	return nf(b), nil
}
