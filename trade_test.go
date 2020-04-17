package atomicswap_test

import (
	"os"
	"testing"
	// 	"bytes"
	// 	"encoding/hex"
	// 	"errors"
	// 	"fmt"
	// 	"os"
	// 	"reflect"
	// 	"strings"
	// 	"testing"
	// 	"time"

	// 	"github.com/stretchr/testify/require"
	// 	"golang.org/x/sync/errgroup"
	// 	"gopkg.in/yaml.v2"
	// 	"transmutate.io/pkg/atomicswap"
	// 	"transmutate.io/pkg/atomicswap/hash"
	// 	"transmutate.io/pkg/atomicswap/params"
	// 	"transmutate.io/pkg/atomicswap/params/cryptos"
	// 	"transmutate.io/pkg/atomicswap/script"
	// 	"transmutate.io/pkg/atomicswap/stages"
	"transmutate.io/pkg/cryptocore"
	cctypes "transmutate.io/pkg/cryptocore/types"
)

func envOr(envName string, defaultValue string) string {
	e, ok := os.LookupEnv(envName)
	if !ok {
		return defaultValue
	}
	return e
}

type testCrypto struct {
	crypto    string
	cl        cryptocore.Client
	minerAddr string
	amount    cctypes.Amount
	decimals  int
}

var testCryptos = []*testCrypto{
	{
		"bitcoin",
		cryptocore.NewClientBTC(
			envOr("GO_TEST_BTC_NODE", "bitcoin-core-regtest.docker:4444"),
			"admin", "pass", false,
		),
		"",
		cctypes.Amount("50"),
		8,
	},
	{
		"litecoin",
		cryptocore.NewClientLTC(
			envOr("GO_TEST_LTC_NODE", "litecoin-regtest.docker:4444"),
			"admin", "pass", false,
		),
		"",
		cctypes.Amount("50"),
		8,
	},
	// {
	// 	params.Dogecoin,
	// 	cryptocore.NewClientDOGE(
	// 		envOr("GO_TEST_DOGE_NODE", "dogecoin-regtest.docker:4444"),
	// 		"admin", "pass", false,
	// 	),
	// 	"",
	// 	cctypes.Amount("50"),
	// 	8,
	// },
	// {
	// 	"bitcoin-cash",
	// 	cryptocore.NewClientBTCCash(
	// 		envOr("GO_TEST_BCH_NODE", "bitcoin-cash-regtest.docker:4444"),
	// 		"admin", "pass", false,
	// 	),
	// 	"",
	// 	cctypes.Amount("50"),
	// 	8,
	// },
}

// func init() {
// 	for _, i := range testCryptos {
// 		if err := setupMiner(i); err != nil {
// 			panic(err)
// 		}
// 	}
// }

// func setupMiner(c *testCrypto) error {
// 	// find existing addresses
// 	funds, err := c.cl.ReceivedByAddress(0, true, nil)
// 	if err != nil {
// 		return err
// 	}
// 	if len(funds) > 0 {
// 		// use already existing address
// 		c.minerAddr = funds[0].Address
// 	} else {
// 		// generate new address
// 		addr, err := c.cl.NewAddress()
// 		if err != nil {
// 			return err
// 		}
// 		c.minerAddr = addr
// 	}
// 	// generate funds
// 	for {
// 		// check balance
// 		amt, err := c.cl.Balance(0)
// 		if err != nil {
// 			return err
// 		}
// 		if amt.UInt64(c.decimals) >= c.amount.UInt64(c.decimals) {
// 			break
// 		}
// 		_, err = c.cl.GenerateToAddress(1, c.minerAddr)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func fmtPrintf(f string, a ...interface{}) { fmt.Printf(f, a...) }

// type printfFunc = func(f string, a ...interface{})

// func newPrintf(oldPrintf printfFunc, name string) printfFunc {
// 	return func(f string, args ...interface{}) { oldPrintf(name+": "+f, args...) }
// }

// func TestAtomicSwapRedeemManualExchange(t *testing.T) {
// 	for _, i := range testCryptos[1:] {
// 		t.Run(i.crypto+"_bitcoin", newTestAtomicSwapRedeem(testCryptos[0], i, true))
// 	}
// }

// func TestAtomicSwapRedeemOnChainExchange(t *testing.T) {
// 	for _, i := range testCryptos[1:] {
// 		t.Run(i.crypto+"_bitcoin", newTestAtomicSwapRedeem(testCryptos[0], i, false))
// 	}
// }

