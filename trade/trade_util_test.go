package trade

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/transmutate-io/atomicswap/internal/testutil"
	"github.com/transmutate-io/atomicswap/key"
	"github.com/transmutate-io/atomicswap/networks"
	"github.com/transmutate-io/atomicswap/roles"
	"github.com/transmutate-io/atomicswap/script"
	"github.com/transmutate-io/atomicswap/stages"
	"github.com/transmutate-io/cryptocore/tx"
	"gopkg.in/yaml.v2"
)

func newPrintf(oldPrintf printfFunc, name string) printfFunc {
	return func(f string, args ...interface{}) { oldPrintf(name+": "+f, args...) }
}

type printfFunc = func(f string, a ...interface{})

type testExchanger struct {
	a2b    chan []byte
	pf     printfFunc
	own    *testutil.Crypto
	trader *testutil.Crypto
}

func newTestExchanger(a2b chan []byte, pf printfFunc, own, trader *testutil.Crypto) *testExchanger {
	return &testExchanger{
		a2b:    a2b,
		pf:     pf,
		own:    own,
		trader: trader,
	}
}

func newTestExchangers(buyer, seller *testutil.Crypto, pf printfFunc) (*testExchanger, *testExchanger) {
	a2b := make(chan []byte, 0)
	return newTestExchanger(
			a2b,
			newPrintf(pf, "bob, buyer, "+buyer.Name),
			buyer,
			seller,
		),
		newTestExchanger(
			a2b,
			newPrintf(pf, "alice, seller, "+seller.Name),
			seller,
			buyer,
		)
}

func buyProposalString(p *BuyProposal) string {
	return fmt.Sprintf(
		"%s %s (%s) for %s %s (%s)",
		p.Buyer.Amount,
		p.Buyer.Crypto,
		time.Duration(p.Buyer.LockDuration).String(),
		p.Seller.Amount,
		p.Seller.Crypto,
		time.Duration(p.Seller.LockDuration).String(),
	)
}

var ErrTimedOut = errors.New("timed out")

func trySend(c chan []byte, b []byte) bool {
	timer := time.NewTimer(time.Minute)
	defer timer.Stop()
	select {
	case c <- b:
		return true
	case <-timer.C:
		return false
	}
}

func (m *testExchanger) sendProposal(t Trade) error {
	bt, err := t.Buyer()
	if err != nil {
		return err
	}
	prop, err := bt.GenerateBuyProposal()
	if err != nil {
		return err
	}
	b, err := yaml.Marshal(prop)
	if err != nil {
		return err
	}
	if !trySend(m.a2b, b) {
		return ErrTimedOut
	}
	m.pf("proposal sent: %s\n", buyProposalString(prop))
	return nil
}

func tryReceive(c chan []byte) ([]byte, bool) {
	timer := time.NewTimer(time.Minute)
	defer timer.Stop()
	select {
	case b := <-c:
		return b, true
	case <-timer.C:
		return nil, false
	}
}

func (m *testExchanger) receiveProposal(t Trade) error {
	b, ok := tryReceive(m.a2b)
	if !ok {
		return ErrTimedOut
	}
	prop, err := UnamrshalBuyProposal(b)
	if err != nil {
		return err
	}
	m.pf("got a proposal: %s\n", buyProposalString(prop))
	st, err := t.Seller()
	if err != nil {
		return err
	}
	return st.AcceptBuyProposal(prop)
}

func (m *testExchanger) sendProposalResponse(t Trade) error {
	st, err := t.Seller()
	if err != nil {
		return err
	}
	resp := st.Locks()
	b, err := yaml.Marshal(resp)
	if err != nil {
		return err
	}
	if !trySend(m.a2b, b) {
		return ErrTimedOut
	}
	m.pf("proposal response sent: %s\n", buyProposalResponseString(resp))
	return nil
}

func buyProposalResponseString(r *BuyProposalResponse) string {
	return fmt.Sprintf(
		"buyer lock: %s, seller lock: %s",
		r.Buyer.Bytes().Hex(),
		r.Seller.Bytes().Hex(),
	)
}

