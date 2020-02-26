package atomicswap_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
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
	"transmutate.io/pkg/atomicswap/params"
	"transmutate.io/pkg/atomicswap/script"
	"transmutate.io/pkg/atomicswap/types"
	"transmutate.io/pkg/atomicswap/types/stages"
	"transmutate.io/pkg/btccore"
)

var (
	btcClient = &btccore.Client{
		Address:  envOr("BTC_HOST", "bitcoin-core-testnet.docker:3333"),
		Username: "admin",
		Password: "pass",
	}
	ltcClient = &btccore.Client{
		Address:  envOr("LTC_HOST", "litecoin-testnet.docker:2222"),
		Username: "admin",
		Password: "pass",
	}
	btcMinerAddr, ltcMinerAddr string
)

func setupMinerAddress(cl *btccore.Client, minerAddr *string) error {
	if minerAddr == nil {
		return errors.New("need a miner address")
	}
	if *minerAddr != "" {
		return nil
	}
	// find existing addresses
	funds, err := cl.ListReceivedByAddress(0, true, false, nil)
	if err != nil {
		return err
	}
	if len(funds) > 0 {
		*minerAddr = funds[0].Address
	} else {
		// generate new address
		addr, err := cl.GetNewAddress()
		if err != nil {
			return err
		}
		*minerAddr = addr
	}
	return nil
}

func setupMiner(cl *btccore.Client, minerAddr *string, minValue int64) error {
	if err := setupMinerAddress(cl, minerAddr); err != nil {
		return err
	}
	if err := generateFunds(cl, minerAddr, minValue); err != nil {
		return err
	}
	return nil
}

func init() {
	if err := setupMiner(btcClient, &btcMinerAddr, 500000000); err != nil {
		panic(err)
	}
	if err := setupMiner(ltcClient, &ltcMinerAddr, 5000000000); err != nil {
		panic(err)
	}
}

const stdFee = 1000

