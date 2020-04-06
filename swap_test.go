package atomicswap_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
	"transmutate.io/pkg/atomicswap"
	"transmutate.io/pkg/atomicswap/addr"
	"transmutate.io/pkg/atomicswap/hash"
	"transmutate.io/pkg/atomicswap/params"
	"transmutate.io/pkg/atomicswap/script"
	"transmutate.io/pkg/atomicswap/stages"
	"transmutate.io/pkg/btccore"
	bctypes "transmutate.io/pkg/btccore/types"
)

func envOr(envName string, defaultValue string) string {
	e, ok := os.LookupEnv(envName)
	if !ok {
		return defaultValue
	}
	return e
}

type testCrypto struct {
	crypto    params.Crypto
	cl        btccore.Client
	minerAddr string
	amount    bctypes.Amount
	decimals  int
}

var testCryptos = []*testCrypto{
	{
		params.Bitcoin,
		btccore.NewClientBTC(
			envOr("GO_TEST_BTC_NODE", "bitcoin-core-testnet.docker:4444"),
			"admin", "pass", false,
		),
		"",
		bctypes.Amount("2.2"),
		8,
	},
	{
		params.Litecoin,
		btccore.NewClientLTC(
			envOr("GO_TEST_LTC_NODE", "litecoin-testnet.docker:4444"),
			"admin", "pass", false,
		),
		"",
		bctypes.Amount("1.1"),
		8,
	},
	{
		params.Dogecoin,
		btccore.NewClientDOGE(
			envOr("GO_TEST_DOGE_NODE", "dogecoin-testnet.docker:4444"),
			"admin", "pass", false,
		),
		"",
		bctypes.Amount("1.1"),
		8,
	},
}

func init() {
	for _, i := range testCryptos {
		if err := setupMiner(i); err != nil {
			panic(err)
		}
	}
}

func setupMiner(c *testCrypto) error {
	// find existing addresses
	funds, err := c.cl.ReceivedByAddress(0, true, nil)
	if err != nil {
		return err
	}
	if len(funds) > 0 {
		// use already existing address
		c.minerAddr = funds[0].Address
	} else {
		// generate new address
		addr, err := c.cl.NewAddress()
		if err != nil {
			return err
		}
		c.minerAddr = addr
	}
	// generate funds
	for {
		// check balance
		amt, err := c.cl.Balance(0)
		if err != nil {
			return err
		}
		if amt.UInt64(c.decimals) >= c.amount.UInt64(c.decimals) {
			break
		}
		_, err = c.cl.GenerateToAddress(1, c.minerAddr)
		if err != nil {
			return err
		}
	}
	return nil
}

const stdFee uint64 = 1000

func fmtPrintf(f string, a ...interface{}) { fmt.Printf(f, a...) }

type printfFunc = func(f string, a ...interface{})

func newPrintf(oldPrintf printfFunc, name string) printfFunc {
	return func(f string, args ...interface{}) { oldPrintf(name+": "+f, args...) }
}

func TestAtomicSwapRedeem(t *testing.T) {
	for _, i := range testCryptos[1:] {
		t.Run(i.crypto.String()+"_bitcoin", newTestAtomicSwapRedeemManualExchange(testCryptos[0], i))
	}
}

