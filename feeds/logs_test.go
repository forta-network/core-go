package feeds

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/forta-network/core-go/etherclient"
	mock_etherclient "github.com/forta-network/core-go/etherclient/mocks"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/require"
)

const testEventSignature = "SomeEvent(bytes32,address,uint256,uint256,uint256,uint256,uint256,string)"

var testEventTopic = crypto.Keccak256Hash([]byte(testEventSignature)).Hex()

func testLogs(indexes ...int) []types.Log {
	var result []types.Log
	for _, index := range indexes {
		result = append(result, types.Log{
			TxHash:  common.HexToHash(fmt.Sprintf("%x", index)),
			TxIndex: uint(index),
		})
	}
	return result
}

func TestLogFeed_ForEachLog(t *testing.T) {
	r := require.New(t)

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	client := mock_etherclient.NewMockEtherClient(ctrl)
	addr := "0x38C1e080BeEb26eeA91932178E62987598230271"
	logs := testLogs(0, 1, 2)

	client.EXPECT().GetBlockByNumber(gomock.Any(), nil).Return(&etherclient.Block{Number: "0x0"}, nil).Times(1)
	client.EXPECT().FilterLogs(gomock.Any(), gomock.Any()).Return([]types.Log{logs[0]}, nil).Times(1)

	client.EXPECT().GetBlockByNumber(gomock.Any(), big.NewInt(1)).Return(&etherclient.Block{Number: "0x0"}, nil).Times(1)
	client.EXPECT().FilterLogs(gomock.Any(), gomock.Any()).Return([]types.Log{logs[1]}, nil).Times(1)

	client.EXPECT().GetBlockByNumber(gomock.Any(), big.NewInt(2)).Return(&etherclient.Block{Number: "0x0"}, nil).Times(1)
	client.EXPECT().FilterLogs(gomock.Any(), gomock.Any()).Return([]types.Log{logs[2]}, nil).Times(1)

	lf, err := NewLogFeed(ctx, client, LogFeedConfig{
		Addresses: []string{addr},
		Topics:    [][]string{{testEventTopic}},
	})
	r.NoError(err)

	var foundLogs []types.Log
	err = lf.ForEachLog(func(blk *etherclient.Block, logEntry types.Log) error {
		foundLogs = append(foundLogs, logEntry)
		// return early
		if len(foundLogs) == 3 {
			return context.Canceled
		}
		return nil
	}, func(blk *etherclient.Block) error {
		return nil
	})
	// ensure expected error is the one returned
	r.ErrorIs(err, context.Canceled)

	r.Equal(len(logs), len(foundLogs), "should find all logs")
	for idx, fl := range foundLogs {
		r.Equal(logs[idx].TxIndex, fl.TxIndex)
		r.Equal(logs[idx].TxHash.Hex(), fl.TxHash.Hex())
	}
}

func TestLogFeed_ForEachLogPolling(t *testing.T) {
	r := require.New(t)

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	client := mock_etherclient.NewMockEtherClient(ctrl)
	addr := "0x38C1e080BeEb26eeA91932178E62987598230271"
	logs := testLogs(0, 1, 2)

	client.EXPECT().GetBlockByNumber(gomock.Any(), nil).Return(&etherclient.Block{Number: "0xf"}, nil).Times(1)

	client.EXPECT().GetBlockByNumber(gomock.Any(), big.NewInt(13)).Return(&etherclient.Block{Number: "0x0"}, nil).Times(1)
	client.EXPECT().FilterLogs(gomock.Any(), gomock.Any()).Return([]types.Log{logs[0]}, nil).Times(1)

	client.EXPECT().GetBlockByNumber(gomock.Any(), big.NewInt(14)).Return(&etherclient.Block{Number: "0x0"}, nil).Times(1)
	client.EXPECT().FilterLogs(gomock.Any(), gomock.Any()).Return([]types.Log{logs[1]}, nil).Times(1)

	client.EXPECT().GetBlockByNumber(gomock.Any(), big.NewInt(15)).Return(&etherclient.Block{Number: "0x0"}, nil).Times(1)
	client.EXPECT().FilterLogs(gomock.Any(), gomock.Any()).Return([]types.Log{logs[2]}, nil).Times(1)

	lf, err := NewLogFeed(ctx, client, LogFeedConfig{
		Addresses:  []string{addr},
		Topics:     [][]string{{testEventTopic}},
		StartBlock: big.NewInt(13),
	})
	r.NoError(err)

	var found []types.Log
	err = lf.ForEachLogPolling(
		1*time.Millisecond, // fast tick for test
		func(_ *etherclient.Block, lg types.Log) error {
			found = append(found, lg)
			if len(found) == 3 {
				return context.Canceled // stop after all logs
			}
			return nil
		},
		func(_ *etherclient.Block) error { return nil },
	)
	// ensure expected error is the one returned
	r.ErrorIs(err, context.Canceled)

	r.Equal(len(logs), len(found), "should find all logs")
	for idx, fl := range found {
		r.Equal(logs[idx].TxIndex, fl.TxIndex)
		r.Equal(logs[idx].TxHash.Hex(), fl.TxHash.Hex())
	}
}