func handleTrade(c chan interface{}, t *atomicswap.Trade, printf printfFunc, failAfterLock bool, contractDuration time.Duration) error {
	var (
		ownCl, traderCl               *btccore.Client
		ownCp, traderCp               *params.Params
		ownMinerAddr, traderMinerAddr string
	)
	if t.Own.Crypto == params.Bitcoin {
		ownCl = btcClient
		ownCp = params.BTC_RegressionNet
		ownMinerAddr = btcMinerAddr
		traderCl = ltcClient
		traderCp = params.LTC_RegressionNet
		traderMinerAddr = ltcMinerAddr
	} else {
		ownCl = ltcClient
		ownCp = params.LTC_RegressionNet
		ownMinerAddr = ltcMinerAddr
		traderCl = btcClient
		traderCp = params.BTC_RegressionNet
		traderMinerAddr = btcMinerAddr
	}

	for {
		switch t.Stage {
		case stages.SendPublicKeyHash:
			// use a channel to exchange data back and forth
			c <- types.Bytes(t.Own.RedeemKey.Public().Hash160())
			printf("sent public key hash\n")
			t.NextStage()
		case stages.ReceivePublicKeyHash:
			// use a channel to exchange data back and forth
			t.Trader.RedeemKeyHash = (<-c).(types.Bytes)
			printf("received public key hash: %s\n", t.Trader.RedeemKeyHash.Hex())
			t.NextStage()
		case stages.SendTokenHash:
			// use a channel to exchange data back and forth
			c <- t.TokenHash()
			printf("sent token hash\n")
			t.NextStage()
		case stages.ReceiveTokenHash:
			// use a channel to exchange data back and forth
			t.SetTokenHash((<-c).(types.Bytes))
			printf("received token hash: %s\n", t.TokenHash().Hex())
			t.NextStage()
		case stages.ReceiveLockScript:
			// use a channel to exchange data back and forth
			ls := (<-c).(types.Bytes)
			ds, err := script.DisassembleString(ls)
			if err != nil {
				return err
			}
			// check lock script
			if err := t.CheckTraderLockScript(ls); err != nil {
				printf("received invalid lock script: %s %s\n", ls.Hex(), ds)
				return err
			}
			printf("received lock script: %s %s\n", ls.Hex(), ds)
			// save lock script after checking
			t.Trader.LockScript = ls
			t.NextStage()
		case stages.GenerateLockScript:
			// generate lock script
			if err := t.GenerateOwnLockScript(); err != nil {
				return err
			}
			printf("generated lock script: %s\n", t.Own.LockScript.Hex())
			t.NextStage()
		case stages.SendLockScript:
			// use a channel to exchange data back and forth
			c <- t.Own.LockScript
			printf("sent lock script\n")
			t.NextStage()
		case stages.WaitLockTransaction:
			// use the client to find a deposit
			txOut, err := waitDeposit(traderCl, traderCp, t.Trader.LastBlockHeight, t.Trader.LockScript.Hash160())
			if err != nil {
				return err
			}
			printf(
				"found lock transaction: %s, %d (block %d)\n",
				hex.EncodeToString(txOut.txID),
				txOut.n,
				txOut.blockHeight,
			)
			// save redeeamable output
			if t.Outputs == nil {
				t.Outputs = &atomicswap.Outputs{}
			}
			t.Outputs.Redeemable = &atomicswap.Output{
				TxID: txOut.txID,
				N:    txOut.n,
			}
			t.Trader.LastBlockHeight = txOut.blockHeight
			t.NextStage()
		case stages.LockFunds:
			// use the client to make a deposit
			var amt *btccore.Amount
			if t.Own.Crypto == params.Bitcoin {
				amt = (*btccore.Amount)(big.NewInt(100000000 + stdFee))
			} else {
				amt = (*btccore.Amount)(big.NewInt(1000000000 + stdFee))
			}
			// calculate deposit address
			depositAddr, err := addr.P2SH(t.Own.LockScript.Hash160(), ownCp)
			if err != nil {
				return err
			}
			txID, err := ownCl.SendToAddress(depositAddr, amt)
			if err != nil {
				return err
			}
			if _, err = ownCl.GenerateToAddress(101, ownMinerAddr); err != nil {
				return err
			}
			b, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}
			// find transaction
			tx, err := ownCl.GetRawTransaction(b)
			if err != nil {
				return err
			}
			if t.Outputs == nil {
				t.Outputs = &atomicswap.Outputs{}
			}
			// save recoverable output
			for _, i := range tx.VOut {
				if i.ScriptPubKey.Type == "scripthash" && len(i.ScriptPubKey.Addresses) > 0 && i.ScriptPubKey.Addresses[0] == depositAddr {
					t.Outputs.Recoverable = &atomicswap.Output{TxID: b, N: uint32(i.N)}
					break
				}
			}
			printf("funds locked: tx %s\n", txID)
			t.NextStage()
			if failAfterLock {
				return nil
			}
		case stages.WaitRedeemTransaction:
			// use the client to find the trader redeem transaction and extract token
			token, err := waitRedeem(ownCl, ownCp, t.Own.LastBlockHeight, t.Outputs.Recoverable.TxID, int(t.Outputs.Recoverable.N))
			if err != nil {
				return err
			}
			printf("redeem transaction found: token %s\n", hex.EncodeToString(token))
			// save token
			t.SetToken(token)
			t.NextStage()
		case stages.RedeemFunds:
			// create a raw transaction and use the client to send it to redeem the funds
			var amt int64
			if t.Trader.Crypto == params.Bitcoin {
				amt = 100000000
			} else {
				amt = 1000000000
			}
			redeemTx, err := t.RedeemTransaction(amt)
			b, err := redeemTx.Serialize()
			if err != nil {
				return err
			}
			txID, err := traderCl.SendRawTransaction(b, nil)
			if err != nil {
				return err
			}
			if _, err = traderCl.GenerateToAddress(101, traderMinerAddr); err != nil {
				return err
			}
			printf("funds redeemed: tx %s\n", txID)
			t.NextStage()
		case stages.Done:
			printf("stage: %s\n", stages.Done)
			return nil
		default:
			return errors.New("invalid stage")
		}
	}
}

