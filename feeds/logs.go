package feeds

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/forta-network/core-go/etherclient"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type logFeed struct {
	ctx        context.Context
	startBlock *big.Int
	endBlock   *big.Int
	topics     [][]string
	client     etherclient.EtherClient
	offset     int

	addresses []common.Address
	addrsMu   sync.RWMutex
}

var _ LogFeed = &logFeed{}

func (l *logFeed) GetLogsForRange(startBlock *big.Int, endBlock *big.Int) ([]types.Log, error) {
	addrs := l.getAddrs()

	var topics [][]common.Hash
	for _, topicSet := range l.topics {
		var topicPosition []common.Hash
		for _, topic := range topicSet {
			topicHash := common.HexToHash(topic)
			topicPosition = append(topicPosition, topicHash)
		}
		topics = append(topics, topicPosition)
	}

	q := ethereum.FilterQuery{
		FromBlock: startBlock,
		ToBlock:   endBlock,
		Addresses: addrs,
		Topics:    topics,
	}

	return l.client.FilterLogs(l.ctx, q)
}

func (l *logFeed) GetLogsForLastBlocks(blocksAgo int64) ([]types.Log, error) {

	blk, err := l.client.GetBlockByNumber(l.ctx, nil)
	if err != nil {
		return nil, err
	}

	endBlock, err := hexutil.DecodeBig(blk.Number)
	if err != nil {
		return nil, err
	}

	startBlock := big.NewInt(endBlock.Int64() - blocksAgo)

	return l.GetLogsForRange(startBlock, endBlock)
}

func (l *logFeed) ForEachLog(handler func(blk *etherclient.Block, logEntry types.Log) error, finishBlockHandler func(blk *etherclient.Block) error) error {
	eg, ctx := errgroup.WithContext(l.ctx)

	var topics [][]common.Hash
	for _, topicSet := range l.topics {
		var topicPosition []common.Hash
		for _, topic := range topicSet {
			topicHash := common.HexToHash(topic)
			topicPosition = append(topicPosition, topicHash)
		}
		topics = append(topics, topicPosition)
	}

	currentBlock := l.startBlock
	increment := big.NewInt(1)
	eg.Go(func() error {
		for {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if currentBlock != nil && l.endBlock != nil {
				if currentBlock.Cmp(l.endBlock) > 0 {
					log.Infof("completed processing logs (endBlock reached)")
					return nil
				}
			}

			blk, err := l.client.GetBlockByNumber(l.ctx, currentBlock)
			if err != nil {
				log.WithError(err).Error("error while getting latest block number")
				return err
			}

			// initialize current block if nil
			if currentBlock == nil {
				currentBlock, err = hexutil.DecodeBig(blk.Number)
				if err != nil {
					log.Errorf("error while converting latest block number: %s, %s", blk.Number, err)
					return err
				}
			}

			blockToRetrieve := big.NewInt(currentBlock.Int64() - int64(l.offset))

			// if offset is set, get previous block instead
			if l.offset > 0 {
				pastBlock, err := l.client.GetBlockByNumber(l.ctx, blockToRetrieve)
				if err != nil {
					log.WithError(err).Error("error while getting past block")
					return err
				}
				blk = pastBlock
			}

			addrs := l.getAddrs()

			q := ethereum.FilterQuery{
				FromBlock: blockToRetrieve,
				ToBlock:   blockToRetrieve,
				Addresses: addrs,
				Topics:    topics,
			}
			logs, err := l.client.FilterLogs(l.ctx, q)
			if err != nil {
				return err
			}

			for _, lg := range logs {
				if err := handler(blk, lg); err != nil {
					log.Error("handler returned error, exiting log subscription:", err)
					return err
				}
			}

			currentBlock = currentBlock.Add(currentBlock, increment)
			if err := finishBlockHandler(blk); err != nil {
				return err
			}
		}
	})
	log.Infof("subscribed to logs: address=%v, topics=%v, startBlock=%s, endBlock=%s", l.addresses, l.topics, l.startBlock, l.endBlock)
	defer func() {
		log.Info("log subscription closed")
	}()
	return eg.Wait()
}

