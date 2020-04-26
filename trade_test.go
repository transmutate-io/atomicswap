package atomicswap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/roles"
	"transmutate.io/pkg/atomicswap/stages"
	"transmutate.io/pkg/cryptocore"
	"transmutate.io/pkg/cryptocore/types"
)

var testCryptos = []*testCrypto{
	{
		"bitcoin",
		cryptocore.NewClientBTC(
			envOr("GO_TEST_BTC_NODE", "bitcoin-core-regtest.docker:4444"),
			"admin", "pass", false,
		),
		"",
		types.Amount("50"),
		8,
	},
	{
		"litecoin",
		cryptocore.NewClientLTC(
			envOr("GO_TEST_LTC_NODE", "litecoin-regtest.docker:4444"),
			"admin", "pass", false,
		),
		"",
		types.Amount("50"),
		8,
	},
	{
		"dogecoin",
		cryptocore.NewClientDOGE(
			envOr("GO_TEST_DOGE_NODE", "dogecoin-regtest.docker:4444"),
			"admin", "pass", false,
		),
		"",
		types.Amount("50"),
		8,
	},
	{
		"bitcoin-cash",
		cryptocore.NewClientBTCCash(
			envOr("GO_TEST_BCH_NODE", "bitcoin-cash-regtest.docker:4444"),
			"admin", "pass", false,
		),
		"",
		types.Amount("50"),
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

func TestAtomicSwapRedeem(t *testing.T) {
	for _, i := range testCryptos[1:] {
		t.Run("bitcoin_"+i.crypto,
			newTestAtomicSwapRedeem(testCryptos[0], i, 48*time.Hour),
		)
	}
}

func newTestHandlers(pf printfFunc) StageHandlerMap {
	return StageHandlerMap{
		stages.GenerateKeys:  newGenerateKeysHandler(pf),
		stages.GenerateToken: newGenerateTokenHandler(pf),
		stages.Done:          newDoneHandler(pf),
	}
}

func newTestAtomicSwapRedeem(btc, alt *testCrypto, htlcDuration time.Duration) func(*testing.T) {
	return func(t *testing.T) {
		// generate communication channel
		a2b := make(chan []byte, 0)
		defer close(a2b)
		// parse cryptos names
		altCrypto, err := cryptos.Parse(alt.crypto)
		require.NoError(t, err, "can't parse alt crypto")
		btcCrypto, err := cryptos.Parse(btc.crypto)
		require.NoError(t, err, "can't parse btc crypto")
		// set the buyer proposal values
		buyerTrade := NewTrade().
			WithRole(roles.Buyer).
			WithDuration(htlcDuration).
			WithStages(DefaultTradeStages[roles.Buyer]...).
			WithOwnAmountCrypto(types.Amount("1"), btcCrypto).
			WithTraderAmountCrypto(types.Amount("1"), altCrypto)
		sellerTrade := NewTrade().
			WithRole(roles.Seller).
			WithStages(DefaultTradeStages[roles.Seller]...)
		eg := &errgroup.Group{}
		// alice (alt)
		eg.Go(func() error {
			// generate-lock wait-locked-funds lock-funds wait-funds-redeemed redeem
			pf := newPrintf(t.Logf, "alice ("+alt.crypto+")")
			me := newTestExchanger(a2b, pf)
			tradeHandler := NewStageHandler(me.stageMap())
			tradeHandler.InstallHandlers(newTestHandlers(pf))
			return tradeHandler.HandleTrade(sellerTrade)
		})
		// bob (BTC)
		eg.Go(func() error {
			// generate-lock wait-locked-funds lock-funds redeem
			pf := newPrintf(t.Logf, "bob ("+btc.crypto+")")
			me := newTestExchanger(a2b, pf)
			tradeHandler := NewStageHandler(me.stageMap())
			tradeHandler.InstallHandlers(newTestHandlers(pf))
			return tradeHandler.HandleTrade(buyerTrade)
		})
		err = eg.Wait()
		require.NoError(t, err, "unexpected error")
	}
}

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
// 		a2b <- types.Bytes(hash.Hash160(at.Own.RedeemKey.Public().SerializeCompressed()))
// 		pf("sent public key hash\n")
// 	case stages.ReceivePublicKeyHash:
// 		// use a channel to exchange data back and forth
// 		at.Trader.RedeemKeyHash = (<-a2b).(types.Bytes)
// 		pf("received public key hash: %s\n", at.Trader.RedeemKeyHash.Hex())
// 	case stages.ShareTokenHash:
// 		// use a channel to exchange data back and forth
// 		a2b <- at.TokenHash()
// 		pf("sent token hash\n")
// 	case stages.ReceiveTokenHash:
// 		// use a channel to exchange data back and forth
// 		at.SetTokenHash((<-a2b).(types.Bytes))
// 		pf("received token hash: %s\n", at.TokenHash().Hex())
// 	case stages.ReceiveLockScript:
// 		// use a channel to exchange data back and forth
// 		ls := (<-a2b).(types.Bytes)
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
// 			txID, err := ownCrypto.cl.SendToAddress(depositAddr, types.NewAmount(at.Own.Amount.UInt64(ownCrypto.decimals), 8))
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

func newTestAtomicSwapRecover(btc, alt *testCrypto) func(*testing.T) {
	return func(t *testing.T) {
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
		// 				types.Amount("1"),
		// 				altCrypto,
		// 				types.Amount("1"),
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
		// 				types.Amount("1"),
		// 				btcCrypto,
		// 				types.Amount("1"),
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
	}
}

func TestAtomicSwapRecover(t *testing.T) {
	for _, i := range testCryptos[1:] {
		t.Run("bitcoin_"+i.crypto, newTestAtomicSwapRecover(testCryptos[0], i))
	}
}

func TestTradeMarshalUnamarshal(t *testing.T) {
	for _, i := range testCryptos[1:] {
		t.Run("bitcoin_"+i.crypto, func(t *testing.T) {
			ownCrypto, err := cryptos.Parse(testCryptos[0].crypto)
			require.NoError(t, err, "can't parse coin name")
			traderCrypto, err := cryptos.Parse(i.crypto)
			require.NoError(t, err, "can't parse coin name")
			redeemKey := newTestPrivateKey(t, traderCrypto)
			recoveryKey := newTestPrivateKey(t, ownCrypto)
			trade := &Trade{
				Role:     roles.Buyer,
				Duration: Duration(48 * time.Hour),
				OwnInfo: &TraderInfo{
					Crypto: ownCrypto,
					Amount: "1",
				},
				TraderInfo: &TraderInfo{
					Crypto: traderCrypto,
					Amount: "1",
				},
				RedeemKey:        redeemKey,
				RecoveryKey:      recoveryKey,
				RedeemableFunds:  newTestFundsData(t, traderCrypto),
				RecoverableFunds: newTestFundsData(t, ownCrypto),
				Stages:           stages.NewStager(stages.Done),
			}
			token, err := readRandomToken()
			require.NoError(t, err, "can't read random token")
			trade.SetToken(token)
			b, err := yaml.Marshal(trade)
			require.NoError(t, err, "can't marshal")
			trade2 := &Trade{}
			err = yaml.Unmarshal(b, trade2)
			require.NoError(t, err, "can't unmarshal")
			require.Equal(t, trade.Duration, trade2.Duration, "mismatch")
			require.Equal(t, trade.Token, trade2.Token, "mismatch")
			require.Equal(t, trade.TokenHash, trade2.TokenHash, "mismatch")
			requireTradeInfoEqual(t, trade.OwnInfo, trade2.OwnInfo)
			requireTradeInfoEqual(t, trade.TraderInfo, trade2.TraderInfo)
			require.Equal(t, trade.RedeemKey, trade2.RedeemKey, "redeem keys mismatch")
			require.Equal(t, trade.RecoveryKey, trade2.RecoveryKey, "recovery keys mismatch")
			require.Equal(t, trade.RedeemableFunds, trade2.RedeemableFunds, "redeemable funds mismatch")
			require.Equal(t, trade.RecoverableFunds, trade2.RecoverableFunds, "recoverable funds mismatch")
		})
	}
}
