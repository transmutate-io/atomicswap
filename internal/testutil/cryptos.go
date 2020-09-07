package testutil

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/params"
	"github.com/transmutate-io/cryptocore"
	"github.com/transmutate-io/cryptocore/tx"
	"github.com/transmutate-io/cryptocore/types"
)

type Crypto struct {
	Name          string
	Chain         params.Chain
	Client        cryptocore.Client
	MinerAddr     string
	MatureBlocks  int
	ConfirmBlocks int
	FeePerByte    uint64
}

func envOr(envName string, defaultValue string) string {
	e, ok := os.LookupEnv(envName)
	if !ok {
		return defaultValue
	}
	return e
}

var Cryptos = []*Crypto{
	{
		Name:  "bitcoin",
		Chain: params.RegressionNet,
		Client: MustNewClient(
			cryptocore.NewClientBTC,
			envOr("GO_TEST_BTC", "bitcoin-core-localnet.docker:4444"),
			"admin",
			"pass",
			nil),
		MinerAddr:     "",
		MatureBlocks:  101,
		ConfirmBlocks: 6,
		FeePerByte:    2,
	},
	{
		Name:  "litecoin",
		Chain: params.RegressionNet,
		Client: MustNewClient(
			cryptocore.NewClientLTC,
			envOr("GO_TEST_LTC", "litecoin-localnet.docker:4444"),
			"admin",
			"pass",
			nil,
		),
		MinerAddr:     "",
		MatureBlocks:  101,
		ConfirmBlocks: 6,
		FeePerByte:    2,
	},
	{
		Name:  "dogecoin",
		Chain: params.RegressionNet,
		Client: MustNewClient(
			cryptocore.NewClientDOGE,
			envOr("GO_TEST_DOGE", "dogecoin-localnet.docker:4444"),
			"admin",
			"pass",
			nil,
		),
		MinerAddr:     "",
		MatureBlocks:  61,
		ConfirmBlocks: 92,
		FeePerByte:    2,
	},
	{
		Name:  "bitcoin-cash",
		Chain: params.RegressionNet,
		Client: MustNewClient(
			cryptocore.NewClientBCH,
			envOr("GO_TEST_BCH", "bitcoin-cash-localnet.docker:4444"),
			"admin",
			"pass",
			nil,
		),
		MinerAddr:     "",
		MatureBlocks:  0,
		ConfirmBlocks: 17,
		FeePerByte:    2,
	},
	{
		Name:  "decred",
		Chain: params.SimNet,
		Client: MustNewClient(
			cryptocore.NewClientDCR,
			envOr("GO_TEST_DCR", "decred-wallet-localnet.docker:4444"),
			"admin",
			"pass",
			&cryptocore.TLSConfig{SkipVerify: true},
		),
		MinerAddr:     "",
		MatureBlocks:  40,
		ConfirmBlocks: 20,
		FeePerByte:    10,
	},
}

func MustNewClient(newClFn func(addr, user, pass string, tlsConf *cryptocore.TLSConfig) (cryptocore.Client, error), addr, user, pass string, tlsConf *cryptocore.TLSConfig) cryptocore.Client {
	c, err := newClFn(addr, user, pass, tlsConf)
	if err != nil {
		panic(err)
	}
	return c
}

func MustParseCrypto(t *testing.T, c string) *cryptos.Crypto {
	r, err := cryptos.Parse(c)
	require.NoError(t, err, "can't parse crypto")
	return r
}

func SetupMinerAddress(tc *Crypto) error {
	if tc.MinerAddr != "" {
		return nil
	}
	funds, err := tc.Client.ReceivedByAddress(0, true, nil)
	if err != nil {
		return err
	}
	if len(funds) == 0 {
		addr, err := tc.Client.NewAddress()
		if err != nil {
			return err
		}
		tc.MinerAddr = addr
	} else {
		tc.MinerAddr = funds[0].Address
	}
	return nil
}

func EnsureBalance(tc *Crypto, bal types.Amount) error {
	crypto, err := cryptos.Parse(tc.Name)
	if err != nil {
		return err
	}
	bb := bal.UInt64(crypto.Decimals)
	bc, err := tc.Client.BlockCount()
	if err != nil {
		return err
	}
	if mb := uint64(tc.MatureBlocks); bc < mb {
		if _, err = GenerateBlocks(tc, int(mb-bc)); err != nil {
			return err
		}
	}
	for {
		b, err := tc.Client.Balance(0)
		if err != nil {
			return err
		}
		if b.UInt64(crypto.Decimals) > bb {
			return nil
		}
		if _, err = GenerateBlocks(tc, 1); err != nil {
			return err
		}
	}
}

func MustEnsureBalance(t *testing.T, tc *Crypto, bal types.Amount) {
	require.NoError(t, EnsureBalance(tc, bal), "can't ensure balance")
}

func GenerateBlocks(tc *Crypto, n int) ([]types.Bytes, error) {
	if n <= 0 {
		return []types.Bytes{}, nil
	}
	if tc.Client.CanGenerateBlocksToAddress() {
		return tc.Client.GenerateBlocksToAddress(n, tc.MinerAddr)
	} else {
		return tc.Client.GenerateBlocks(n)
	}
}

func MustGenerateBlocks(t *testing.T, tc *Crypto, n int) []types.Bytes {
	r, err := GenerateBlocks(tc, n)
	require.NoError(t, err, "can't generate blocks")
	return r
}

func FindIdx(tc *Crypto, txid []byte, addr string) (tx.Output, error) {
	tx, err := tc.Client.Transaction(txid)
	if err != nil {
		return nil, err
	}
	txUTXO, ok := tx.UTXO()
	if !ok {
		return nil, errors.New("not implemented")
	}
	for _, out := range txUTXO.Outputs() {
		for _, j := range out.LockScript().Addresses() {
			if j == addr {
				return out, nil
			}
		}
	}
	return nil, errors.New("not found")
}

func FindIdxs(tc *Crypto, txids [][]byte, addr string) ([]tx.Output, error) {
	r := make([]tx.Output, 0, len(txids))
	for _, i := range txids {
		out, err := FindIdx(tc, i, addr)
		if err != nil {
			return nil, err
		}
		r = append(r, out)
	}
	return r, nil
}

func MustFindIdxs(t *testing.T, tc *Crypto, txids [][]byte, addr string) []tx.Output {
	r, err := FindIdxs(tc, txids, addr)
	require.NoError(t, err, "can't find indexes")
	return r
}

func RetrySendRawTransaction(tc *Crypto, tx []byte, nBlocks int) (r []byte, err error) {
	for {
		if r, err = tc.Client.SendRawTransaction((tx)); err != nil {
			if err != cryptocore.ErrNonFinal {
				return
			}
			if _, err = GenerateBlocks(tc, nBlocks); err != nil {
				return
			}
		}
		break
	}
	_, err = GenerateBlocks(tc, nBlocks)
	return
}

func MustRetrySendRawTransaction(t *testing.T, tc *Crypto, tx []byte, nBlocks int) []byte {
	r, err := RetrySendRawTransaction(tc, tx, nBlocks)
	require.NoError(t, err, "can't retry sending raw transaction")
	return r
}
