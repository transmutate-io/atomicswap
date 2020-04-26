package atomicswap

import (
	"crypto/rand"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"transmutate.io/pkg/atomicswap/cryptos"
	"transmutate.io/pkg/atomicswap/key"
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
	a2b chan []byte
	pf  printfFunc
}

func newTestExchanger(a2b chan []byte, pf printfFunc) *testExchanger {
	return &testExchanger{a2b: a2b, pf: pf}
}

func (m *testExchanger) stageMap() StageHandlerMap { return StageHandlerMap{} }

func newGenerateKeysHandler(pf printfFunc) func(*Trade) error {
	return func(t *Trade) error {
		if err := t.GenerateKeys(); err != nil {
			return err
		}
		pf("generated keys\n")
		return nil
	}
}

func newGenerateTokenHandler(pf printfFunc) func(*Trade) error {
	return func(t *Trade) error {
		if _, err := t.GenerateToken(); err != nil {
			return err
		}
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

func newTestFundsData(t *testing.T, c *cryptos.Crypto) Funds {
	fd := newFunds(c)
	switch ffd := fd.(type) {
	case *fundsUTXO:
		txid1, err := readRandom(32)
		require.NoError(t, err, "can't read random bytes")
		nOut, err := rand.Int(rand.Reader, big.NewInt(5))
		require.NoError(t, err, "can't read random int")
		n := nOut.Uint64()
		ffd.Outputs = append(ffd.Outputs, &Output{
			TxID:   txid1,
			N:      uint32(n),
			Amount: n * 100000000,
		})
		ffd.LockScript = types.Bytes{}
	default:
		panic("not supported")
	}
	return fd
}

func newTestPrivateKey(t *testing.T, c *cryptos.Crypto) key.Private {
	p, err := key.NewPrivate(c)
	require.NoError(t, err, "can't create new key")
	return p
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
