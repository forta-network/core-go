package etherclient

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/forta-network/core-go/etherclient/provider"
	"github.com/sirupsen/logrus"
)

const (
	backoffInitialInterval = time.Second
	backoffMaxInterval     = time.Minute
	backoffMaxElapsedTime  = time.Minute * 5
	backoffContextTimeout  = time.Minute
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
}

// EtherClient is the extended interface implemented implemented by this package.
type EtherClient interface {
	EthClient
	DebugTraceCall(
		ctx context.Context, req *TraceCallTransaction,
		block any, traceCallConfig TraceCallConfig,
		result interface{},
	) error
}

// etherClient is a wrapper of go-ethereum ethclient.Client which uses multiple fallback
// clients and retries every request.
type etherClient struct {
	provider provider.Provider[*ethclient.Client]
}

var _ EtherClient = &etherClient{}

// NewRetrierClient dials all given URLs and creates a client that works with multiple clients
// and a backoff logic.
func DialContext(ctx context.Context, rawurls ...string) (*etherClient, error) {
	var clients []*ethclient.Client
	for _, rawurl := range rawurls {
		c, err := ethclient.DialContext(ctx, rawurl)
		if err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return &etherClient{provider: provider.NewRingProvider(clients...)}, nil
}

func (ec *etherClient) Close() {
	ec.provider.Close()
}

func (ec *etherClient) withBackoff(
	ctx context.Context,
	method string,
	operation func(ctx context.Context, ethClient *ethclient.Client) error,
) error {
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = backoffInitialInterval
	bo.MaxInterval = backoffMaxInterval
	bo.MaxElapsedTime = backoffMaxElapsedTime
	err := backoff.Retry(func() error {
		if ctx.Err() != nil {
			return backoff.Permanent(ctx.Err())
		}
		tCtx, cancel := context.WithTimeout(ctx, backoffContextTimeout)
		err := operation(tCtx, ec.provider.Provide())
		cancel()
		if err != nil {
			// Move onto the next provider.
			ec.provider.Next()
		}
		return handleRetryErr(ctx, method, err)
	}, bo)
	if err != nil {
		logrus.WithError(err).WithField("method", method).Error("retry failed with error")
	}
	return err
}

func (ec *etherClient) ChainID(ctx context.Context) (ret1 *big.Int, err error) {
	err = ec.withBackoff(ctx, "ChainID()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.ChainID(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) BlockByHash(ctx context.Context, hash common.Hash) (ret1 *types.Block, err error) {
	err = ec.withBackoff(ctx, "BlockByHash()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.BlockByHash(ctx, hash)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) BlockByNumber(ctx context.Context, number *big.Int) (ret1 *types.Block, err error) {
	err = ec.withBackoff(ctx, "BlockByNumber()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.BlockByNumber(ctx, number)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) BlockNumber(ctx context.Context) (ret1 uint64, err error) {
	err = ec.withBackoff(ctx, "BlockNumber()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.BlockNumber(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) PeerCount(ctx context.Context) (ret1 uint64, err error) {
	err = ec.withBackoff(ctx, "PeerCount()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.PeerCount(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) BlockReceipts(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (ret1 []*types.Receipt, err error) {
	err = ec.withBackoff(ctx, "BlockReceipts()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.BlockReceipts(ctx, blockNrOrHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) HeaderByHash(ctx context.Context, hash common.Hash) (ret1 *types.Header, err error) {
	err = ec.withBackoff(ctx, "HeaderByHash()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.HeaderByHash(ctx, hash)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) HeaderByNumber(ctx context.Context, number *big.Int) (ret1 *types.Header, err error) {
	err = ec.withBackoff(ctx, "HeaderByNumber()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.HeaderByNumber(ctx, number)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) TransactionByHash(ctx context.Context, hash common.Hash) (ret1 *types.Transaction, ret2 bool, err error) {
	err = ec.withBackoff(ctx, "TransactionByHash()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, r2, e := ethClient.TransactionByHash(ctx, hash)
		ret1 = r1
		ret2 = r2
		return e
	})
	return
}

func (ec *etherClient) TransactionSender(ctx context.Context, tx *types.Transaction, block common.Hash, index uint) (ret1 common.Address, err error) {
	err = ec.withBackoff(ctx, "TransactionSender()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.TransactionSender(ctx, tx, block, index)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) TransactionCount(ctx context.Context, blockHash common.Hash) (ret1 uint, err error) {
	err = ec.withBackoff(ctx, "TransactionCount()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.TransactionCount(ctx, blockHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (ret1 *types.Transaction, err error) {
	err = ec.withBackoff(ctx, "TransactionInBlock()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.TransactionInBlock(ctx, blockHash, index)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (ret1 *types.Receipt, err error) {
	err = ec.withBackoff(ctx, "TransactionReceipt()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.TransactionReceipt(ctx, txHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) SyncProgress(ctx context.Context) (ret1 *ethereum.SyncProgress, err error) {
	err = ec.withBackoff(ctx, "SyncProgress()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.SyncProgress(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ret1 ethereum.Subscription, err error) {
	err = ec.withBackoff(ctx, "SubscribeNewHead()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.SubscribeNewHead(ctx, ch)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) NetworkID(ctx context.Context) (ret1 *big.Int, err error) {
	err = ec.withBackoff(ctx, "NetworkID()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.NetworkID(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (ret1 *big.Int, err error) {
	err = ec.withBackoff(ctx, "BalanceAt()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.BalanceAt(ctx, account, blockNumber)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) BalanceAtHash(ctx context.Context, account common.Address, blockHash common.Hash) (ret1 *big.Int, err error) {
	err = ec.withBackoff(ctx, "BalanceAtHash()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.BalanceAtHash(ctx, account, blockHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) (ret1 []byte, err error) {
	err = ec.withBackoff(ctx, "StorageAt()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.StorageAt(ctx, account, key, blockNumber)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) StorageAtHash(ctx context.Context, account common.Address, key common.Hash, blockHash common.Hash) (ret1 []byte, err error) {
	err = ec.withBackoff(ctx, "StorageAtHash()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.StorageAtHash(ctx, account, key, blockHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) (ret1 []byte, err error) {
	err = ec.withBackoff(ctx, "CodeAt()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.CodeAt(ctx, account, blockNumber)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) CodeAtHash(ctx context.Context, account common.Address, blockHash common.Hash) (ret1 []byte, err error) {
	err = ec.withBackoff(ctx, "CodeAtHash()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.CodeAtHash(ctx, account, blockHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (ret1 uint64, err error) {
	err = ec.withBackoff(ctx, "NonceAt()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.NonceAt(ctx, account, blockNumber)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) NonceAtHash(ctx context.Context, account common.Address, blockHash common.Hash) (ret1 uint64, err error) {
	err = ec.withBackoff(ctx, "NonceAtHash()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.NonceAtHash(ctx, account, blockHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) (ret1 []types.Log, err error) {
	err = ec.withBackoff(ctx, "FilterLogs()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.FilterLogs(ctx, q)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ret1 ethereum.Subscription, err error) {
	err = ec.withBackoff(ctx, "SubscribeFilterLogs()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.SubscribeFilterLogs(ctx, q, ch)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) PendingBalanceAt(ctx context.Context, account common.Address) (ret1 *big.Int, err error) {
	err = ec.withBackoff(ctx, "PendingBalanceAt()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.PendingBalanceAt(ctx, account)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) PendingStorageAt(ctx context.Context, account common.Address, key common.Hash) (ret1 []byte, err error) {
	err = ec.withBackoff(ctx, "PendingStorageAt()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.PendingStorageAt(ctx, account, key)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) PendingCodeAt(ctx context.Context, account common.Address) (ret1 []byte, err error) {
	err = ec.withBackoff(ctx, "PendingCodeAt()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.PendingCodeAt(ctx, account)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) PendingNonceAt(ctx context.Context, account common.Address) (ret1 uint64, err error) {
	err = ec.withBackoff(ctx, "PendingNonceAt()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.PendingNonceAt(ctx, account)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) PendingTransactionCount(ctx context.Context) (ret1 uint, err error) {
	err = ec.withBackoff(ctx, "PendingTransactionCount()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.PendingTransactionCount(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) (ret1 []byte, err error) {
	err = ec.withBackoff(ctx, "CallContract()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.CallContract(ctx, msg, blockNumber)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) CallContractAtHash(ctx context.Context, msg ethereum.CallMsg, blockHash common.Hash) (ret1 []byte, err error) {
	err = ec.withBackoff(ctx, "CallContractAtHash()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.CallContractAtHash(ctx, msg, blockHash)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) PendingCallContract(ctx context.Context, msg ethereum.CallMsg) (ret1 []byte, err error) {
	err = ec.withBackoff(ctx, "PendingCallContract()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.PendingCallContract(ctx, msg)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) SuggestGasPrice(ctx context.Context) (ret1 *big.Int, err error) {
	err = ec.withBackoff(ctx, "SuggestGasPrice()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.SuggestGasPrice(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) SuggestGasTipCap(ctx context.Context) (ret1 *big.Int, err error) {
	err = ec.withBackoff(ctx, "SuggestGasTipCap()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.SuggestGasTipCap(ctx)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (ret1 *ethereum.FeeHistory, err error) {
	err = ec.withBackoff(ctx, "FeeHistory()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.FeeHistory(ctx, blockCount, lastBlock, rewardPercentiles)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (ret1 uint64, err error) {
	err = ec.withBackoff(ctx, "EstimateGas()", func(ctx context.Context, ethClient *ethclient.Client) error {
		r1, e := ethClient.EstimateGas(ctx, msg)
		ret1 = r1
		return e
	})
	return
}

func (ec *etherClient) SendTransaction(ctx context.Context, tx *types.Transaction) (err error) {
	return ec.withBackoff(ctx, "SendTransaction()", func(ctx context.Context, ethClient *ethclient.Client) error {
		return ethClient.SendTransaction(ctx, tx)
	})
}

// Extra methods

// TraceCallTransaction contains the fields of the to-be-simulated transaction.
type TraceCallTransaction struct {
	From     string       `json:"from"`
	To       string       `json:"to"`
	Gas      *hexutil.Big `json:"gas,omitempty"`
	GasPrice *hexutil.Big `json:"gasPrice,omitempty"`
	Value    *hexutil.Big `json:"value,omitempty"`
	Data     string       `json:"data"`
}

// TraceCallConfig contains the tracer configuration to be used while simulating the transaction.
type TraceCallConfig struct {
	Tracer         string                 `json:"tracer,omitempty"`
	TracerConfig   *TracerConfig          `json:"tracerConfig,omitempty"`
	StateOverrides map[string]interface{} `json:"stateOverrides,omitempty"`
}

// TracerConfig contains some extra tracer parameters.
type TracerConfig struct {
	WithLog     bool `json:"withLog,omitempty"`
	OnlyTopCall bool `json:"onlyTopCall,omitempty"`
}

// TracedCall contains traced call data. This also represents the top level object
// in the debug_traceCall response.
type TracedCall struct {
	From     common.Address  `json:"from"`
	To       common.Address  `json:"to"`
	CallType string          `json:"type"`
	GasUsed  *hexutil.Big    `json:"gasUsed"`
	Input    string          `json:"input"`
	Output   string          `json:"output"`
	Error    string          `json:"error"`
	Calls    []*TracedCall   `json:"calls"`
	Logs     []*TracedLog    `json:"logs"`
	Raw      json.RawMessage `json:"-"`
	Value    *hexutil.Big    `json:"value"`
}

// TracedLog contains log data from trace.
type TracedLog struct {
	Index   int            `json:"index"`
	Address common.Address `json:"address"`
	Topics  []string       `json:"topics"`
	Data    hexutil.Bytes  `json:"data"`
}

func (ec *etherClient) DebugTraceCall(
	ctx context.Context, req *TraceCallTransaction,
	block any, traceCallConfig TraceCallConfig,
	result interface{},
) error {
	switch block.(type) {
	case string:
	case *rpc.BlockNumberOrHash:
	default:
		return errors.New("invalid block number type")
	}

	return ec.withBackoff(ctx, "DebugTraceCall()", func(ctx context.Context, ethClient *ethclient.Client) error {
		return ethClient.Client().CallContext(ctx, &result, "debug_traceCall", req, block, traceCallConfig)
	})
}
