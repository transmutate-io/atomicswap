package trade

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/hash"
	"transmutate.io/pkg/atomicswap/key"
	"transmutate.io/pkg/atomicswap/networks"
	"transmutate.io/pkg/atomicswap/params"
	"transmutate.io/pkg/atomicswap/stages"
	"transmutate.io/pkg/cryptocore"
	"transmutate.io/pkg/cryptocore/types"
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
	amount    types.Amount
	decimals  int
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

type testExchanger struct {
	a2b    chan []byte
	pf     printfFunc
	own    *testCrypto
	trader *testCrypto
}

func newTestExchanger(a2b chan []byte, pf printfFunc, own, trader *testCrypto) *testExchanger {
	return &testExchanger{
		a2b:    a2b,
		pf:     pf,
		own:    own,
		trader: trader,
	}
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

func (m *testExchanger) sendProposal(t *Trade) error {
	prop, err := t.GenerateBuyProposal()
	if err != nil {
		return err
	}
	b, err := yaml.Marshal(prop)
	if err != nil {
		return err
	}
	m.a2b <- b
	m.pf("proposal sent: %s\n", buyProposalString(prop))
	return nil
}

func (m *testExchanger) receiveProposal(t *Trade) error {
	prop, err := UnamrshalBuyProposal(<-m.a2b)
	if err != nil {
		return err
	}
	m.pf("got a proposal: %s\n", buyProposalString(prop))
	return t.AcceptBuyProposal(prop)
}

func (m *testExchanger) sendProposalResponse(t *Trade) error {
	resp := &BuyProposalResponse{
		Buyer:  t.RedeemableFunds.Lock(),
		Seller: t.RecoverableFunds.Lock(),
	}
	b, err := yaml.Marshal(resp)
	if err != nil {
		return err
	}
	m.a2b <- b
	m.pf("proposal response sent: %s\n", buyProposalResponseString(resp))
	return nil
}

func buyProposalResponseString(r *BuyProposalResponse) string {
	return fmt.Sprintf(
		"buyer lock: %s (%s), seller lock: %s (%s)",
		r.Buyer.Bytes().Hex(),
		hex.EncodeToString(hash.Hash160(r.Buyer.Bytes())),
		r.Seller.Bytes().Hex(),
		hex.EncodeToString(hash.Hash160(r.Seller.Bytes())),
	)
}

func (m *testExchanger) receiveProposalResponse(t *Trade) error {
	resp, err := UnamrshalBuyProposalResponse(t.OwnInfo.Crypto, t.TraderInfo.Crypto, <-m.a2b)
	if err != nil {
		m.pf("can't unmarshal err: %#v\n", err)
		return err
	}
	t.SetLocks(resp)
	m.pf("proposal response received: %s\n", buyProposalResponseString(resp))
	return nil
}

func (m *testExchanger) waitLockedFunds(t *Trade) error {
	lf, err := waitDeposit(
		m.trader,
		networks.RegressionByName[t.TraderInfo.Crypto.Name],
		1,
		hash.Hash160(t.RedeemableFunds.Lock().Bytes()),
	)
	if err != nil {
		return err
	}
	t.RedeemableFunds.AddFunds(&lf.Output)
	m.pf("redeemable output found: %d %s, %d %d\n", lf.blockHeight, lf.Output.TxID.Hex(), lf.Output.N, lf.Output.Amount)
	return nil
}

func generateToAddress(cl cryptocore.Client, minerAddr string, nBlocks int) error {
	if nBlocks > 0 {
		if _, err := cl.GenerateToAddress(nBlocks, minerAddr); err != nil {
			return err
		}
	}
	return nil
}

func sendToAddress(cl cryptocore.Client, addr, minerAddr string, amount types.Amount, nBlocks int) (types.Bytes, error) {
	r, err := cl.SendToAddress(addr, amount)
	if err != nil {
		return nil, err
	}
	if err = generateToAddress(cl, minerAddr, nBlocks); err != nil {
		return nil, err
	}
	return r, nil
}

func (m *testExchanger) lockFunds(t *Trade) error {
	// deposit address
	depositAddr, err := t.RecoverableFunds.Lock().Address(t.OwnInfo.Crypto, params.RegressionNet)
	if err != nil {
		return err
	}
	m.pf("deposit address: %s\n", depositAddr)
	txID, err := sendToAddress(m.own.cl, depositAddr, m.own.minerAddr, t.OwnInfo.Amount, 101)
	if err != nil {
		return err
	}
	// find transaction
	tx, err := m.own.cl.Transaction(txID)
	if err != nil {
		return err
	}
	m.pf("deposit tx: %s\n", txID.Hex())
	// save recoverable output
	var found bool
	for _, i := range tx.Outputs {
		if i.UnlockScript.Type == "scripthash" &&
			len(i.UnlockScript.Addresses) > 0 &&
			i.UnlockScript.Addresses[0] == depositAddr {
			t.RecoverableFunds.AddFunds(&Output{
				TxID:   txID,
				N:      uint32(i.N),
				Amount: i.Value.UInt64(m.own.decimals),
			})
			found = true
			break
		}
	}
	if !found {
		return errors.New("recoverable output not found")
	}
	m.pf("funds locked: tx %s\n", txID.Hex())
	return nil
}

func (m *testExchanger) waitFundsRedeem(t *Trade) error {
	outputs := t.RecoverableFunds.Funds().([]*Output)
	token, err := waitRedeem(
		m.own.cl,
		networks.RegressionByName[m.own.crypto],
		1,
		outputs[0].TxID,
		int(outputs[0].N),
	)
	if err != nil {
		return err
	}
	t.SetToken(token)
	m.pf("redeemed transaction found, token: %s\n", hex.EncodeToString(token))
	return nil
}

const stdFeePerByte = 2

func sendRawTransaction(cl cryptocore.Client, tx []byte, minerAddr string, nBlocks int) ([]byte, error) {
	r, err := cl.SendRawTransaction(tx)
	if err != nil {
		return nil, err
	}
	if err = generateToAddress(cl, minerAddr, nBlocks); err != nil {
		return nil, err
	}
	return r, nil
}

func (m *testExchanger) redeem(t *Trade) error {
	destKey, err := key.NewPrivate(t.OwnInfo.Crypto)
	if err != nil {
		return err
	}
	m.pf("generated key: %s\n", destKey.Public().KeyData().Hex())
	rtx, err := t.RedeemTransaction(destKey.Public().KeyData(), stdFeePerByte)
	if err != nil {
		return err
	}
	b, err := rtx.Serialize()
	if err != nil {
		return err
	}
	m.pf("generate redeem transaction: %s\n", hex.EncodeToString(b))
	if b, err = sendRawTransaction(m.trader.cl, b, m.trader.minerAddr, 101); err != nil {
		return err
	}
	m.pf("redeemed funds: %s\n", hex.EncodeToString(b))
	return nil
}

func (m *testExchanger) stageMap() StageHandlerMap {
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

func newGenerateKeysHandler(pf printfFunc) func(*Trade) error {
	return func(t *Trade) error {
		if err := t.GenerateKeys(); err != nil {
			return err
		}
		pf("redeem key: %s %s\n", t.TraderInfo.Crypto.String(), t.RedeemKey.Public().KeyData().Hex())
		pf("recovery key: %s %s\n", t.OwnInfo.Crypto.String(), t.RecoveryKey.Public().KeyData().Hex())
		pf("generated keys\n")
		return nil
	}
}

func newGenerateTokenHandler(pf printfFunc) func(*Trade) error {
	return func(t *Trade) error {
		if _, err := t.GenerateToken(); err != nil {
			return err
		}
		pf("token: %s\n", t.Token.Hex())
		pf("token hash: %s\n", t.TokenHash.Hex())
		pf("generated token\n")
		return nil
	}
}

func newDoneHandler(pf printfFunc) func(*Trade) error {
	return func(t *Trade) error {
		pf("trade done\n")
		return nil
	}
}

// func fmtPrintf(f string, a ...interface{}) { fmt.Printf(f, a...) }

type printfFunc = func(f string, a ...interface{})

func newPrintf(oldPrintf printfFunc, name string) printfFunc {
	return func(f string, args ...interface{}) { oldPrintf(name+": "+f, args...) }
}

func requireCryptoEqual(t *testing.T, e, a *cryptos.Crypto) {
	require.Equal(t, e.Name, a.Name, "name mismatch")
	require.Equal(t, e.Short, a.Short, "short name mismatch")
	require.Equal(t, e.Type, a.Type, "type mismatch")
}

func requireTradeInfoEqual(t *testing.T, e, a *TraderInfo) {
	requireCryptoEqual(t, e.Crypto, a.Crypto)
	require.Equal(t, e.Amount, a.Amount)
}

type txOut struct {
	Output
	blockHeight uint64
}

func waitDeposit(crypto *testCrypto, chainParams params.Params, startBlockHeight uint64, scriptHash []byte) (*txOut, error) {
	depositAddr, err := chainParams.P2SH(scriptHash)
	if err != nil {
		return nil, err
	}
	next, closeIter := cryptocore.NewBlockIterator(crypto.cl, startBlockHeight)
	defer closeIter()
	for {
		blk, err := next()
		if err != nil {
			return nil, err
		}
		for _, i := range blk.Transactions {
			tx, err := crypto.cl.Transaction(i)
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
				return &txOut{
					blockHeight: startBlockHeight,
					Output: Output{
						TxID:   tx.ID,
						N:      uint32(j.N),
						Amount: j.Value.UInt64(crypto.decimals),
					},
				}, nil
			}
		}
		startBlockHeight++
	}
}

func waitRedeem(cl cryptocore.Client, chainParams params.Params, startBlockHeight uint64, txID []byte, idx int) ([]byte, error) {
	next, closeIter := cryptocore.NewTransactionIterator(cl, startBlockHeight)
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
			if len(inst) != 5 || inst[3] != "0" {
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
