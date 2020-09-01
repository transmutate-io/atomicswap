package cmds

import (
	"bytes"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/atomicswap/hash"
	"github.com/transmutate-io/atomicswap/internal/cmdutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil/exitcodes"
	"github.com/transmutate-io/atomicswap/script"
	"github.com/transmutate-io/atomicswap/trade"
	"github.com/transmutate-io/cryptocore"
	"github.com/transmutate-io/cryptocore/block"
	"github.com/transmutate-io/cryptocore/tx"
	"github.com/transmutate-io/cryptocore/types"
)

type newClientFunc func(addr, user, pass string, tlsConf *cryptocore.TLSConfig) (cryptocore.Client, error)

var newClientFuncs = map[string]newClientFunc{
	cryptos.Bitcoin.Name:     cryptocore.NewClientBTC,
	cryptos.Litecoin.Name:    cryptocore.NewClientLTC,
	cryptos.Dogecoin.Name:    cryptocore.NewClientDOGE,
	cryptos.Decred.Name:      cryptocore.NewClientDCR,
	cryptos.BitcoinCash.Name: cryptocore.NewClientBCH,
}

func newClient(
	c *cryptos.Crypto,
	address string,
	username string,
	password string,
	tlsConf *cryptocore.TLSConfig,
) (cryptocore.Client, error) {
	nc, ok := newClientFuncs[c.Name]
	if !ok {
		return nil, errors.New("client unavailable")
	}
	return nc(address, username, password, tlsConf)
}

func mustNewclient(
	c *cryptos.Crypto,
	address string,
	username string,
	password string,
	tlsConf *cryptocore.TLSConfig,
) cryptocore.Client {
	r, err := newClient(c, address, username, password, tlsConf)
	if err != nil {
		cmdutil.ErrorExit(exitcodes.UnknownCrypto, c.Name)
	}
	return r
}

type blockWatchData struct {
	Bottom uint64 `yaml:"bottom"`
	Top    uint64 `yaml:"top"`
}

type watchData struct {
	Own    *blockWatchData `yaml:"own"`
	Trader *blockWatchData `yaml:"trader"`
}

func getBlockAtHeight(cl cryptocore.Client, height uint64) (block.Block, error) {
	bh, err := cl.BlockHash(height)
	if err != nil {
		return nil, err
	}
	return cl.Block(bh)
}

var errClosed = errors.New("closed")

func blockTransactions(cl cryptocore.Client, txs []types.Bytes, closec chan struct{}) ([]tx.Tx, error) {
	r := make([]tx.Tx, 0, len(txs))
	for _, i := range txs {
		select {
		case <-closec:
			return nil, errClosed
		default:
		}
		tx, err := cl.Transaction(i)
		if err != nil {
			return nil, err
		}
		r = append(r, tx)
	}
	return r, nil
}

type blockData struct {
	height uint64
	txs    []tx.Tx
}

const (
	initTimeout = time.Second
	maxTimeout  = time.Minute
)

func iterateBlocks(cl cryptocore.Client, wd *blockWatchData, stopBottom uint64) (chan *blockData, chan error, func()) {
	closec := make(chan struct{}, 0)
	bdc := make(chan *blockData, 0)
	errc := make(chan error, 1)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	closeIter := func() {
		close(closec)
		wg.Wait()
		close(errc)
		close(bdc)
	}
	blockCount, err := cl.BlockCount()
	if err != nil {
		errc <- err
		return bdc, errc, closeIter
	}
	go func() {
		defer wg.Done()
		var (
			height        uint64
			nextBlockHash []byte
		)
		if wd.Top == 0 {
			height = blockCount
		} else {
			height = wd.Top + 1
		}
		timeout := initTimeout
		for {
			var (
				block block.Block
				err   error
			)
			if nextBlockHash != nil {
				block, err = cl.Block(nextBlockHash)
			} else {
				block, err = getBlockAtHeight(cl, height)
			}
			if err != nil {
				if err == cryptocore.ErrNoBlock {
					timeout *= 2
					if timeout > maxTimeout {
						timeout = maxTimeout
					}
					timer := time.NewTimer(timeout)
					select {
					case <-closec:
						timer.Stop()
						return
					case <-timer.C:
						timer.Stop()
						continue
					}
				}
				errc <- err
				return
			}
			timeout = initTimeout
			blockTxs, err := blockTransactions(cl, block.Transactions(), closec)
			if err != nil {
				if err == errClosed {
					return
				}
				errc <- err
				return
			}
			select {
			case <-closec:
				return
			case bdc <- &blockData{
				height: uint64(block.Height()),
				txs:    blockTxs,
			}:
				height++
				nextBlockHash = block.NextBlockHash()
			}
		}
	}()
	go func() {
		defer wg.Done()
		var (
			height        uint64
			prevBlockHash []byte
		)
		if wd.Bottom == 0 {
			height = blockCount - 1
		} else {
			height = wd.Bottom - 1
		}
		for {
			if height < stopBottom {
				return
			}
			var (
				block block.Block
				err   error
			)
			if prevBlockHash == nil {
				block, err = getBlockAtHeight(cl, height)
			} else {
				block, err = cl.Block(prevBlockHash)
			}
			if err != nil {
				errc <- err
				return
			}
			blockTxs, err := blockTransactions(cl, block.Transactions(), closec)
			if err != nil {
				if err == errClosed {
					return
				}
				errc <- err
				return
			}
			select {
			case <-closec:
				return
			case bdc <- &blockData{
				height: uint64(block.Height()),
				txs:    blockTxs,
			}:
				prevBlockHash = block.PreviousBlockHash()
				height = uint64(block.Height()) - 1
			}
			if height+1 == stopBottom {
				return
			}
		}
	}()
	return bdc, errc, closeIter
}

func extractToken(c *cryptos.Crypto, t tx.Tx, lock trade.Lock) (types.Bytes, error) {
	txUtxo, ok := t.UTXO()
	if !ok {
		panic("not implemented")
	}
	ld, err := lock.LockData()
	if err != nil {
		return nil, err
	}
	for _, j := range txUtxo.Inputs() {
		if j.Coinbase() != nil {
			continue
		}
		dis, err := script.DisassembleStrings(c, j.UnlockScript().Bytes())
		if err != nil {
			continue
		}
		if len(dis) != 5 {
			continue
		}
		if dis[3] != "0" {
			continue
		}
		h, err := hash.New(c)
		if err != nil {
			return nil, err
		}
		b, err := hex.DecodeString(dis[1])
		if err != nil {
			continue
		}
		hb := h.Hash160(b)
		if !bytes.Equal(ld.RedeemKeyData, hb) {
			continue
		}
		if b, err = hex.DecodeString(dis[4]); err != nil {
			continue
		}
		if !bytes.Equal(lock.Bytes(), b) {
			continue
		}
		if b, err = hex.DecodeString(dis[2]); err != nil {
			continue
		}
		return b, nil
	}
	return nil, nil
}
