package etherclient

import (
	"context"
	"math/big"

	ethereum "github.com/forta-network/op-geth"
	"github.com/forta-network/op-geth/common"
	"github.com/forta-network/op-geth/core/types"
	"github.com/forta-network/op-geth/ethclient"
	"github.com/forta-network/op-geth/rpc"
)

// opEtherClient is a wrapper of go-ethereum ethclient.Client which uses multiple fallback
// clients and retries every request.
type opEtherClient struct {
	provider Provider[OpEthClient]
}

var _ OpEtherClient = &opEtherClient{}

// DialContextOp dials all given URLs and creates a client that works with multiple clients
// and a backoff logic.
func DialContextOp(ctx context.Context, rawurls ...string) (*opEtherClient, error) {
	var clients []OpEthClient
	for _, rawurl := range rawurls {
		c, err := ethclient.DialContext(ctx, rawurl)
		if err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return &opEtherClient{provider: NewRingProvider(clients...)}, nil
}

func (ec *opEtherClient) Close() {
	ec.provider.Close()
}

func (ec *opEtherClient) Client() *rpc.Client {
	return ec.provider.Provide().Client()
}

func (ec *opEtherClient) ChainID(ctx context.Context) (ret1 *big.Int, err error) {
	err = withBackoff(ctx, ec.provider, "ChainID()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.ChainID(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) BlockByHash(ctx context.Context, hash common.Hash) (ret1 *types.Block, err error) {
	err = withBackoff(ctx, ec.provider, "BlockByHash()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.BlockByHash(ctx, hash)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) BlockByNumber(ctx context.Context, number *big.Int) (ret1 *types.Block, err error) {
	err = withBackoff(ctx, ec.provider, "BlockByNumber()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.BlockByNumber(ctx, number)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) BlockNumber(ctx context.Context) (ret1 uint64, err error) {
	err = withBackoff(ctx, ec.provider, "BlockNumber()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.BlockNumber(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) PeerCount(ctx context.Context) (ret1 uint64, err error) {
	err = withBackoff(ctx, ec.provider, "PeerCount()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.PeerCount(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) BlockReceipts(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (ret1 []*types.Receipt, err error) {
	err = withBackoff(ctx, ec.provider, "BlockReceipts()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.BlockReceipts(ctx, blockNrOrHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) HeaderByHash(ctx context.Context, hash common.Hash) (ret1 *types.Header, err error) {
	err = withBackoff(ctx, ec.provider, "HeaderByHash()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.HeaderByHash(ctx, hash)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) HeaderByNumber(ctx context.Context, number *big.Int) (ret1 *types.Header, err error) {
	err = withBackoff(ctx, ec.provider, "HeaderByNumber()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.HeaderByNumber(ctx, number)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) TransactionByHash(ctx context.Context, hash common.Hash) (ret1 *types.Transaction, ret2 bool, err error) {
	err = withBackoff(ctx, ec.provider, "TransactionByHash()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, r2, e := ethClient.TransactionByHash(ctx, hash)
		ret1 = r1
		ret2 = r2
		return e
	})
	return
}

func (ec *opEtherClient) TransactionSender(ctx context.Context, tx *types.Transaction, block common.Hash, index uint) (ret1 common.Address, err error) {
	err = withBackoff(ctx, ec.provider, "TransactionSender()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.TransactionSender(ctx, tx, block, index)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) TransactionCount(ctx context.Context, blockHash common.Hash) (ret1 uint, err error) {
	err = withBackoff(ctx, ec.provider, "TransactionCount()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.TransactionCount(ctx, blockHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (ret1 *types.Transaction, err error) {
	err = withBackoff(ctx, ec.provider, "TransactionInBlock()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.TransactionInBlock(ctx, blockHash, index)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (ret1 *types.Receipt, err error) {
	err = withBackoff(ctx, ec.provider, "TransactionReceipt()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.TransactionReceipt(ctx, txHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) SyncProgress(ctx context.Context) (ret1 *ethereum.SyncProgress, err error) {
	err = withBackoff(ctx, ec.provider, "SyncProgress()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.SyncProgress(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ret1 ethereum.Subscription, err error) {
	err = withBackoff(ctx, ec.provider, "SubscribeNewHead()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.SubscribeNewHead(ctx, ch)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) NetworkID(ctx context.Context) (ret1 *big.Int, err error) {
	err = withBackoff(ctx, ec.provider, "NetworkID()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.NetworkID(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (ret1 *big.Int, err error) {
	err = withBackoff(ctx, ec.provider, "BalanceAt()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.BalanceAt(ctx, account, blockNumber)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) BalanceAtHash(ctx context.Context, account common.Address, blockHash common.Hash) (ret1 *big.Int, err error) {
	err = withBackoff(ctx, ec.provider, "BalanceAtHash()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.BalanceAtHash(ctx, account, blockHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) (ret1 []byte, err error) {
	err = withBackoff(ctx, ec.provider, "StorageAt()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.StorageAt(ctx, account, key, blockNumber)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) StorageAtHash(ctx context.Context, account common.Address, key common.Hash, blockHash common.Hash) (ret1 []byte, err error) {
	err = withBackoff(ctx, ec.provider, "StorageAtHash()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.StorageAtHash(ctx, account, key, blockHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) (ret1 []byte, err error) {
	err = withBackoff(ctx, ec.provider, "CodeAt()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.CodeAt(ctx, account, blockNumber)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) CodeAtHash(ctx context.Context, account common.Address, blockHash common.Hash) (ret1 []byte, err error) {
	err = withBackoff(ctx, ec.provider, "CodeAtHash()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.CodeAtHash(ctx, account, blockHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (ret1 uint64, err error) {
	err = withBackoff(ctx, ec.provider, "NonceAt()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.NonceAt(ctx, account, blockNumber)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) NonceAtHash(ctx context.Context, account common.Address, blockHash common.Hash) (ret1 uint64, err error) {
	err = withBackoff(ctx, ec.provider, "NonceAtHash()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.NonceAtHash(ctx, account, blockHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) (ret1 []types.Log, err error) {
	err = withBackoff(ctx, ec.provider, "FilterLogs()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.FilterLogs(ctx, q)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ret1 ethereum.Subscription, err error) {
	err = withBackoff(ctx, ec.provider, "SubscribeFilterLogs()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.SubscribeFilterLogs(ctx, q, ch)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) PendingBalanceAt(ctx context.Context, account common.Address) (ret1 *big.Int, err error) {
	err = withBackoff(ctx, ec.provider, "PendingBalanceAt()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.PendingBalanceAt(ctx, account)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) PendingStorageAt(ctx context.Context, account common.Address, key common.Hash) (ret1 []byte, err error) {
	err = withBackoff(ctx, ec.provider, "PendingStorageAt()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.PendingStorageAt(ctx, account, key)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) PendingCodeAt(ctx context.Context, account common.Address) (ret1 []byte, err error) {
	err = withBackoff(ctx, ec.provider, "PendingCodeAt()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.PendingCodeAt(ctx, account)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) PendingNonceAt(ctx context.Context, account common.Address) (ret1 uint64, err error) {
	err = withBackoff(ctx, ec.provider, "PendingNonceAt()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.PendingNonceAt(ctx, account)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) PendingTransactionCount(ctx context.Context) (ret1 uint, err error) {
	err = withBackoff(ctx, ec.provider, "PendingTransactionCount()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.PendingTransactionCount(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) (ret1 []byte, err error) {
	err = withBackoff(ctx, ec.provider, "CallContract()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.CallContract(ctx, msg, blockNumber)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) CallContractAtHash(ctx context.Context, msg ethereum.CallMsg, blockHash common.Hash) (ret1 []byte, err error) {
	err = withBackoff(ctx, ec.provider, "CallContractAtHash()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.CallContractAtHash(ctx, msg, blockHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) PendingCallContract(ctx context.Context, msg ethereum.CallMsg) (ret1 []byte, err error) {
	err = withBackoff(ctx, ec.provider, "PendingCallContract()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.PendingCallContract(ctx, msg)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) SuggestGasPrice(ctx context.Context) (ret1 *big.Int, err error) {
	err = withBackoff(ctx, ec.provider, "SuggestGasPrice()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.SuggestGasPrice(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) SuggestGasTipCap(ctx context.Context) (ret1 *big.Int, err error) {
	err = withBackoff(ctx, ec.provider, "SuggestGasTipCap()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.SuggestGasTipCap(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (ret1 *ethereum.FeeHistory, err error) {
	err = withBackoff(ctx, ec.provider, "FeeHistory()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.FeeHistory(ctx, blockCount, lastBlock, rewardPercentiles)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (ret1 uint64, err error) {
	err = withBackoff(ctx, ec.provider, "EstimateGas()", func(ctx context.Context, ethClient OpEthClient) error {
		r1, e := ethClient.EstimateGas(ctx, msg)
		ret1 = r1
		return e
	})
	return
}

func (ec *opEtherClient) SendTransaction(ctx context.Context, tx *types.Transaction) (err error) {
	return withBackoff(ctx, ec.provider, "SendTransaction()", func(ctx context.Context, ethClient OpEthClient) error {
		return ethClient.SendTransaction(ctx, tx)
	})
}