// handle stages
func handleStage(
	pf printfFunc,
	a2b chan interface{},
	at *atomicswap.Trade,
	ownCrypto *testCrypto,
	traderCrypto *testCrypto,
) error {
	switch at.Stage {
	case stages.SharePublicKeyHash:
		// use a channel to exchange data back and forth
		a2b <- bctypes.Bytes(hash.Hash160(at.Own.RedeemKey.Public().SerializeCompressed()))
		pf("sent public key hash\n")
	case stages.ReceivePublicKeyHash:
		// use a channel to exchange data back and forth
		at.Trader.RedeemKeyHash = (<-a2b).(bctypes.Bytes)
		pf("received public key hash: %s\n", at.Trader.RedeemKeyHash.Hex())
	case stages.ShareTokenHash:
		// use a channel to exchange data back and forth
		a2b <- at.TokenHash()
		pf("sent token hash\n")
	case stages.ReceiveTokenHash:
		// use a channel to exchange data back and forth
		at.SetTokenHash((<-a2b).(bctypes.Bytes))
		pf("received token hash: %s\n", at.TokenHash().Hex())
	case stages.ReceiveLockScript:
		// use a channel to exchange data back and forth
		ls := (<-a2b).(bctypes.Bytes)
		ds, err := script.DisassembleString(ls)
		if err != nil {
			return err
		}
		// check lock script
		if err := at.CheckTraderLockScript(ls); err != nil {
			pf("received invalid lock script: %s %s\n", ls.Hex(), ds)
			return err
		}
		// save lock script after checking
		at.Trader.LockScript = ls
		pf("received lock script: %s %s\n", ls.Hex(), ds)
	case stages.GenerateLockScript:
		// generate lock script
		if err := at.GenerateOwnLockScript(); err != nil {
			return err
		}
		pf("generated lock script: %s\n", at.Own.LockScript.Hex())
	case stages.ShareLockScript:
		// use a channel to exchange data back and forth
		a2b <- at.Own.LockScript
		pf("sent lock script\n")
	case stages.WaitLockTransaction:
		// use the client to find a deposit
		txOut, err := waitDeposit(
			traderCrypto.cl,
			params.Networks[traderCrypto.crypto][params.RegressionNet],
			at.Trader.LastBlockHeight,
			hash.Hash160(at.Trader.LockScript),
		)
		if err != nil {
			return err
		}
		at.Trader.LastBlockHeight = txOut.blockHeight
		// save redeeamable output
		at.AddRedeemableOutput(&atomicswap.Output{
			TxID: txOut.txID,
			N:    txOut.n,
		})
		at.Trader.LastBlockHeight = txOut.blockHeight
		pf(
			"found lock transaction: %s, %d (block %d)\n",
			hex.EncodeToString(txOut.txID),
			txOut.n,
			txOut.blockHeight,
		)
	case stages.LockFunds:
		// calculate deposit address
		depositAddr, err := addr.P2SH(hash.Hash160(at.Own.LockScript), params.Networks[ownCrypto.crypto][params.RegressionNet])
		if err != nil {
			return err
		}
		txID, err := ownCrypto.cl.SendToAddress(depositAddr, bctypes.NewAmount(at.Own.Amount.UInt64(ownCrypto.decimals)+stdFee, 8))
		if err != nil {
			return err
		}
		if _, err = ownCrypto.cl.GenerateToAddress(101, ownCrypto.minerAddr); err != nil {
			return err
		}
		// find transaction
		tx, err := ownCrypto.cl.Transaction(txID)
		if err != nil {
			return err
		}
		// save recoverable output
		for _, i := range tx.Outputs {
			if i.UnlockScript.Type == "scripthash" && len(i.UnlockScript.Addresses) > 0 && i.UnlockScript.Addresses[0] == depositAddr {
				at.SetRecoverableOutput(&atomicswap.Output{TxID: txID, N: uint32(i.N)})
				break
			}
		}
		if at.Outputs == nil || at.Outputs.Recoverable == nil {
			return errors.New("recoverable output not found")
		}
		pf("funds locked: tx %s\n", txID.Hex())
	case stages.WaitRedeemTransaction:
		// use the client to find the trader redeem transaction and extract token
		token, err := waitRedeem(ownCrypto.cl, params.Networks[ownCrypto.crypto][params.RegressionNet], at.Own.LastBlockHeight, at.Outputs.Recoverable.TxID, int(at.Outputs.Recoverable.N))
		if err != nil {
			return err
		}
		// save token
		at.SetToken(token)
		pf("redeem transaction found: token %s\n", hex.EncodeToString(token))
	case stages.RedeemFunds:
		redeemTx, err := at.RedeemTransaction(at.Trader.Amount.UInt64(traderCrypto.decimals))
		b, err := redeemTx.Serialize()
		if err != nil {
			return err
		}
		txID, err := traderCrypto.cl.SendRawTransaction(b)
		if err != nil {
			return err
		}
		if _, err = traderCrypto.cl.GenerateToAddress(101, traderCrypto.minerAddr); err != nil {
			return err
		}
		pf("funds redeemed: tx %s\n", txID.Hex())
	default:
		return errors.New("invalid stage")
	}
	return nil
}

func handleTrade(
	pf printfFunc,
	a2b chan interface{},
	ownCrypto *testCrypto,
	traderCrypto *testCrypto,
	htlcDuration time.Duration,
	at *atomicswap.Trade,
	failToRedeem bool,
) error {
	for {
		pf("stage: %s\n", at.Stage)
		if at.Stage == stages.Done {
			break
		}
		if err := handleStage(pf, a2b, at, ownCrypto, traderCrypto); err != nil {
			return err
		}
		at.NextStage()
		if failToRedeem && (at.Stage == stages.RedeemFunds || at.Stage == stages.WaitRedeemTransaction) {
			break
		}
	}
	b, err := yaml.Marshal(at)
	if err != nil {
		return err
	}
	pf("trade data:\n%s\n", string(b))
	at2 := &atomicswap.Trade{}
	if err = yaml.Unmarshal(b, at2); err != nil {
		return err
	}
	if !reflect.DeepEqual(at, at) {
		return errors.New("marshal/unmarshal error")
	}
	return nil
}

