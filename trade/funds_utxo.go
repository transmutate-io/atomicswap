package trade

import (
	"encoding/hex"
	"errors"
	"time"

	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/hash"
	"transmutate.io/pkg/atomicswap/networks"
	"transmutate.io/pkg/atomicswap/params"
	"transmutate.io/pkg/atomicswap/script"
	"transmutate.io/pkg/cryptocore/types"
)

// Output represents an output
type Output struct {
	TxID   types.Bytes `yaml:"txid"`
	N      uint32      `yaml:"n"`
	Amount uint64      `yaml:"amount"`
}

type fundsUTXO struct {
	Outputs    []*Output   `yaml:"outputs"`
	LockScript types.Bytes `yaml:"lock_script"`
}

func newFundsUTXO() *fundsUTXO { return &fundsUTXO{Outputs: make([]*Output, 0, 4)} }

func (f *fundsUTXO) CryptoType() cryptos.Type { return cryptos.UTXO }

func (f *fundsUTXO) Funds() interface{} { return f.Outputs }

func (f *fundsUTXO) AddFunds(funds interface{}) {
	f.Outputs = append(f.Outputs, funds.(*Output))
}

func (f fundsUTXO) Lock() Lock { return fundsUTXOLock(f.LockScript) }

func (f *fundsUTXO) SetLock(lock Lock) { f.LockScript = lock.(fundsUTXOLock).Bytes() }

type fundsUTXOLock types.Bytes

func (fl fundsUTXOLock) Bytes() types.Bytes { return types.Bytes(fl) }

func (fl fundsUTXOLock) Data() types.Bytes { return fl.Bytes() }

func (fl fundsUTXOLock) LockData() (*LockData, error) {
	return parseLockScript(fl)
}

func (fl fundsUTXOLock) Address(crypto *cryptos.Crypto, chain params.Chain) (string, error) {
	return networks.AllByName[crypto.Name][chain].P2SH(hash.Hash160(fl))
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

func parseLockScript(ls []byte) (*LockData, error) {
	eng := script.NewEngineBTC()
	// check contract format
	inst, err := eng.DisassembleStrings(ls)
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
	n, err := eng.ParseInt64(b)
	if err != nil {
		return nil, err
	}
	r := &LockData{}
	r.Locktime = time.Unix(n, 0)
	// token hash
	if r.TokenHash, err = hex.DecodeString(inst[11]); err != nil {
		return nil, err
	}
	// // redeem key hash
	// if r.redeemKeyHash, err = hex.DecodeString(inst[15]); err != nil {
	// 	return nil, err
	// }
	return r, nil
}

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
