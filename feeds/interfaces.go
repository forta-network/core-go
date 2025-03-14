package feeds

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

// LogFeed is a feed of logs
type LogFeed interface {
	ForEachLog(handler func(blk *types.Block, logEntry types.Log) error, finishBlockHandler func(blk *types.Block) error) error
	GetLogsForLastBlocks(blocksAgo int64) ([]types.Log, error)
	GetLogsForRange(blockStart *big.Int, blockEnd *big.Int) ([]types.Log, error)
	AddAddress(newAddr string)
}
