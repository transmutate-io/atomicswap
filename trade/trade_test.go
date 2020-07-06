package trade

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/internal/testutil"
	"github.com/transmutate-io/atomicswap/key"
	"github.com/transmutate-io/cryptocore/types"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
)

func init() {
	for _, i := range testutil.Cryptos {
		if err := testutil.SetupMinerAddress(i); err != nil {
			panic(err)
		}
	}
}

func testTradeMarshalUnamarshal(t *testing.T, tc *testutil.Crypto) {
	ownCrypto := testutil.MustParseCrypto(t, testutil.Cryptos[0].Name)
	traderCrypto := testutil.MustParseCrypto(t, tc.Name)
	trade, err := NewOnChainBuy(
		types.Amount("1"), ownCrypto,
		types.Amount("1"), traderCrypto,
		48*time.Hour,
	)
	require.NoError(t, err, "can't create the buy")
	_, err = trade.GenerateToken()
	require.NoError(t, err, "can't generate token")
	err = trade.GenerateKeys()
	require.NoError(t, err, "can't generate keys")
	b, err := yaml.Marshal(trade)
	require.NoError(t, err, "can't marshal")
	trade2 := &OnChainTrade{baseTrade: &baseTrade{}}
	err = yaml.Unmarshal(b, trade2)
	require.NoError(t, err, "can't unmarshal")
	require.Equal(t, trade.Duration(), trade2.Duration(), "mismatch")
	require.Equal(t, trade.Token(), trade2.Token(), "mismatch")
	require.Equal(t, trade.TokenHash(), trade2.TokenHash(), "mismatch")
	requireTradeInfoEqual(t, trade.OwnInfo(), trade2.OwnInfo())
	requireTradeInfoEqual(t, trade.TraderInfo(), trade2.TraderInfo())
	require.Equal(t, trade.RedeemKey(), trade2.RedeemKey(), "redeem keys mismatch")
	require.Equal(t, trade.RecoveryKey(), trade2.RecoveryKey(), "recovery keys mismatch")
}

func requireTradeInfoEqual(t *testing.T, e, a *TraderInfo) {
	requireCryptoEqual(t, e.Crypto, a.Crypto)
	require.Equal(t, e.Amount, a.Amount)
}

func requireCryptoEqual(t *testing.T, e, a *cryptos.Crypto) {
	require.Equal(t, e.Name, a.Name, "name mismatch")
	require.Equal(t, e.Short, a.Short, "short name mismatch")
	require.Equal(t, e.Type, a.Type, "type mismatch")
}

func TestTradeMarshalUnamarshal(t *testing.T) {
	for _, i := range testutil.Cryptos[1:] {
		t.Run("bitcoin_"+i.Name, func(t *testing.T) {
			testTradeMarshalUnamarshal(t, i)
		})
	}
}

func newTestOnChainRedeemTrades(btc, alt *testutil.Crypto, htlcDuration time.Duration) (Trade, Trade, error) {
	// parse cryptos names
	altCrypto, err := cryptos.Parse(alt.Name)
	if err != nil {
		return nil, nil, err
	}
	btcCrypto, err := cryptos.Parse(btc.Name)
	if err != nil {
		return nil, nil, err
	}
	// set the buyer proposal values
	buy, err := NewOnChainBuy(
		types.Amount("1"), btcCrypto,
		types.Amount("1"), altCrypto,
		htlcDuration,
	)
	if err != nil {
		return nil, nil, err
	}
	return buy, NewOnChainSell(), nil
}

func testOnChainRedeem(t *testing.T, ownCrypto, traderCrypto *testutil.Crypto, dur time.Duration, buyerEx, sellerEx *testExchanger) func(*testing.T) {
	return func(t *testing.T) {
		buyerTrade, sellerTrade, err := newTestOnChainRedeemTrades(ownCrypto, traderCrypto, dur)
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

func TestOnChainRedeem(t *testing.T) {
	for _, i := range testutil.Cryptos[1:] {
		buyerEx, sellerEx := newTestExchangers(testutil.Cryptos[0], i, t.Logf)
		t.Run("bitcoin_"+i.Name, testOnChainRedeem(
			t,
			testutil.Cryptos[0],
			i,
			48*time.Hour,
			buyerEx,
			sellerEx,
		))
	}
}

func newTestOnChainRecover(btc, alt *testutil.Crypto, htlcDuration time.Duration, buyerEx, sellerEx *testExchanger) func(*testing.T) {
	return func(t *testing.T) {
		buyerTrade, sellerTrade, err := newTestOnChainRedeemTrades(btc, alt, htlcDuration)
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
			return recoverFunds(sellerTrade, alt, sellerEx.pf)
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
			return recoverFunds(buyerTrade, btc, buyerEx.pf)
		})
		err = eg.Wait()
		require.NoError(t, err, "unexpected error")
	}
}

func recoverFunds(t Trade, tc *testutil.Crypto, pf printfFunc) error {
	ld, err := t.RecoverableFunds().Lock().LockData()
	if err != nil {
		return err
	}
	if remDur := ld.Locktime.Add(time.Second).Sub(time.Now()); remDur > 0 {
		pf("waiting for %s\n", remDur)
		time.Sleep(remDur)
	}
	recKey, err := key.NewPrivate(t.OwnInfo().Crypto)
	if err != nil {
		return err
	}
	pf("new key generated\n")
	tx, err := t.RecoveryTx(recKey.Public().KeyData(), tc.FeePerByte)
	if err != nil {
		return err
	}
	b, err := tx.Serialize()
	if err != nil {
		return err
	}
	pf("tx: %s\n", hex.EncodeToString(b))
	txid, err := testutil.RetrySendRawTransaction(tc, b)
	if err != nil {
		return err
	}
	pf("txid: %s\n", hex.EncodeToString(txid))
	return nil
}

func TestOnChainRecover(t *testing.T) {
	for _, i := range testutil.Cryptos[1:] {
		buyerExchanger, sellerExchanger := newTestExchangers(testutil.Cryptos[0], i, t.Logf)
		t.Run("bitcoin_"+i.Name, newTestOnChainRecover(
			testutil.Cryptos[0],
			i,
			2*time.Second,
			buyerExchanger,
			sellerExchanger,
		))
	}
}
