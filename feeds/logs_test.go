package feeds

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	mock_etherclient "github.com/forta-network/core-go/etherclient/mocks"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/require"
)

const testBlockHeader = `{"parentHash":"0x6d4482241235590f6740041d4922f2a5f71fec845cd1c097092a9daede043f1d","sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","miner":"0x388c818ca8b9251b393131c08a736a67ccb19297","stateRoot":"0xdf7beb53c01d9d8f24431419626f25d8dc2f1dda6b756dfc9e33d2c500b4a5c3","transactionsRoot":"0x1810ca93b0ae67a03a5c690b3f39de092e723ea7e15f828f4b07c12fdcb80c08","receiptsRoot":"0xa4687f5b2d340bd5fb6ba44971f55c387efd0f70e3c0f33af12684a76a013e5f","logsBloom":"0x092089449c0000b92d3000c581c891e41498802c04c2e00b004141a8a6130a47012000a080b10080e24d1f50506e610d2a01826098323ed8900500401172216044060e4c34803b097880040ec8caa0300044969ba148644a18348855804108048560708a0240802a5841918113000a9882002830403824004228855702888810020c0117084041350849c08955232250d10430430dc90689a028d06600b102a6e300203022b07c011101d0a008040ee1a41506c0201152080d011806000020500c9c80a221609810223e4491108454d4054a0006002a01b80130c10b5002a2e00c90e09911010012239cc50291818c00106468474d054044e20828cc41011403","difficulty":"0x0","number":"0x15062c5","gasLimit":"0x2243e17","gasUsed":"0x6018ee","timestamp":"0x67d42d5f","extraData":"0x","mixHash":"0x29beb9649ff00a98217a78c32eaff8b9527ded07222d3fffb33f672fc2e5374e","nonce":"0x0000000000000000","baseFeePerGas":"0x244e4035","withdrawalsRoot":"0xe11546bd54459b7b1db2a978a67f1c30dc8640464043bad33828d51516091881","blobGasUsed":"0xa0000","excessBlobGas":"0x4560000","parentBeaconBlockRoot":"0x340bd006d0dec90aa0a562d491fc6fed47283f746d4a1bb57a14f9ea3062599a","requestsRoot":null,"hash":"0x60afd7ae876ef01c47e8dbfd289d1ce63462f406d7bf392ee99353a048814d48"}`

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

func testBlock(r *require.Assertions) *types.Block {
	var header *types.Header
	err := json.Unmarshal([]byte(testBlockHeader), &header)
	r.NoError(err)
	return types.NewBlockWithHeader(header)
}

func TestLogFeed_ForEachLog(t *testing.T) {
	r := require.New(t)

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	client := mock_etherclient.NewMockEtherClient(ctrl)
	addr := "0x38C1e080BeEb26eeA91932178E62987598230271"
	logs := testLogs(0, 1, 2)

	client.EXPECT().BlockByNumber(gomock.Any(), nil).Return(testBlock(r), nil).Times(1)
	client.EXPECT().FilterLogs(gomock.Any(), gomock.Any()).Return([]types.Log{logs[0]}, nil).Times(1)

	client.EXPECT().BlockByNumber(gomock.Any(), big.NewInt(22045382)).Return(testBlock(r), nil).Times(1)
	client.EXPECT().FilterLogs(gomock.Any(), gomock.Any()).Return([]types.Log{logs[1]}, nil).Times(1)

	client.EXPECT().BlockByNumber(gomock.Any(), big.NewInt(22045383)).Return(testBlock(r), nil).Times(1)
	client.EXPECT().FilterLogs(gomock.Any(), gomock.Any()).Return([]types.Log{logs[2]}, nil).Times(1)

	lf, err := NewLogFeed(ctx, client, LogFeedConfig{
		Addresses: []string{addr},
		Topics:    [][]string{{testEventTopic}},
	})
	r.NoError(err)

	var foundLogs []types.Log
	err = lf.ForEachLog(func(blk *types.Block, logEntry types.Log) error {
		foundLogs = append(foundLogs, logEntry)
		// return early
		if len(foundLogs) == 3 {
			return context.Canceled
		}
		return nil
	}, func(blk *types.Block) error {
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
