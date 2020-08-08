package cmds

import (
	"sync"
	"time"

	"github.com/spf13/pflag"
	"github.com/transmutate-io/atomicswap/cryptos"
	"github.com/transmutate-io/cryptocore"
	"github.com/transmutate-io/cryptocore/block"
	"github.com/transmutate-io/cryptocore/tx"
	"github.com/transmutate-io/cryptocore/types"
)

type newClientFunc func(addr, user, pass string, tlsConf *cryptocore.TLSConfig) cryptocore.Client

var newClientFuncs = map[string]newClientFunc{
	cryptos.Bitcoin.Name:     cryptocore.NewClientBTC,
	cryptos.Litecoin.Name:    cryptocore.NewClientLTC,
	cryptos.Dogecoin.Name:    cryptocore.NewClientDOGE,
	cryptos.Decred.Name:      cryptocore.NewClientDCR,
	cryptos.BitcoinCash.Name: cryptocore.NewClientBCH,
}

func newClient(fs *pflag.FlagSet, c *cryptos.Crypto) cryptocore.Client {
	nc, ok := newClientFuncs[c.Name]
	if !ok {
		errorExit(ecUnknownCrypto, c.Name)
	}
	return nc(
		flagRPCAddress(fs),
		flagRPCUsername(fs),
		flagRPCPassword(fs),
		flagRPCTLSConfig(fs),
	)
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

const (
	initTimeout = time.Second
	maxTimeout  = time.Second * 30
)

func blockTransactions(cl cryptocore.Client, txs []types.Bytes) ([]tx.Tx, error) {
	r := make([]tx.Tx, 0, len(txs))
	for _, i := range txs {
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
					defer timer.Stop()
					select {
					case <-closec:
						return
					case <-timer.C:
					}
					continue
				}
				errc <- err
				return
			}
			blockTxs, err := blockTransactions(cl, block.Transactions())
			if err != nil {
				errc <- err
				return
			}
			timeout = initTimeout
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
			blockTxs, err := blockTransactions(cl, block.Transactions())
			if err != nil {
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
