package etherclient

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

// EthClient is the original interface from go-ethereum.
type EthClient interface {
	ethereum.ChainReader
	ethereum.ChainStateReader
	ethereum.TransactionReader
	ethereum.LogFilterer
	ethereum.ContractCaller
	ethereum.GasEstimator
	ethereum.GasPricer
	ethereum.GasPricer1559
	ethereum.ChainSyncReader
	ethereum.ChainIDReader
	ethereum.PendingContractCaller
	ethereum.PendingStateReader
	ethereum.TransactionReader
	ethereum.TransactionSender
	ethereum.BlockNumberReader
	ethereum.FeeHistoryReader

	PeerCount(ctx context.Context) (uint64, error)
	BlockReceipts(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) ([]*types.Receipt, error)
	TransactionSender(ctx context.Context, tx *types.Transaction, block common.Hash, index uint) (common.Address, error)
	NetworkID(ctx context.Context) (*big.Int, error)
	BalanceAtHash(ctx context.Context, account common.Address, blockHash common.Hash) (*big.Int, error)
	StorageAtHash(ctx context.Context, account common.Address, key common.Hash, blockHash common.Hash) ([]byte, error)
	CodeAtHash(ctx context.Context, account common.Address, blockHash common.Hash) ([]byte, error)
	NonceAtHash(ctx context.Context, account common.Address, blockHash common.Hash) (uint64, error)
	CallContractAtHash(ctx context.Context, msg ethereum.CallMsg, blockHash common.Hash) ([]byte, error)

	Client() *rpc.Client
	Close()
}

type Extras interface {
	DebugTraceCall(
		ctx context.Context, req *TraceCallTransaction,
		block any, traceCallConfig TraceCallConfig,
		result interface{},
	) error
	GetBlockRawTransactions(ctx context.Context, number *big.Int) ([]string, error)
}

// EtherClient is the extended interface implemented implemented by this package.
type EtherClient interface {
	EthClient
	Extras
}
