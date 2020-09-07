package trade

import (
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/internal/testutil"
	"github.com/transmutate-io/atomicswap/key"
	"github.com/transmutate-io/atomicswap/script"
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
	trade, err := NewOnChainTrade(
		types.Amount("1"), ownCrypto,
		types.Amount("1"), traderCrypto,
		48*time.Hour,
	)
	require.NoError(t, err, "can't create the buy")
	btr, err := trade.Buyer()
	require.NoError(t, err, "can't get buyer trade")
	_, err = btr.GenerateToken()
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

func newBuyerTrade(own, trader *testutil.Crypto, dur time.Duration) (Trade, error) {
	// parse cryptos names
	ownCrypto, err := cryptos.Parse(own.Name)
	if err != nil {
		return nil, err
	}
	traderCrypto, err := cryptos.Parse(trader.Name)
	if err != nil {
		return nil, err
	}
	// set the buyer proposal values
	r, err := NewOnChainTrade(types.Amount("1"), ownCrypto, types.Amount("1"), traderCrypto, dur)
	if err != nil {
		return nil, err
	}
	return r, nil
}

var errTimedOut = errors.New("timed out")

func trySend(c chan interface{}, b interface{}) error {
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	select {
	case c <- b:
		return nil
	case <-timer.C:
		return errTimedOut
	}
}

func tryReceive(c chan interface{}) (interface{}, error) {
	timer := time.NewTimer(time.Minute)
	defer timer.Stop()
	select {
	case b := <-c:
		return b, nil
	case <-timer.C:
		return nil, errTimedOut
	}
}

func tryReceiveBuyProposal(a2b chan interface{}) (*BuyProposal, error) {
	v, err := tryReceive(a2b)
	if err != nil {
		return nil, err
	}
	r, ok := v.(*BuyProposal)
	if !ok {
		return nil, errors.New("not a buy proposal")
	}
	return r, nil
}

func tryReceiveLocks(a2b chan interface{}) (*Locks, error) {
	v, err := tryReceive(a2b)
	if err != nil {
		return nil, err
	}
	r, ok := v.(*Locks)
	if !ok {
		return nil, errors.New("not a lockset")
	}
	return r, nil
}

func tryReceiveOutputs(a2b chan interface{}) ([]*Output, error) {
	v, err := tryReceive(a2b)
	if err != nil {
		return nil, err
	}
	r, ok := v.([]*Output)
	if !ok {
		return nil, errors.New("not an *Output slice")
	}
	return r, nil
}

func tryReceiveBytes(a2b chan interface{}) (types.Bytes, error) {
	v, err := tryReceive(a2b)
	if err != nil {
		return nil, err
	}
	r, ok := v.(types.Bytes)
	if !ok {
		return nil, errors.New("not an *Output slice")
	}
	return r, nil
}

func lockFunds(t *testing.T, tr Trade, c *testutil.Crypto) ([]*Output, error) {
	// deposit address
	depositAddr, err := tr.RecoverableFunds().Lock().Address(c.Chain)
	if err != nil {
		return nil, err
	}
	if err = testutil.EnsureBalance(c, tr.OwnInfo().Amount); err != nil {
		return nil, err
	}
	t.Logf("deposit address: %s\n", depositAddr)
	txID, err := c.Client.SendToAddress(depositAddr, tr.OwnInfo().Amount)
	if err != nil {
		return nil, err
	}
	if _, err = testutil.GenerateBlocks(c, 1); err != nil {
		return nil, err
	}
	// find transaction
	tx, err := c.Client.Transaction(txID)
	if err != nil {
		return nil, err
	}
	t.Logf("deposit tx: %s\n", txID.Hex())
	// save recoverable output
	var found bool
	txUTXO, ok := tx.UTXO()
	if !ok {
		return nil, errors.New("not implemented")
	}
	for _, i := range txUTXO.Outputs() {
		addrs := i.LockScript().Addresses()
		if i.LockScript().Type() == "scripthash" &&
			len(addrs) > 0 &&
			addrs[0] == depositAddr {
			tr.RecoverableFunds().AddFunds(&Output{
				TxID:   txID,
				N:      uint32(i.N()),
				Amount: i.Value().UInt64(tr.OwnInfo().Crypto.Decimals),
			})
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("recoverable output not found")
	}
	t.Logf("funds locked: tx %s\n", txID.Hex())
	funds := tr.RecoverableFunds().Funds()
	r, ok := funds.([]*Output)
	if !ok {
		return nil, errors.New("not implemented")
	}
	return r, nil
}

func redeemFunds(t *testing.T, tr Trade, c *testutil.Crypto) error {
	destKey, err := key.NewPrivate(tr.OwnInfo().Crypto)
	if err != nil {
		return err
	}
	t.Logf("generated key: %s\n", destKey.Public().KeyData().Hex())
	gen, err := script.NewGenerator(tr.OwnInfo().Crypto)
	if err != nil {
		return err
	}
	rtx, err := tr.RedeemTx(gen.P2PKHHash(destKey.Public().KeyData()), c.FeePerByte)
	if err != nil {
		return err
	}
	b, err := rtx.Serialize()
	if err != nil {
		return err
	}
	t.Logf("generate redeem transaction: %s\n", hex.EncodeToString(b))
	if b, err = testutil.RetrySendRawTransaction(c, b); err != nil {
		return err
	}
	t.Logf("redeemed funds: %s\n", hex.EncodeToString(b))
	_, err = testutil.GenerateBlocks(c, 1)
	return err
}

func initBuyerTrade(t *testing.T, a2b chan interface{}, tr Trade, c *testutil.Crypto) error {
	// generate proposal and send
	btr, err := tr.Buyer()
	if err != nil {
		return err
	}
	prop, err := btr.GenerateBuyProposal()
	if err != nil {
		return err
	}
	if err = trySend(a2b, prop); err != nil {
		return err
	}
	// receive locks and accept
	ls, err := tryReceiveLocks(a2b)
	if err != nil {
		return err
	}
	if err = btr.SetLocks(ls); err != nil {
		return err
	}
	// lock funds and send info
	outputs, err := lockFunds(t, tr, c)
	if err != nil {
		return err
	}
	if err = trySend(a2b, outputs); err != nil {
		return err
	}
	// receive redeemable outputs
	if outputs, err = tryReceiveOutputs(a2b); err != nil {
		return err
	}
	funds := tr.RedeemableFunds()
	for _, i := range outputs {
		funds.AddFunds(i)
	}
	return nil
}

func initSellerTrade(t *testing.T, a2b chan interface{}, c *testutil.Crypto) (Trade, error) {
	// receive proposal
	prop, err := tryReceiveBuyProposal(a2b)
	if err != nil {
		return nil, err
	}
	// accept
	tr, err := AcceptProposal(prop)
	if err != nil {
		return nil, err
	}
	// generate locks and send
	str, err := tr.Seller()
	if err != nil {
		return nil, err
	}
	if err = trySend(a2b, str.Locks()); err != nil {
		return nil, err
	}
	// receive outputs
	outputs, err := tryReceiveOutputs(a2b)
	if err != nil {
		return nil, err
	}
	funds := tr.RedeemableFunds()
	for _, i := range outputs {
		funds.AddFunds(i)
	}
	// lock funds
	if outputs, err = lockFunds(t, tr, c); err != nil {
		return nil, err
	}
	if err = trySend(a2b, outputs); err != nil {
		return nil, err
	}
	// receive token
	token, err := tryReceiveBytes(a2b)
	if err != nil {
		return nil, err
	}
	tr.SetToken(token)
	return tr, nil
}

func testOnChainRedeem(t *testing.T, buyerCrypto, sellerCrypto *testutil.Crypto, dur time.Duration) func(*testing.T) {
	return func(t *testing.T) {
		a2b := make(chan interface{}, 0)
		eg := &errgroup.Group{}
		// bob, buyer, btc
		eg.Go(func() error {
			// new trade
			tr, err := newBuyerTrade(buyerCrypto, sellerCrypto, dur)
			if err != nil {
				return err
			}
			if err = initBuyerTrade(t, a2b, tr, buyerCrypto); err != nil {
				return err
			}
			// redeem
			if err = redeemFunds(t, tr, sellerCrypto); err != nil {
				return err
			}
			return trySend(a2b, tr.Token())
		})
		// alice, seller, alt
		eg.Go(func() error {
			tr, err := initSellerTrade(t, a2b, sellerCrypto)
			if err != nil {
				return err
			}
			// redeem
			if err = redeemFunds(t, tr, buyerCrypto); err != nil {
				return err
			}
			return nil
		})
		err := eg.Wait()
		require.NoError(t, err, "unexpected error")
	}
}

func TestOnChainRedeem(t *testing.T) {
	for _, i := range testutil.Cryptos[1:] {
		t.Run("bitcoin_"+i.Name, testOnChainRedeem(
			t,
			testutil.Cryptos[0],
			i,
			48*time.Hour,
		))
	}
}

func newTestOnChainRecover(buyerCrypto, sellerCrypto *testutil.Crypto, dur time.Duration) func(*testing.T) {
	return func(t *testing.T) {
		a2b := make(chan interface{}, 0)
		eg := &errgroup.Group{}
		// bob, buyer, btc
		eg.Go(func() error {
			// new trade
			tr, err := newBuyerTrade(buyerCrypto, sellerCrypto, dur)
			if err != nil {
				return err
			}
			if err = initBuyerTrade(t, a2b, tr, buyerCrypto); err != nil {
				return err
			}
			return recoverFunds(t, tr, buyerCrypto)
		})
		// alice, seller, alt
		eg.Go(func() error {
			tr, err := initSellerTrade(t, a2b, sellerCrypto)
			if err != nil {
				return err
			}
			return recoverFunds(t, tr, sellerCrypto)
		})
		err := eg.Wait()
		require.NoError(t, err, "unexpected error")
	}
}

func recoverFunds(t *testing.T, tr Trade, tc *testutil.Crypto) error {
	ld, err := tr.RecoverableFunds().Lock().LockData()
	if err != nil {
		return err
	}
	if remDur := ld.LockTime.Add(time.Second).Sub(time.Now()); remDur > 0 {
		t.Logf("waiting for %s\n", remDur)
		time.Sleep(remDur)
	}
	recKey, err := key.NewPrivate(tr.OwnInfo().Crypto)
	if err != nil {
		return err
	}
	t.Logf("new key generated\n")
	gen, err := script.NewGenerator(tr.OwnInfo().Crypto)
	if err != nil {
		return err
	}
	tx, err := tr.RecoveryTx(gen.P2PKHHash(recKey.Public().KeyData()), tc.FeePerByte)
	if err != nil {
		return err
	}
	b, err := tx.Serialize()
	if err != nil {
		return err
	}
	t.Logf("tx: %s\n", hex.EncodeToString(b))
	txid, err := testutil.RetrySendRawTransaction(tc, b)
	if err != nil {
		return err
	}
	t.Logf("txid: %s\n", hex.EncodeToString(txid))
	return nil
}

func TestOnChainRecover(t *testing.T) {
	for _, i := range testutil.Cryptos[1:] {
		t.Run("bitcoin_"+i.Name, newTestOnChainRecover(
			testutil.Cryptos[0],
			i,
			2*time.Second,
		))
	}
}
