package etherclient

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Block contains common block data available across EVM chains and clients.
type Block struct {
	BaseFeePerGas         *string       `json:"baseFeePerGas"`
	BlobGasUsed           *string       `json:"blobGasUsed"`
	Difficulty            *string       `json:"difficulty"`
	ExcessBlobGas         *string       `json:"excessBlobGas"`
	ExtraData             *string       `json:"extraData"`
	GasLimit              *string       `json:"gasLimit"`
	GasUsed               *string       `json:"gasUsed"`
	Hash                  string        `json:"hash"`
	LogsBloom             *string       `json:"logsBloom"`
	Miner                 *string       `json:"miner"`
	MixHash               *string       `json:"mixHash"`
	Nonce                 *string       `json:"nonce"`
	Number                string        `json:"number"`
	ParentBeaconBlockRoot *string       `json:"parentBeaconBlockRoot"`
	ParentHash            string        `json:"parentHash"`
	ReceiptsRoot          *string       `json:"receiptsRoot"`
	Sha3Uncles            *string       `json:"sha3Uncles"`
	Size                  *string       `json:"size"`
	StateRoot             *string       `json:"stateRoot"`
	Timestamp             string        `json:"timestamp"`
	TotalDifficulty       *string       `json:"totalDifficulty"`
	Transactions          []Transaction `json:"transactions"`
	TransactionsRoot      *string       `json:"transactionsRoot"`
	Uncles                []*string     `json:"uncles"`
	Withdrawals           []Withdrawal  `json:"withdrawals"`
	WithdrawalsRoot       *string       `json:"withdrawalsRoot"`
}

// Transaction contains common transaction data available across EVM chains and clients.
type Transaction struct {
	BlockHash            string       `json:"blockHash"`
	BlockNumber          string       `json:"blockNumber"`
	From                 string       `json:"from"`
	Gas                  string       `json:"gas"`
	GasPrice             string       `json:"gasPrice"`
	MaxPriorityFeePerGas *string      `json:"maxPriorityFeePerGas,omitempty"`
	MaxFeePerGas         *string      `json:"maxFeePerGas,omitempty"`
	Hash                 string       `json:"hash"`
	Input                *string      `json:"input"`
	Nonce                string       `json:"nonce"`
	To                   *string      `json:"to"`
	TransactionIndex     string       `json:"transactionIndex"`
	Value                *string      `json:"value"`
	Type                 string       `json:"type"`
	AccessList           []AccessList `json:"accessList,omitempty"`
	ChainID              *string      `json:"chainId"`
	V                    string       `json:"v"`
	R                    string       `json:"r"`
	S                    string       `json:"s"`
	YParity              *string      `json:"yParity"`
}

// AccessList contains common access list data available across EVM chains and clients.
type AccessList struct {
	Address     string   `json:"address"`
	StorageKeys []string `json:"storageKeys"`
}

// Withdrawal contains common withdrawal data available across EVM chains and clients.
type Withdrawal struct {
	Index          string `json:"index"`
	ValidatorIndex string `json:"validatorIndex"`
	Address        string `json:"address"`
	Amount         string `json:"amount"`
}

func (ec *etherClient) BlockByHashCommon(ctx context.Context, hash common.Hash) (ret1 *Block, err error) {
	err = ec.withBackoff(ctx, "BlockByHashCommon()", func(ctx context.Context, ethClient *ethclient.Client) error {
		var r1 Block
		e := ethClient.Client().CallContext(
			ctx, &r1, "eth_getBlockByHash", hash,
			true,
		)
		if e != nil {
			return e
		}
		if r1.Hash == "" {
			return ErrNotFound
		}
		ret1 = &r1
		return e
	}, retryOptions{
		MinBackoff:     5 * time.Second,
		MaxElapsedTime: 12 * time.Hour,
		MaxBackoff:     15 * time.Second,
	})
	return
}

func (ec *etherClient) BlockByNumberCommon(ctx context.Context, number *big.Int) (ret1 *Block, err error) {
	err = ec.withBackoff(ctx, "BlockByNumberCommon()", func(ctx context.Context, ethClient *ethclient.Client) error {
		var r1 Block
		e := ethClient.Client().CallContext(
			ctx, &r1, "eth_getBlockByNumber", toBlockNumArg(number),
			true,
		)
		if e != nil {
			return e
		}
		if r1.Hash == "" {
			return ErrNotFound
		}
		ret1 = &r1
		return e
	}, retryOptions{
		MinBackoff:     5 * time.Second,
		MaxElapsedTime: 12 * time.Hour,
		MaxBackoff:     15 * time.Second,
	})
	return
}
