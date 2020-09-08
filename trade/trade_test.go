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
	if b, err = testutil.SendRawTransaction(t, c, b, 1); err != nil {
		return err
	}
	t.Logf("redeemed funds: %s\n", hex.EncodeToString(b))
	_, err = testutil.GenerateBlocks(c, 1)
	return err
}

func testOnChainRedeem(t *testing.T, buyerCrypto, sellerCrypto *testutil.Crypto, dur time.Duration) func(*testing.T) {
	return func(t *testing.T) {
		// create buyer trade
		buyerTrade, err := newBuyerTrade(buyerCrypto, sellerCrypto, dur)
		require.NoError(t, err, "can't create buyer trade")
		// the buyer generates a buy proposal
		btr, err := buyerTrade.Buyer()
		require.NoError(t, err)
		prop, err := btr.GenerateBuyProposal()
		require.NoError(t, err, "can't generate buy proposal")
		// the seller accepts the proposal and the seller trade is created
		sellerTrade, err := AcceptProposal(prop)
		require.NoError(t, err, "can't accept proposal")
		// the seller generates a pair of locks and the buyer accepts them
		str, err := sellerTrade.Seller()
		require.NoError(t, err, "can't generate locks")
		err = btr.SetLocks(str.Locks())
		require.NoError(t, err, "can't accept locks")
		// the buyer locks his funds
		outputs, err := lockFunds(t, buyerTrade, buyerCrypto)
		require.NoError(t, err, "buyer can't lock funds")
		funds := sellerTrade.RedeemableFunds()
		for _, i := range outputs {
			funds.AddFunds(i)
		}
		// the seller locks his funds
		outputs, err = lockFunds(t, sellerTrade, sellerCrypto)
		require.NoError(t, err, "seller can't lock funds")
		funds = buyerTrade.RedeemableFunds()
		for _, i := range outputs {
			funds.AddFunds(i)
		}
		// the buyer redeems the funds, revealing the token
		err = redeemFunds(t, buyerTrade, sellerCrypto)
		require.NoError(t, err, "buyer can't redeem funds")
		sellerTrade.SetToken(buyerTrade.Token())
		// the seller redeems the funds, using the token
		err = redeemFunds(t, sellerTrade, buyerCrypto)
		require.NoError(t, err, "seller can't redeem funds")
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
		// create buyer trade
		buyerTrade, err := newBuyerTrade(buyerCrypto, sellerCrypto, dur)
		require.NoError(t, err, "can't create buyer trade")
		// the buyer generates a buy proposal
		btr, err := buyerTrade.Buyer()
		require.NoError(t, err)
		prop, err := btr.GenerateBuyProposal()
		require.NoError(t, err, "can't generate buy proposal")
		// the seller accepts the proposal and the seller trade is created
		sellerTrade, err := AcceptProposal(prop)
		require.NoError(t, err, "can't accept proposal")
		// the seller generates a pair of locks and the buyer accepts them
		str, err := sellerTrade.Seller()
		require.NoError(t, err, "can't generate locks")
		err = btr.SetLocks(str.Locks())
		require.NoError(t, err, "can't accept locks")
		// the buyer locks his funds
		_, err = lockFunds(t, buyerTrade, buyerCrypto)
		require.NoError(t, err, "buyer can't lock funds")
		// the seller locks his funds
		_, err = lockFunds(t, sellerTrade, sellerCrypto)
		require.NoError(t, err, "seller can't lock funds")
		// wait for the lock to expire
		time.Sleep(dur)
		// the buyer recovers his funds
		err = recoverFunds(t, buyerTrade, buyerCrypto)
		require.NoError(t, err, "the buyer can't recover funds")
		// the seller recovers his funds
		err = recoverFunds(t, sellerTrade, sellerCrypto)
		require.NoError(t, err, "the seller can't recover funds")
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
	txid, err := testutil.SendRawTransaction(t, tc, b, 1)
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