func bytesJoin(b ...[]byte) []byte { return bytes.Join(b, []byte{}) }

var errClosed = errors.New("closed")

func blockIterator(cl *btccore.Client, startBlockHeight int) (func() (*btccore.Block, error), func()) {
	cc := make(chan struct{})
	errc := make(chan error, 1)
	blkc := make(chan *btccore.Block)
	go func() {
		defer close(blkc)
		defer close(errc)
		var (
			bh  btccore.HexBytes
			err error
		)
		for {
			if bh == nil {
				bh, err = cl.GetBlockHash(startBlockHeight)
				if err != nil {
					e, ok := err.(*btccore.WalletError)
					if !ok || e.Code != -8 {
						errc <- err
						return
					}
					time.Sleep(time.Second)
					continue
				}
			}
			blk, err := cl.GetBlock(bh, true)
			if err != nil {
				errc <- err
				return
			}
			bh = blk.NextBlockHash
			startBlockHeight++
			var blockSent bool
			for !blockSent {
				select {
				case blkc <- blk:
					blockSent = true
				case <-cc:
					return
				}
			}
		}
	}()
	return func() (*btccore.Block, error) {
			select {
			case err := <-errc:
				return nil, err
			case blk := <-blkc:
				return blk, nil
			case <-cc:
				return nil, errClosed
			}
		},
		func() {
			close(cc)
		}
}

type txOut struct {
	blockHeight int
	txID        []byte
	n           uint32
}

func waitDeposit(cl *btccore.Client, chainParams *params.Params, startBlockHeight int, scriptHash []byte) (*txOut, error) {
	depositAddr, err := addr.P2SH(scriptHash, chainParams)
	if err != nil {
		return nil, err
	}
	next, close := blockIterator(cl, startBlockHeight)
	defer close()
	for {
		blk, err := next()
		if err != nil {
			return nil, err
		}
		startBlockHeight++
		for _, i := range blk.Transactions.Raw() {
			for _, j := range i.VOut {
				if j.ScriptPubKey.Type != "scripthash" {
					continue
				}
				if len(j.ScriptPubKey.Addresses) < 1 {
					continue
				}
				if j.ScriptPubKey.Addresses[0] != depositAddr {
					continue
				}
				return &txOut{blockHeight: startBlockHeight, txID: i.ID, n: uint32(j.N)}, nil
			}
		}

	}
}

