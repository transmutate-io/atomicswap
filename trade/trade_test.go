package trade

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/key"
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
		buyerExchanger, sellerExchanger := newTestExchangers(testCryptos[0], i, t.Logf)
		t.Run("bitcoin_"+i.crypto,
			newTestAtomicSwapRedeem(
				testCryptos[0],
				i,
				48*time.Hour,
				buyerExchanger,
				sellerExchanger,
			),
		)
	}
}

func newTestExchangers(buyer, seller *testCrypto, pf printfFunc) (*testExchanger, *testExchanger) {
	a2b := make(chan []byte, 0)
	return newTestExchanger(a2b, newPrintf(pf, "bob, buyer, "+buyer.crypto), buyer, seller),
		newTestExchanger(a2b, newPrintf(pf, "alice, seller, "+seller.crypto), seller, buyer)
}

func newTestAtomicSwapTrades(btc, alt *testCrypto, htlcDuration time.Duration) (*Trade, *Trade, error) {
	// parse cryptos names
	altCrypto, err := cryptos.Parse(alt.crypto)
	if err != nil {
		return nil, nil, err
	}
	btcCrypto, err := cryptos.Parse(btc.crypto)
	if err != nil {
		return nil, nil, err
	}
	// set the buyer proposal values
	return NewBuy(
		types.Amount("1"), btcCrypto,
		types.Amount("1"), altCrypto,
		htlcDuration,
	), NewSell(), nil
}

func newTestAtomicSwapRedeem(btc, alt *testCrypto, htlcDuration time.Duration, buyerEx, sellerEx *testExchanger) func(*testing.T) {
	return func(t *testing.T) {
		buyerTrade, sellerTrade, err := newTestAtomicSwapTrades(btc, alt, htlcDuration)
		require.NoError(t, err, "can't create trades")
		eg := &errgroup.Group{}
		// alice, seller, alt
		eg.Go(func() error {
			b, err := yaml.Marshal(sellerTrade)
			if err != nil {
				return err
			}
			if err = yaml.Unmarshal(b, sellerTrade); err != nil {
				return err
			}
			return NewHandler(sellerEx.redeemStageMap()).HandleTrade(sellerTrade)
		})
		// bob, buyer, btc
		eg.Go(func() error {
			b, err := yaml.Marshal(buyerTrade)
			if err != nil {
				return err
			}
			if err = yaml.Unmarshal(b, buyerTrade); err != nil {
				return err
			}
			return NewHandler(buyerEx.redeemStageMap()).HandleTrade(buyerTrade)
		})
		err = eg.Wait()
		require.NoError(t, err, "unexpected error")
	}
}

func TestAtomicSwapRecover(t *testing.T) {
	for _, i := range testCryptos[1:] {
		buyerExchanger, sellerExchanger := newTestExchangers(testCryptos[0], i, t.Logf)
		t.Run("bitcoin_"+i.crypto, newTestAtomicSwapRecover(
			testCryptos[0],
			i,
			2*time.Second,
			buyerExchanger,
			sellerExchanger,
		))
	}
}

func newTestAtomicSwapRecover(btc, alt *testCrypto, htlcDuration time.Duration, buyerEx, sellerEx *testExchanger) func(*testing.T) {
	return func(t *testing.T) {
		buyerTrade, sellerTrade, err := newTestAtomicSwapTrades(btc, alt, htlcDuration)
		require.NoError(t, err, "can't create trades")
		eg := &errgroup.Group{}
		// alice, seller, alt
		eg.Go(func() error {
			b, err := yaml.Marshal(sellerTrade)
			if err != nil {
				return err
			}
			if err = yaml.Unmarshal(b, sellerTrade); err != nil {
				return err
			}
			if err = NewHandler(sellerEx.recoverStageMap()).HandleTrade(sellerTrade); err != nil {
				return err
			}
			return recover(sellerTrade, alt.cl, alt.minerAddr, sellerEx.pf)
		})
		// bob, buyer, btc
		eg.Go(func() error {
			b, err := yaml.Marshal(buyerTrade)
			if err != nil {
				return err
			}
			if err = yaml.Unmarshal(b, buyerTrade); err != nil {
				return err
			}
			if err = NewHandler(buyerEx.recoverStageMap()).HandleTrade(buyerTrade); err != nil {
				return err
			}
			return recover(buyerTrade, btc.cl, btc.minerAddr, buyerEx.pf)
		})
		err = eg.Wait()
		require.NoError(t, err, "unexpected error")
	}
}

func recover(t *Trade, cl cryptocore.Client, minerAddr string, pf printfFunc) error {
	ld, err := t.RecoverableFunds.Lock().LockData()
	if err != nil {
		return err
	}
	if remDur := ld.Locktime.Add(time.Second).Sub(time.Now()); remDur > 0 {
		pf("waiting for %s\n", remDur)
		time.Sleep(remDur)
	}
	recKey, err := key.NewPrivate(t.OwnInfo.Crypto)
	if err != nil {
		return err
	}
	pf("new key generated\n")
	tx, err := t.RecoveryTx(recKey.Public().KeyData(), stdFeePerByte)
	if err != nil {
		return err
	}
	b, err := tx.Serialize()
	if err != nil {
		return err
	}
	pf("tx: %s\n", hex.EncodeToString(b))
	txid, err := cl.SendRawTransaction(b)
	if err != nil {
		return err
	}
	if err = generateToAddress(cl, minerAddr, 101); err != nil {
		return err
	}
	pf("txid: %s\n", txid.Hex())
	return nil
}

func TestTradeMarshalUnamarshal(t *testing.T) {
	for _, i := range testCryptos[1:] {
		t.Run("bitcoin_"+i.crypto, func(t *testing.T) {
			ownCrypto, err := cryptos.Parse(testCryptos[0].crypto)
			require.NoError(t, err, "can't parse coin name")
			traderCrypto, err := cryptos.Parse(i.crypto)
			require.NoError(t, err, "can't parse coin name")
			trade := NewBuy(
				types.Amount("1"), ownCrypto,
				types.Amount("1"), traderCrypto,
				48*time.Hour,
			)
			_, err = trade.GenerateToken()
			require.NoError(t, err, "can't generate token")
			err = trade.GenerateKeys()
			require.NoError(t, err, "can't generate keys")
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
		})
	}
}