func newTestAtomicSwapRedeemManualExchange(btc, alt *testCrypto) func(*testing.T) {
	return func(t *testing.T) {
		// generate communication channel
		a2b := make(chan interface{})
		defer close(a2b)
		eg := &errgroup.Group{}
		htlcDuration := 48 * time.Hour
		// alice (alt)
		eg.Go(func() error {
			at, err := atomicswap.NewBuyerTrade(alt.amount, alt.crypto, btc.amount, btc.crypto)
			if err != nil {
				return err
			}
			return handleTrade(newPrintf(t.Logf, "alice (buyer)"), a2b, alt, btc, htlcDuration, at, false)
		})
		// bob (BTC)
		eg.Go(func() error {
			at, err := atomicswap.NewSellerTrade(btc.amount, btc.crypto, alt.amount, alt.crypto)
			if err != nil {
				return err
			}
			return handleTrade(newPrintf(t.Logf, "bob (seller)"), a2b, btc, alt, htlcDuration, at, false)
		})
		err := eg.Wait()
		require.NoError(t, err, "unexpected error")
	}
}

type txOut struct {
	blockHeight uint64
	txID        []byte
	n           uint32
}

func waitDeposit(cl btccore.Client, chainParams *params.Params, startBlockHeight uint64, scriptHash []byte) (*txOut, error) {
	depositAddr, err := addr.P2SH(scriptHash, chainParams)
	if err != nil {
		return nil, err
	}
	next, closeIter := btccore.NewTransactionIterator(cl, startBlockHeight)
	defer closeIter()
	for {
		tx, err := next()
		if err != nil {
			return nil, err
		}
		for _, j := range tx.Outputs {
			if j.UnlockScript.Type != "scripthash" {
				continue
			}
			if len(j.UnlockScript.Addresses) < 1 {
				continue
			}
			if j.UnlockScript.Addresses[0] != depositAddr {
				continue
			}
			return &txOut{blockHeight: startBlockHeight, txID: tx.ID, n: uint32(j.N)}, nil
		}
	}
}

func waitRedeem(cl btccore.Client, chainParams *params.Params, startBlockHeight uint64, txID []byte, idx int) ([]byte, error) {
	next, closeIter := btccore.NewTransactionIterator(cl, startBlockHeight)
	defer closeIter()
	for {
		tx, err := next()
		if err != nil {
			return nil, err
		}
		for _, j := range tx.Inputs {
			if !bytes.Equal(j.TransactionID, txID) || j.N != idx {
				continue
			}
			inst := strings.Split(j.LockScript.Asm, " ")
			if len(inst) != 5 || !strings.HasSuffix(inst[0], "[ALL]") || inst[3] != "0" {
				continue
			}
			b, err := hex.DecodeString(inst[2])
			if err != nil {
				continue
			}
			return b, nil
		}
	}
}

func recoverFunds(ownCrypto *testCrypto, at *atomicswap.Trade, pf printfFunc) error {
	tx, err := at.RecoveryTransaction(at.Own.Amount.UInt64(8))
	if err != nil {
		return err
	}
	b, err := tx.Serialize()
	if err != nil {
		return err
	}
	pf("generate recovery transaction: %s\n", hex.EncodeToString(b))
	txID, err := ownCrypto.cl.SendRawTransaction(b)
	if err != nil {
		return err
	}
	if _, err = ownCrypto.cl.GenerateToAddress(1, ownCrypto.minerAddr); err != nil {
		return err
	}
	pf("funds recovered: tx: %s\n", txID.Hex())
	return nil
}

func newTestAtomicSwapRecover(btc, alt *testCrypto) func(*testing.T) {
	return func(t *testing.T) {
		// generate communication channel
		a2b := make(chan interface{})
		defer close(a2b)
		eg := &errgroup.Group{}
		const htlcDuration = 0
		// alice (alt)
		eg.Go(func() error {
			at, err := atomicswap.NewBuyerTrade(alt.amount, alt.crypto, btc.amount, btc.crypto)
			if err != nil {
				return err
			}
			pf := newPrintf(t.Logf, "alice (buyer)")
			err = handleTrade(pf, a2b, alt, btc, htlcDuration, at, true)
			if err != nil {
				return err
			}
			time.Sleep(htlcDuration * 2)
			return recoverFunds(alt, at, pf)
		})
		// bob (BTC)
		eg.Go(func() error {
			at, err := atomicswap.NewSellerTrade(btc.amount, btc.crypto, alt.amount, alt.crypto)
			if err != nil {
				return err
			}
			pf := newPrintf(t.Logf, "bob (seller)")
			err = handleTrade(pf, a2b, btc, alt, htlcDuration, at, true)
			if err != nil {
				return err
			}
			time.Sleep(htlcDuration * 2)
			return recoverFunds(btc, at, pf)
		})
		err := eg.Wait()
		require.NoError(t, err, "unexpected error")
	}
}

func TestAtomicSwapRecover(t *testing.T) {
	for _, i := range testCryptos[1:] {
		t.Run(i.crypto.String()+"_bitcoin", newTestAtomicSwapRecover(testCryptos[0], i))
	}
}
