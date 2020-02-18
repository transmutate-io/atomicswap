package atomicswap

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
	"transmutate.io/pkg/btccore"
)

var (
	btcClient = &btccore.Client{
		Address:  "bitcoin-core-testnet.docker:3333",
		Username: "admin",
		Password: "pass",
	}
	ltcClient = &btccore.Client{
		Address:  "litecoin-testnet.docker:2222",
		Username: "admin",
		Password: "pass",
	}
	btcMinerAddr, ltcMinerAddr string
)

func initTest(t *testing.T) {
	cmds := [][]string{
		{"kill", "bitcoin-core-testnet"},
		{"kill", "litecoin-testnet"},
		{"run", "--rm", "-d", "--name", "bitcoin-core-testnet", "bitcoin-core:0.19.0.1",
			"-rpcuser=admin", "-rpcpassword=pass",
			"-rpcbind=0.0.0.0:3333", "-rpcallowip=172.0.0.1/8",
			"-txindex", "-server", "-regtest",
		},
		{"run", "--rm", "-d", "--name", "litecoin-testnet", "litecoin:0.17.1",
			"litecoind",
			"-rpcuser=admin", "-rpcpassword=pass",
			"-rpcbind=0.0.0.0:2222", "-rpcallowip=172.0.0.1/8",
			"-txindex", "-server", "-regtest",
		},
	}
	// kill and restart containers
	for _, i := range cmds {
		exec.Command("docker", i...).CombinedOutput()
	}
	// sanity check
	bc, err := btcClient.GetBlockCount()
	require.NoError(t, err, "can't get block count for BTC")
	require.Zero(t, bc, "expecting 0 BTC blocks")
	bc, err = ltcClient.GetBlockCount()
	require.NoError(t, err, "can't get block count for LTC")
	require.Zero(t, bc, "expecting 0 LTC blocks")
	// generate miners addresses and miner 101 blocks
	btcMinerAddr, err = btcClient.GetNewAddress()
	require.NoError(t, err, "can't generate new BTC address")
	_, err = btcClient.GenerateToAddress(101, btcMinerAddr)
	require.NoError(t, err, "can't generate BTC for the miner")
	ltcMinerAddr, err = ltcClient.GetNewAddress()
	require.NoError(t, err, "can't generate new LTC address")
	_, err = ltcClient.GenerateToAddress(101, ltcMinerAddr)
	require.NoError(t, err, "can't generate LTC for the miner")
}

func TestAtomicSwap_BTC_LTC(t *testing.T) {
	initTest(t)

	// a2b := make(chan interface{})
	// eg := &errgroup.Group{}
	// // alice (LTC)
	// eg.Go(func() error {
	// 	require.True(t, false, "yolo it")
	// 	// generate a new LTC private key and fund it
	// 	ltcPriv, err := btcec.NewPrivateKey(btcec.S256())
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// generate a new token
	// 	token, err := readRandomToken()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// retrieve LTC private key from bob
	// 	ltcBobPub := (<-a2b).([]byte)
	// 	// generate a new BTC private key
	// 	btcPriv, err := btcec.NewPrivateKey(btcec.S256())
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// send BTC public key to bob
	// 	a2b <- btcPriv.PubKey().SerializeCompressed()
	// 	// generate htlc script
	// 	ltcLockScript := script.HTLC(
	// 		script.LockTimeTime(time.Now().UTC().Add(48*time.Hour)),
	// 		hash.Hash160(token),
	// 		script.P2PKHPublicBytes(ltcPriv.PubKey().SerializeCompressed()),
	// 		script.P2PKHPublicBytes(ltcBobPub),
	// 	)
	// 	// send htlc script to bob
	// 	a2b <- ltcLockScript
	// 	// generate htlc address
	// 	ltcDepositAddr, err := addr.P2SHFromScript(ltcLockScript, params.LTC_RegressionNet)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// fund deposit address
	// 	ltcDepositTxID, err := ltcClient.SendToAddress(ltcDepositAddr, (*btccore.Amount)(big.NewInt(1000000000)))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if _, err = ltcClient.GenerateToAddress(101, ltcMinerAddr); err != nil {
	// 		return err
	// 	}
	// 	t.Log("::::", ltcDepositTxID)
	// 	return nil
	// })
	// // bob (BTC)
	// eg.Go(func() error {
	// 	// generate a new LTC private key
	// 	ltcPriv, err := btcec.NewPrivateKey(btcec.S256())
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// send public key to alice
	// 	a2b <- ltcPriv.PubKey().SerializeCompressed()
	// 	// get key
	// 	btcAlicePub := (<-a2b).([]byte)
	// 	// get script
	// 	ltcLockScript := (<-a2b).([]byte)
	// 	_, _ = btcAlicePub, ltcLockScript
	// 	return nil
	// })
	// err := eg.Wait()
	// require.NoError(t, err, "unexpected error")
}