// // handle stages
// func handleStage(
// 	pf printfFunc,
// 	a2b chan interface{},
// 	at *atomicswap.Trade,
// 	ownCrypto *testCrypto,
// 	traderCrypto *testCrypto,
// 	nOutputs int,
// ) error {
// 	switch at.Stage {
// 	case stages.SharePublicKeyHash:
// 		// use a channel to exchange data back and forth
// 		a2b <- cctypes.Bytes(hash.Hash160(at.Own.RedeemKey.Public().SerializeCompressed()))
// 		pf("sent public key hash\n")
// 	case stages.ReceivePublicKeyHash:
// 		// use a channel to exchange data back and forth
// 		at.Trader.RedeemKeyHash = (<-a2b).(cctypes.Bytes)
// 		pf("received public key hash: %s\n", at.Trader.RedeemKeyHash.Hex())
// 	case stages.ShareTokenHash:
// 		// use a channel to exchange data back and forth
// 		a2b <- at.TokenHash()
// 		pf("sent token hash\n")
// 	case stages.ReceiveTokenHash:
// 		// use a channel to exchange data back and forth
// 		at.SetTokenHash((<-a2b).(cctypes.Bytes))
// 		pf("received token hash: %s\n", at.TokenHash().Hex())
// 	case stages.ReceiveLockScript:
// 		// use a channel to exchange data back and forth
// 		ls := (<-a2b).(cctypes.Bytes)
// 		ds, err := script.DisassembleString(ls)
// 		if err != nil {
// 			return err
// 		}
// 		// check lock script
// 		if err := at.CheckTraderLockScript(ls); err != nil {
// 			pf("received invalid lock script: %s %s\n", ls.Hex(), ds)
// 			return err
// 		}
// 		// save lock script after checking
// 		at.Trader.LockScript = ls
// 		pf("received lock script: %s %s\n", ls.Hex(), ds)
// 	case stages.GenerateLockScript:
// 		// generate lock script
// 		if err := at.GenerateOwnLockScript(); err != nil {
// 			return err
// 		}
// 		pf("generated lock script: %s\n", at.Own.LockScript.Hex())
// 	case stages.ShareLockScript:
// 		// use a channel to exchange data back and forth
// 		a2b <- at.Own.LockScript
// 		pf("sent lock script\n")
// 	case stages.WaitLockTransaction:
// 		for i := 0; i < nOutputs; i++ {
// 			// use the client to find a deposit
// 			txOut, err := waitDeposit(
// 				traderCrypto,
// 				params.Networks[traderCrypto.crypto][params.RegressionNet],
// 				at.Trader.LastBlockHeight,
// 				hash.Hash160(at.Trader.LockScript),
// 			)
// 			if err != nil {
// 				return err
// 			}
// 			// save redeeamable output
// 			at.AddRedeemableOutput(&atomicswap.Output{
// 				TxID:   txOut.txID,
// 				N:      txOut.n,
// 				Amount: txOut.amount,
// 			})
// 			at.Trader.LastBlockHeight = txOut.blockHeight + 1
// 			pf(
// 				"found lock transaction: txid: %s, n: %d amount: %d (block %d)\n",
// 				hex.EncodeToString(txOut.txID),
// 				txOut.n,
// 				txOut.amount,
// 				txOut.blockHeight,
// 			)
// 		}
// 	case stages.LockFunds:
// 		for i := 0; i < nOutputs; i++ {
// 			// calculate deposit address
// 			depositAddr, err := params.Networks[ownCrypto.crypto][params.RegressionNet].P2SH(hash.Hash160(at.Own.LockScript))
// 			if err != nil {
// 				return err
// 			}
// 			pf("deposit address: %s\n", depositAddr)
// 			txID, err := ownCrypto.cl.SendToAddress(depositAddr, cctypes.NewAmount(at.Own.Amount.UInt64(ownCrypto.decimals), 8))
// 			if err != nil {
// 				return err
// 			}
// 			if _, err = ownCrypto.cl.GenerateToAddress(101, ownCrypto.minerAddr); err != nil {
// 				return err
// 			}
// 			// find transaction
// 			tx, err := ownCrypto.cl.Transaction(txID)
// 			if err != nil {
// 				return err
// 			}
// 			pf("deposit tx: %s\n", txID.Hex())
// 			// save recoverable output
// 			for _, i := range tx.Outputs {
// 				pf("\toutput: %#v\n", i.UnlockScript)
// 				if i.UnlockScript.Type == "scripthash" && len(i.UnlockScript.Addresses) > 0 && i.UnlockScript.Addresses[0] == depositAddr {
// 					at.AddRecoverableOutput(&atomicswap.Output{
// 						TxID:   txID,
// 						N:      uint32(i.N),
// 						Amount: i.Value.UInt64(ownCrypto.decimals),
// 					})
// 					break
// 				}
// 			}
// 			if at.Outputs == nil || at.Outputs.Recoverable == nil {
// 				return errors.New("recoverable output not found")
// 			}
// 			pf("funds locked: tx %s\n", txID.Hex())
// 		}
// 	case stages.WaitRedeemTransaction:
// 		// use the client to find the trader redeem transaction and extract token
// 		token, err := waitRedeem(ownCrypto.cl, params.Networks[ownCrypto.crypto][params.RegressionNet], at.Own.LastBlockHeight, at.Outputs.Recoverable[0].TxID, int(at.Outputs.Recoverable[0].N))
// 		if err != nil {
// 			return err
// 		}
// 		// save token
// 		at.SetToken(token)
// 		pf("redeem transaction found: token %s\n", hex.EncodeToString(token))
// 	case stages.RedeemFunds:
// 		redeemTx, err := at.RedeemTransaction(stdFeePerByte)
// 		b, err := redeemTx.Serialize()
// 		if err != nil {
// 			return err
// 		}
// 		txID, err := traderCrypto.cl.SendRawTransaction(b)
// 		if err != nil {
// 			return err
// 		}
// 		if _, err = traderCrypto.cl.GenerateToAddress(101, traderCrypto.minerAddr); err != nil {
// 			return err
// 		}
// 		pf("funds redeemed: tx %s\n", txID.Hex())
// 	default:
// 		return errors.New("invalid stage")
// 	}
// 	return nil
// }