func waitRedeem(cl *btccore.Client, chainParams *params.Params, startBlockHeight int, txID []byte, idx int) ([]byte, error) {
	next, close := blockIterator(cl, startBlockHeight)
	defer close()
	for {
		blk, err := next()
		if err != nil {
			return nil, err
		}
		for _, i := range blk.Transactions.Raw() {
			for _, j := range i.VIn {
				if !bytes.Equal(j.TransactionID, txID) || j.VOut != idx {
					continue
				}
				inst := strings.Split(j.ScriptSig.Asm, " ")
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
}

func fmtPrintf(f string, a ...interface{}) { fmt.Printf(f, a...) }

type printfFunc = func(f string, a ...interface{})

func newPrintf(oldPrintf printfFunc, name string) printfFunc {
	return func(f string, args ...interface{}) { oldPrintf(name+": "+f, args...) }
}

func envOr(envName string, defaultValue string) string {
	e, ok := os.LookupEnv(envName)
	if !ok {
		return defaultValue
	}
	return e
}

func generateFunds(cl *btccore.Client, minerAddr *string, minValue int64) error {
	for {
		// check balance
		amt, err := cl.GetBalance(0, false, nil)
		if err != nil {
			return err
		}
		if amt.BigInt().Int64() >= minValue {
			break
		}
		_, err = cl.GenerateToAddress(1, *minerAddr)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestAtomicSwap_BTC_LTC_Redeem(t *testing.T) {
	// generate communication channel
	a2b := make(chan interface{})
	defer close(a2b)
	eg := &errgroup.Group{}
	htlcDuration := 48 * time.Hour
	// alice (LTC)
	eg.Go(func() error {
		at, err := atomicswap.NewSellerTrade(params.Litecoin, params.Bitcoin)
		if err != nil {
			return err
		}
		pf := newPrintf(t.Logf, "alice (seller)")
		err = handleTrade(a2b, at, pf, false, htlcDuration)
		if err != nil {
			return err
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
	})
	// bob (BTC)
	eg.Go(func() error {
		at, err := atomicswap.NewBuyerTrade(params.Bitcoin, params.Litecoin)
		if err != nil {
			return err
		}
		pf := newPrintf(t.Logf, "bob (buyer)")
		err = handleTrade(a2b, at, pf, false, htlcDuration)
		if err != nil {
			return err
		}
		if err != nil {
			return err
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
	})
	err := eg.Wait()
	require.NoError(t, err, "unexpected error")
}

func recoverFunds(trade *atomicswap.Trade, printf printfFunc) error {
	var (
		ownCl        *btccore.Client
		ownMinerAddr string
		ownAmount    int64
	)
	if trade.Own.Crypto == params.Bitcoin {
		ownCl = btcClient
		ownMinerAddr = btcMinerAddr
		ownAmount = 100000000
	} else {
		ownCl = ltcClient
		ownMinerAddr = ltcMinerAddr
		ownAmount = 1000000000
	}
	tx, err := trade.RecoveryTransaction(ownAmount - stdFee)
	if err != nil {
		return err
	}
	b, err := tx.Serialize()
	if err != nil {
		return err
	}
	printf("generate recovery transaction: %s\n", hex.EncodeToString(b))
	txID, err := ownCl.SendRawTransaction(b, nil)
	if err != nil {
		return err
	}
	printf("funds recovered: tx: %s\n", txID)
	_, err = ownCl.GenerateToAddress(1, ownMinerAddr)
	return err
}

func TestAtomicSwap_BTC_LTC_Recover(t *testing.T) {
	// generate communication channel
	a2b := make(chan interface{})
	defer close(a2b)
	eg := &errgroup.Group{}
	const htlcDuration = 0
	// alice (LTC)
	eg.Go(func() error {
		at, err := atomicswap.NewSellerTrade(params.Litecoin, params.Bitcoin)
		if err != nil {
			return err
		}
		pf := newPrintf(t.Logf, "alice (seller)")
		if err = handleTrade(a2b, at, pf, true, htlcDuration); err != nil {
			return err
		}
		time.Sleep(htlcDuration * 2)
		return recoverFunds(at, pf)
	})
	// bob (BTC)
	eg.Go(func() error {
		at, err := atomicswap.NewBuyerTrade(params.Bitcoin, params.Litecoin)
		if err != nil {
			return err
		}
		pf := newPrintf(t.Logf, "bob (buyer)")
		if err = handleTrade(a2b, at, pf, true, htlcDuration); err != nil {
			return err
		}
		time.Sleep(htlcDuration)
		return recoverFunds(at, pf)
	})
	err := eg.Wait()
	require.NoError(t, err, "unexpected error")
}

func getLastBlock(cl *btccore.Client) (btccore.HexBytes, error) {
	bc, err := cl.GetBlockCount()
	if err != nil {
		return nil, err
	}
	bh, err := cl.GetBlockHash(int(bc))
	if err != nil {
		return nil, err
	}
	blk, err := cl.GetBlockHex(bh)
	if err != nil {
		return nil, err
	}
	return blk, nil
}