// ForEachLogPolling processes every block that appears on‑chain,
// polling the RPC node at the given interval.
//
//   - It remembers the last processed height in-memory.
//   - On every tick it asks the node for the current tip and
//     loops from lastProcessed+1 … tip, invoking the handlers
//     exactly once per block.
//   - If l.endBlock != nil the loop stops after that height
func (l *logFeed) ForEachLogPolling(
	interval time.Duration,
	handler func(blk *etherclient.Block, lg types.Log) error,
	finishBlockHandler func(blk *etherclient.Block) error,
) error {
	// prepare topic matrix once
	topics := make([][]common.Hash, len(l.topics))
	for i, set := range l.topics {
		topics[i] = make([]common.Hash, len(set))
		for j, t := range set {
			topics[i][j] = common.HexToHash(t)
		}
	}

	var cursor *big.Int
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-l.ctx.Done():
			return l.ctx.Err()

		case <-ticker.C:
			// discover the latest tip
			tipUint, err := l.client.BlockNumber(l.ctx)
			if err != nil {
				return fmt.Errorf("tip discovery failed: %w", err)
			}
			tip := big.NewInt(int64(tipUint))

			// initialize cursor on first iteration
			if cursor == nil {
				if l.startBlock != nil {
					cursor = new(big.Int).Set(l.startBlock)
				} else {
					cursor = new(big.Int).Set(tip)
				}
			}

			// if nothing new, wait for next tick
			if cursor.Cmp(tip) > 0 {
				continue
			}

			// walk from cursor … tip
			for cursor.Cmp(tip) <= 0 {
				// optional stop height
				if l.endBlock != nil && cursor.Cmp(l.endBlock) > 0 {
					return nil
				}

				// fetch block
				blk, err := l.client.GetBlockByNumber(l.ctx, cursor)
				if err != nil {
					// node not indexed yet? retry same cursor next tick
					if strings.Contains(err.Error(), "not found") {
						break
					}
					return err
				}

				// build filter query (apply offset)
				from := new(big.Int).Sub(cursor, big.NewInt(int64(l.offset)))
				q := ethereum.FilterQuery{
					FromBlock: from,
					ToBlock:   from,
					Addresses: l.getAddrs(),
					Topics:    topics,
				}
				logs, err := l.client.FilterLogs(l.ctx, q)
				if err != nil {
					if strings.Contains(err.Error(), "not found") {
						break
					}
					return err
				}

				// deliver logs
				for _, lg := range logs {
					if err := handler(blk, lg); err != nil {
						return err
					}
				}
				if err := finishBlockHandler(blk); err != nil {
					return err
				}

				// advance cursor by 1
				cursor.Add(cursor, big.NewInt(1))
			}
		}
	}
}

func (l *logFeed) getAddrs() []common.Address {
	l.addrsMu.RLock()
	defer l.addrsMu.RUnlock()

	return l.addresses
}

func (l *logFeed) AddAddress(newAddr string) {
	l.addrsMu.Lock()
	defer l.addrsMu.Unlock()

	for _, addr := range l.addresses {
		if strings.EqualFold(addr.Hex(), newAddr) {
			return
		}
	}
	l.addresses = append(l.addresses, common.HexToAddress(newAddr))
}

type LogFeedConfig struct {
	Topics     [][]string
	Addresses  []string
	StartBlock *big.Int
	EndBlock   *big.Int
	Offset     int
}

func NewLogFeed(ctx context.Context, client etherclient.EtherClient, cfg LogFeedConfig) (*logFeed, error) {
	if cfg.Offset < 0 {
		return nil, fmt.Errorf("offset cannot be below zero: offset=%d", cfg.Offset)
	}
	addrs := make([]common.Address, 0, len(cfg.Addresses))
	for _, addr := range cfg.Addresses {
		addrs = append(addrs, common.HexToAddress(addr))
	}
	return &logFeed{
		ctx:        ctx,
		client:     client,
		topics:     cfg.Topics,
		addresses:  addrs,
		startBlock: cfg.StartBlock,
		endBlock:   cfg.EndBlock,
		offset:     cfg.Offset,
	}, nil
}