// func handleTrade(
// 	pf printfFunc,
// 	a2b chan interface{},
// 	ownCrypto *testCrypto,
// 	traderCrypto *testCrypto,
// 	htlcDuration time.Duration,
// 	at *atomicswap.Trade,
// 	failToRedeem bool,
// 	nOutputs int,
// ) error {
// 	for {
// 		pf("stage done: %s\n", at.Stage)
// 		if at.Stage == stages.Done {
// 			break
// 		}
// 		if err := handleStage(pf, a2b, at, ownCrypto, traderCrypto, nOutputs); err != nil {
// 			pf("got a oopsie: %v\n", err)
// 			return err
// 		}
// 		at.NextStage()
// 		if failToRedeem && (at.Stage == stages.RedeemFunds || at.Stage == stages.WaitRedeemTransaction) {
// 			break
// 		}
// 	}
// 	b, err := yaml.Marshal(at)
// 	if err != nil {
// 		return err
// 	}
// 	pf("trade data:\n%s\n", string(b))
// 	at2 := &atomicswap.Trade{}
// 	if err = yaml.Unmarshal(b, at2); err != nil {
// 		return err
// 	}
// 	if !reflect.DeepEqual(at, at) {
// 		return errors.New("marshal/unmarshal error")
// 	}
// 	return nil
// }

// func newTestAtomicSwapRedeem(btc, alt *testCrypto, manualExchange bool) func(*testing.T) {
// 	return func(t *testing.T) {
// 		// generate communication channel
// 		a2b := make(chan interface{})
// 		defer close(a2b)
// 		eg := &errgroup.Group{}
// 		htlcDuration := 48 * time.Hour
// 		altCrypto, err := cryptos.ParseCrypto(alt.crypto)
// 		require.NoError(t, err, "can't parse alt")
// 		btcCrypto, err := cryptos.ParseCrypto("bitcoin")
// 		require.NoError(t, err, "can't parse bitcoin")
// 		// alice (alt)
// 		eg.Go(func() error {
// 			at, err := atomicswap.NewBuyerTrade(
// 				cctypes.Amount("1"),
// 				altCrypto,
// 				cctypes.Amount("1"),
// 				btcCrypto,
// 				// manualExchange,
// 			)
// 			if err != nil {
// 				return err
// 			}
// 			return handleTrade(newPrintf(t.Logf, "alice (buyer)"), a2b, alt, btc, htlcDuration, at, false, 2)
// 		})
// 		// bob (BTC)
// 		eg.Go(func() error {
// 			at, err := atomicswap.NewSellerTrade(
// 				cctypes.Amount("1"),
// 				btcCrypto,
// 				cctypes.Amount("1"),
// 				altCrypto,
// 			)
// 			if err != nil {
// 				return err
// 			}
// 			return handleTrade(newPrintf(t.Logf, "bob (seller)"), a2b, btc, alt, htlcDuration, at, false, 2)
// 		})
// 		err = eg.Wait()
// 		require.NoError(t, err, "unexpected error")
// 	}
// }

// type txOut struct {
// 	blockHeight uint64
// 	txID        []byte
// 	n           uint32
// 	amount      uint64
// }