func (m *testExchanger) receiveProposalResponse(t Trade) error {
	b, ok := tryReceive(m.a2b)
	if !ok {
		return ErrTimedOut
	}
	resp, err := UnamrshalBuyProposalResponse(t.OwnInfo().Crypto, t.TraderInfo().Crypto, b)
	if err != nil {
		m.pf("can't unmarshal err: %#v\n", err)
		return err
	}
	m.pf("proposal response received: %s\n", buyProposalResponseString(resp))
	bt, err := t.Buyer()
	if err != nil {
		return err
	}
	return bt.SetLocks(resp)
}

func (m *testExchanger) lockFunds(t Trade) error {
	// deposit address
	depositAddr, err := t.RecoverableFunds().Lock().Address(m.own.Chain)
	if err != nil {
		return err
	}
	if err = testutil.EnsureBalance(m.own, t.OwnInfo().Amount); err != nil {
		return err
	}
	m.pf("deposit address: %s\n", depositAddr)
	txID, err := m.own.Client.SendToAddress(depositAddr, t.OwnInfo().Amount)
	if err != nil {
		return err
	}
	if _, err = testutil.GenerateBlocks(m.own, m.own.ConfirmBlocks+1); err != nil {
		return err
	}
	// find transaction
	tx, err := m.own.Client.Transaction(txID)
	if err != nil {
		return err
	}
	m.pf("deposit tx: %s\n", txID.Hex())
	// save recoverable output
	var found bool
	txUTXO, ok := tx.UTXO()
	if !ok {
		panic("not implemented")
	}
	for _, i := range txUTXO.Outputs() {
		addrs := i.LockScript().Addresses()
		if i.LockScript().Type() == "scripthash" &&
			len(addrs) > 0 &&
			addrs[0] == depositAddr {
			t.RecoverableFunds().AddFunds(&Output{
				TxID:   txID,
				N:      uint32(i.N()),
				Amount: i.Value().UInt64(t.OwnInfo().Crypto.Decimals),
			})
			found = true
			break
		}
	}
	if !found {
		return errors.New("recoverable output not found")
	}
	m.pf("funds locked: tx %s\n", txID.Hex())
	if !trySend(m.a2b, txID) {
		return ErrTimedOut
	}
	return nil
}

func stringsContains(ss []string, s string) bool {
	for _, i := range ss {
		if s == i {
			return true
		}
	}
	return false
}

func (m *testExchanger) waitLockedFunds(t Trade) error {
	b, ok := tryReceive(m.a2b)
	if !ok {
		return ErrTimedOut
	}
	tx, err := m.trader.Client.Transaction(b)
	if err != nil {
		return err
	}
	depositAddr, err := networks.AllByName[m.trader.Name][m.trader.Chain].
		P2SHFromScript(t.RedeemableFunds().Lock().Bytes())
	if err != nil {
		return err
	}
	txUTXO, ok := tx.UTXO()
	if !ok {
		panic("not implemented")
	}
	for _, i := range txUTXO.Outputs() {
		if !stringsContains(i.LockScript().Addresses(), depositAddr) {
			continue
		}
		out := &Output{
			TxID:   tx.ID(),
			N:      uint32(i.N()),
			Amount: i.Value().UInt64(t.TraderInfo().Crypto.Decimals),
		}
		t.RedeemableFunds().AddFunds(out)
		m.pf("redeemable output found: %s, %d %d\n", out.TxID.Hex(), out.N, out.Amount)
	}
	return nil
}

func (m *testExchanger) redeem(t Trade) error {
	destKey, err := key.NewPrivate(t.OwnInfo().Crypto)
	if err != nil {
		return err
	}
	m.pf("generated key: %s\n", destKey.Public().KeyData().Hex())
	gen, err := script.NewGenerator(t.OwnInfo().Crypto)
	if err != nil {
		return err
	}
	rtx, err := t.RedeemTx(gen.P2PKHHash(destKey.Public().KeyData()), m.trader.FeePerByte)
	if err != nil {
		return err
	}
	b, err := rtx.Serialize()
	if err != nil {
		return err
	}
	m.pf("generate redeem transaction: %s\n", hex.EncodeToString(b))
	if b, err = testutil.RetrySendRawTransaction(m.trader, b); err != nil {
		return err
	}
	m.pf("redeemed funds: %s\n", hex.EncodeToString(b))
	if t.Role() == roles.Buyer {
		if !trySend(m.a2b, b) {
			return ErrTimedOut
		}
	}
	return nil
}

