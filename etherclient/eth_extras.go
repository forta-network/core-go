package etherclient

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

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

	return withBackoff(ctx, ec.provider, "DebugTraceCall()", func(ctx context.Context, ethClient EthClient) error {
		return ethClient.Client().CallContext(ctx, &result, "debug_traceCall", req, block, traceCallConfig)
	})
}

// GetBlockRawTransactions returns the raw transactions in a block.
func (ec *etherClient) GetBlockRawTransactions(ctx context.Context, number *big.Int) ([]string, error) {
	block, err := ec.BlockByNumber(ctx, number)
	if err != nil {
		return nil, err
	}

	var rawTxs []string
	for _, tx := range block.Transactions() {
		rawTxBytes, err := tx.MarshalBinary()
		if err != nil {
			return nil, err
		}
		rawTxs = append(rawTxs, (*hexutil.Bytes)(&rawTxBytes).String())
	}

	return rawTxs, nil
}