// func waitDeposit(crypto *testCrypto, chainParams params.Params, startBlockHeight uint64, scriptHash []byte) (*txOut, error) {
// 	depositAddr, err := chainParams.P2SH(scriptHash)
// 	if err != nil {
// 		return nil, err
// 	}
// 	next, closeIter := cryptocore.NewBlockIterator(crypto.cl, startBlockHeight)
// 	defer closeIter()
// 	for {
// 		blk, err := next()
// 		if err != nil {
// 			return nil, err
// 		}
// 		for _, i := range blk.Transactions {
// 			tx, err := crypto.cl.Transaction(i)
// 			if err != nil {
// 				return nil, err
// 			}
// 			for _, j := range tx.Outputs {
// 				if j.UnlockScript.Type != "scripthash" {
// 					continue
// 				}
// 				if len(j.UnlockScript.Addresses) < 1 {
// 					continue
// 				}
// 				if j.UnlockScript.Addresses[0] != depositAddr {
// 					continue
// 				}
// 				return &txOut{
// 					blockHeight: startBlockHeight,
// 					txID:        tx.ID,
// 					n:           uint32(j.N),
// 					amount:      j.Value.UInt64(crypto.decimals),
// 				}, nil
// 			}
// 		}
// 		startBlockHeight++
// 	}
// }

// func waitRedeem(cl cryptocore.Client, chainParams params.Params, startBlockHeight uint64, txID []byte, idx int) ([]byte, error) {
// 	next, closeIter := cryptocore.NewTransactionIterator(cl, startBlockHeight)
// 	defer closeIter()
// 	for {
// 		tx, err := next()
// 		if err != nil {
// 			return nil, err
// 		}
// 		for _, j := range tx.Inputs {
// 			if !bytes.Equal(j.TransactionID, txID) || j.N != idx {
// 				continue
// 			}
// 			inst := strings.Split(j.LockScript.Asm, " ")
// 			if len(inst) != 5 || !strings.HasSuffix(inst[0], "[ALL]") || inst[3] != "0" {
// 				continue
// 			}
// 			b, err := hex.DecodeString(inst[2])
// 			if err != nil {
// 				continue
// 			}
// 			return b, nil
// 		}
// 	}
// }

// const stdFeePerByte = 5

// func recoverFunds(ownCrypto *testCrypto, at *atomicswap.Trade, pf printfFunc) error {
// 	tx, err := at.RecoveryTransaction(stdFeePerByte)
// 	if err != nil {
// 		return err
// 	}
// 	b, err := tx.Serialize()
// 	if err != nil {
// 		return err
// 	}
// 	txID, err := ownCrypto.cl.SendRawTransaction(b)
// 	if err != nil {
// 		return err
// 	}
// 	if _, err = ownCrypto.cl.GenerateToAddress(1, ownCrypto.minerAddr); err != nil {
// 		return err
// 	}
// 	pf("funds recovered: tx: %s\n", txID.Hex())
// 	return nil
// }

// func newTestAtomicSwapRecover(btc, alt *testCrypto) func(*testing.T) {
// 	return func(t *testing.T) {
// 		// generate communication channel
// 		a2b := make(chan interface{})
// 		defer close(a2b)
// 		eg := &errgroup.Group{}
// 		const htlcDuration = 0
// 		const nOutputs = 2
// 		altCrypto, err := cryptos.ParseCrypto(alt.crypto)
// 		require.NoError(t, err, "can't parse alt")
// 		btcCrypto, err := cryptos.ParseCrypto("bitcoin")
// 		require.NoError(t, err, "can't parse bitcoin")
// 		// alice (alt)
// 		eg.Go(func() error {
// 			at, err := atomicswap.NewBuyerTrade(
// 				cctypes.Amount("1"),
// 				altCrypto,
// 				cctypes.Amount("1"),
// 				btcCrypto,
// 			)
// 			if err != nil {
// 				return err
// 			}
// 			pf := newPrintf(t.Logf, "alice (buyer)")
// 			err = handleTrade(pf, a2b, alt, btc, htlcDuration, at, true, nOutputs)
// 			if err != nil {
// 				return err
// 			}
// 			time.Sleep(htlcDuration * 2)
// 			return recoverFunds(alt, at, pf)
// 		})
// 		// bob (BTC)
// 		eg.Go(func() error {
// 			at, err := atomicswap.NewSellerTrade(
// 				cctypes.Amount("1"),
// 				btcCrypto,
// 				cctypes.Amount("1"),
// 				altCrypto,
// 			)
// 			if err != nil {
// 				return err
// 			}
// 			pf := newPrintf(t.Logf, "bob (seller)")
// 			err = handleTrade(pf, a2b, btc, alt, htlcDuration, at, true, nOutputs)
// 			if err != nil {
// 				return err
// 			}
// 			time.Sleep(htlcDuration * 2)
// 			return recoverFunds(btc, at, pf)
// 		})
// 		err = eg.Wait()
// 		require.NoError(t, err, "unexpected error")
// 	}
// }

// func TestAtomicSwapRecover(t *testing.T) {
// 	for _, i := range testCryptos[1:] {
// 		t.Run(i.crypto+"_bitcoin", newTestAtomicSwapRecover(testCryptos[0], i))
// 	}
// }

func TestTrade(t *testing.T) {

}
