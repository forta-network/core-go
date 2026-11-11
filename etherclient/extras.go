package etherclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/holiman/uint256"
)

// TraceCallTransaction contains the fields of the to-be-simulated transaction.
type TraceCallTransaction struct {
	From                 *common.Address `json:"from"`
	To                   *common.Address `json:"to,omitempty"`
	Gas                  *hexutil.Uint64 `json:"gas,omitempty"`
	GasPrice             *hexutil.Big    `json:"gasPrice,omitempty"`
	MaxFeePerGas         *hexutil.Big    `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas *hexutil.Big    `json:"maxPriorityFeePerGas,omitempty"`
	Value                *hexutil.Big    `json:"value,omitempty"`
	Nonce                *hexutil.Uint64 `json:"nonce"`

	// We accept "data" and "input" for backwards-compatibility reasons.
	// "input" is the newer name and should be preferred by clients.
	// Issue detail: https://github.com/ethereum/go-ethereum/issues/15628
	Data  *hexutil.Bytes `json:"data"`
	Input *hexutil.Bytes `json:"input"`

	// Introduced by AccessListTxType transaction.
	AccessList *gethtypes.AccessList `json:"accessList,omitempty"`
	ChainID    *hexutil.Big          `json:"chainId,omitempty"`

	// For BlobTxType
	BlobFeeCap *hexutil.Big  `json:"maxFeePerBlobGas"`
	BlobHashes []common.Hash `json:"blobVersionedHashes,omitempty"`

	// For BlobTxType transactions with blob sidecar
	Blobs       []kzg4844.Blob       `json:"blobs"`
	Commitments []kzg4844.Commitment `json:"commitments"`
	Proofs      []kzg4844.Proof      `json:"proofs"`

	// For SetCodeTxType
	AuthorizationList []gethtypes.SetCodeAuthorization `json:"authorizationList"`
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
	}, retryOptions{
		MinBackoff:     ec.retryInterval,
		MaxElapsedTime: 1 * time.Minute,
		MaxBackoff:     ec.retryInterval,
	})
}

func (ec *etherClient) DebugTraceTransaction(
	ctx context.Context, txHash string, traceCallConfig TraceCallConfig, result interface{},
) error {
	return ec.withBackoff(ctx, "DebugTraceTransaction()", func(ctx context.Context, ethClient *ethclient.Client) error {
		return ethClient.Client().CallContext(ctx, &result, "debug_traceTransaction", txHash, traceCallConfig)
	}, retryOptions{
		MinBackoff:     ec.retryInterval,
		MaxElapsedTime: 1 * time.Minute,
		MaxBackoff:     ec.retryInterval,
	})
}

type BlockTraceTx struct {
	TxHash string      `json:"txHash"`
	Result *TracedCall `json:"result"`
}

type TracedBlock []*BlockTraceTx

func (ec *etherClient) DebugTraceBlockByNumber(
	ctx context.Context, blockNumber *big.Int, traceCallConfig TraceCallConfig, result interface{},
) error {
	return ec.withBackoff(ctx, "DebugTraceBlockByNumber()", func(ctx context.Context, ethClient *ethclient.Client) error {
		return ethClient.Client().CallContext(ctx, &result, "debug_traceBlockByNumber", toBlockNumArg(blockNumber), traceCallConfig)
	}, retryOptions{
		MinBackoff:     ec.retryInterval,
		MaxElapsedTime: 1 * time.Minute,
		MaxBackoff:     ec.retryInterval,
	})
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	if number.Sign() >= 0 {
		return hexutil.EncodeBig(number)
	}
	// It's negative.
	if number.IsInt64() {
		return rpc.BlockNumber(number.Int64()).String()
	}
	// It's negative and large, which is invalid.
	return fmt.Sprintf("<invalid %d>", number)
}

type BlockTx struct {
	From              string                           `json:"from"`
	To                string                           `json:"to"`
	Nonce             *hexutil.Uint64                  `json:"nonce"`
	Value             *hexutil.Big                     `json:"value"`
	Input             string                           `json:"input"`
	Hash              string                           `json:"hash"`
	Type              TxType                           `json:"type"`
	AuthorizationList []gethtypes.SetCodeAuthorization `json:"authorizationList"`
}

// blockTxJSON is the wire form of BlockTx. Its authorization list shadows the
// one of BlockTx, so that the authorizations are decoded by authorizationJSON.
type blockTxJSON struct {
	BlockTx
	AuthorizationList []authorizationJSON `json:"authorizationList"`
}

// authorizationJSON decodes an EIP-7702 authorization and accepts any yParity.
//
// A block can contain an authorization whose yParity is neither 0 nor 1: the
// authorization is skipped and the transaction is still valid. go-ethereum 1.17.9
// and later reject such an authorization in JSON, which would fail the decoding
// of the whole block. This is the decoder go-ethereum generated up to 1.17.8, so
// the result is the same whichever go-ethereum version the build selects.
type authorizationJSON gethtypes.SetCodeAuthorization

func (a *authorizationJSON) UnmarshalJSON(input []byte) error {
	var dec struct {
		ChainID *hexutil.U256   `json:"chainId"`
		Address *common.Address `json:"address"`
		Nonce   *hexutil.Uint64 `json:"nonce"`
		V       *hexutil.Uint64 `json:"yParity"`
		R       *hexutil.U256   `json:"r"`
		S       *hexutil.U256   `json:"s"`
	}
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.ChainID == nil {
		return errors.New("missing required field 'chainId' for SetCodeAuthorization")
	}
	a.ChainID = uint256.Int(*dec.ChainID)
	if dec.Address == nil {
		return errors.New("missing required field 'address' for SetCodeAuthorization")
	}
	a.Address = *dec.Address
	if dec.Nonce == nil {
		return errors.New("missing required field 'nonce' for SetCodeAuthorization")
	}
	a.Nonce = uint64(*dec.Nonce)
	if dec.V == nil {
		return errors.New("missing required field 'yParity' for SetCodeAuthorization")
	}
	a.V = uint8(*dec.V)
	if dec.R == nil {
		return errors.New("missing required field 'r' for SetCodeAuthorization")
	}
	a.R = uint256.Int(*dec.R)
	if dec.S == nil {
		return errors.New("missing required field 's' for SetCodeAuthorization")
	}
	a.S = uint256.Int(*dec.S)
	return nil
}

// toBlockTx returns the transaction with its decoded authorization list.
func (tx *blockTxJSON) toBlockTx() *BlockTx {
	if tx == nil {
		return nil
	}
	if tx.AuthorizationList != nil {
		tx.BlockTx.AuthorizationList = make([]gethtypes.SetCodeAuthorization, len(tx.AuthorizationList))
		for i, auth := range tx.AuthorizationList {
			tx.BlockTx.AuthorizationList[i] = gethtypes.SetCodeAuthorization(auth)
		}
	}
	return &tx.BlockTx
}

// GetBlockTransactions returns the raw transactions in a block.
func (ec *etherClient) GetBlockTransactions(ctx context.Context, number *big.Int) ([]*BlockTx, error) {
	var block struct {
		Hash         string         `json:"hash"`
		Transactions []*blockTxJSON `json:"transactions"`
	}
	err := ec.withBackoff(ctx, "GetBlockTransactions()", func(ctx context.Context, ethClient *ethclient.Client) error {
		err := ethClient.Client().CallContext(
			ctx, &block, "eth_getBlockByNumber", toBlockNumArg(number),
			true,
		)
		if err != nil {
			return err
		}
		if block.Hash == "" {
			return ethereum.NotFound
		}
		return nil
	}, retryOptions{
		MinBackoff:     ec.retryInterval,
		MaxElapsedTime: 12 * time.Hour,
		MaxBackoff:     ec.retryInterval,
	})
	if err != nil {
		return nil, err
	}
	if block.Transactions == nil {
		return nil, nil
	}

	txs := make([]*BlockTx, len(block.Transactions))
	for i, tx := range block.Transactions {
		txs[i] = tx.toBlockTx()
	}
	return txs, nil
}