func (m *testExchanger) waitFundsRedeem(t Trade) error {
	b, ok := tryReceive(m.a2b)
	if !ok {
		return ErrTimedOut
	}
	tx, err := m.own.Client.Transaction(b)
	if err != nil {
		return err
	}
	token, err := extractToken(tx)
	if err != nil {
		return err
	}
	t.SetToken(token)
	m.pf("redeemed transaction found, token: %s\n", hex.EncodeToString(token))
	return nil
}

var ErrInputNotFound = errors.New("input not found")

func extractToken(tx tx.Tx) ([]byte, error) {
	txUTXO, ok := tx.UTXO()
	if !ok {
		panic("not implemented")
	}
	for _, i := range txUTXO.Inputs() {
		inst := strings.Split(i.UnlockScript().Asm(), " ")
		if len(inst) != 5 || inst[3] != "0" {
			continue
		}
		b, err := hex.DecodeString(inst[2])
		if err != nil {
			continue
		}
		return b, nil
	}
	return nil, ErrInputNotFound
}

func (m *testExchanger) redeemStageMap() StageHandlerMap {
	return StageHandlerMap{
		stages.WaitLockedFunds:         m.waitLockedFunds,
		stages.LockFunds:               m.lockFunds,
		stages.WaitFundsRedeemed:       m.waitFundsRedeem,
		stages.RedeemFunds:             m.redeem,
		stages.SendProposal:            m.sendProposal,
		stages.ReceiveProposal:         m.receiveProposal,
		stages.SendProposalResponse:    m.sendProposalResponse,
		stages.ReceiveProposalResponse: m.receiveProposalResponse,
		stages.GenerateToken:           newGenerateTokenHandler(m.pf),
		stages.GenerateKeys:            newGenerateKeysHandler(m.pf),
		stages.Done:                    newDoneHandler(m.pf),
	}
}

func stageNOP(_ Trade) error { return nil }

func (m *testExchanger) receiveNOP(_ Trade) error {
	if _, ok := tryReceive(m.a2b); ok {
		return nil
	}
	return ErrTimedOut
}

func (m *testExchanger) recoverStageMap() StageHandlerMap {
	return StageHandlerMap{
		stages.WaitLockedFunds:         m.receiveNOP,
		stages.WaitFundsRedeemed:       stageNOP,
		stages.RedeemFunds:             stageNOP,
		stages.LockFunds:               m.lockFunds,
		stages.SendProposal:            m.sendProposal,
		stages.ReceiveProposal:         m.receiveProposal,
		stages.SendProposalResponse:    m.sendProposalResponse,
		stages.ReceiveProposalResponse: m.receiveProposalResponse,
		stages.GenerateToken:           newGenerateTokenHandler(m.pf),
		stages.GenerateKeys:            newGenerateKeysHandler(m.pf),
		stages.Done:                    newDoneHandler(m.pf),
	}
}

func newGenerateKeysHandler(pf printfFunc) func(Trade) error {
	return func(t Trade) error {
		if err := t.GenerateKeys(); err != nil {
			return err
		}
		pf("redeem key: %s %s\n",
			t.TraderInfo().Crypto.String(),
			t.RedeemKey().Public().KeyData().Hex(),
		)
		pf("recovery key: %s %s\n",
			t.OwnInfo().Crypto.String(),
			t.RecoveryKey().Public().KeyData().Hex(),
		)
		pf("generated keys\n")
		return nil
	}
}

func newGenerateTokenHandler(pf printfFunc) func(Trade) error {
	return func(t Trade) error {
		bt, err := t.Buyer()
		if err != nil {
			return err
		}
		if _, err := bt.GenerateToken(); err != nil {
			return err
		}
		pf("token: %s\n", t.Token().Hex())
		pf("token hash: %s\n", t.TokenHash().Hex())
		pf("generated token\n")
		return nil
	}
}

func newDoneHandler(pf printfFunc) func(Trade) error {
	return func(t Trade) error {
		pf("trade done\n")
		return nil
	}
}
