package etherclient

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

const testAuthorizationFormat = `{"chainId":"0x1","address":"0x00000000000000000000000000000000000000f2","nonce":"0x5","yParity":"%s","r":"0xa","s":"0xb"}`

func testAuthorization(yParity uint8) gethtypes.SetCodeAuthorization {
	return gethtypes.SetCodeAuthorization{
		ChainID: *uint256.NewInt(1),
		Address: common.HexToAddress("0xf2"),
		Nonce:   5,
		V:       yParity,
		R:       *uint256.NewInt(10),
		S:       *uint256.NewInt(11),
	}
}

func TestGetBlockTransactions_AuthorizationYParity(t *testing.T) {
	r := require.New(t)

	// A block can contain an authorization with any yParity up to 255.
	block := fmt.Sprintf(`{"hash":"0x01","transactions":[
		{"from":"0xf1","to":"0xf2","nonce":"0x7","value":"0x1","input":"0xdeadbeef","hash":"0x02","type":"0x4","authorizationList":[%s,%s,%s]},
		{"from":"0xf1","to":null,"nonce":"0x8","value":"0x0","input":"0x","hash":"0x03","type":"0x2"},
		null,
		{"hash":"0x04","type":"0x4","authorizationList":[]}
	]}`,
		fmt.Sprintf(testAuthorizationFormat, "0x1"),
		fmt.Sprintf(testAuthorizationFormat, "0x2"),
		fmt.Sprintf(testAuthorizationFormat, "0xff"),
	)
	client := newTestClient(t, block)

	// A block that fails to decode is retried until the context is done.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	txs, err := client.GetBlockTransactions(ctx, big.NewInt(16))
	r.NoError(err)
	r.Len(txs, 4)

	nonce := txs[0].Nonce
	r.NotNil(nonce)
	r.Equal(uint64(7), uint64(*nonce))
	r.Equal("0x02", txs[0].Hash)
	r.Equal("0xf2", txs[0].To)
	r.Equal("0xdeadbeef", txs[0].Input)
	r.Equal(SetCodeTxType, txs[0].Type)
	r.Equal([]gethtypes.SetCodeAuthorization{
		testAuthorization(1),
		testAuthorization(2),
		testAuthorization(255),
	}, txs[0].AuthorizationList)

	r.Equal("0x03", txs[1].Hash)
	r.Equal(DynamicFeeTxType, txs[1].Type)
	r.Nil(txs[1].AuthorizationList)

	// A null transaction stays nil.
	r.Nil(txs[2])

	// An empty authorization list stays empty and is not nil.
	r.Equal("0x04", txs[3].Hash)
	r.NotNil(txs[3].AuthorizationList)
	r.Empty(txs[3].AuthorizationList)
}

func TestGetBlockTransactions_NullTransactions(t *testing.T) {
	r := require.New(t)

	client := newTestClient(t, `{"hash":"0x01","transactions":null}`)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	txs, err := client.GetBlockTransactions(ctx, big.NewInt(16))
	r.NoError(err)
	r.Nil(txs)
}

func TestAuthorizationJSON(t *testing.T) {
	// A valid yParity is decoded the same way as go-ethereum does.
	for _, yParity := range []string{"0x0", "0x1"} {
		input := []byte(fmt.Sprintf(testAuthorizationFormat, yParity))

		var expected gethtypes.SetCodeAuthorization
		require.NoError(t, json.Unmarshal(input, &expected))

		var auth authorizationJSON
		require.NoError(t, json.Unmarshal(input, &auth))
		require.Equal(t, expected, gethtypes.SetCodeAuthorization(auth))
	}

	// Every field is required.
	for _, field := range []string{"chainId", "address", "nonce", "yParity", "r", "s"} {
		var fields map[string]string
		require.NoError(t, json.Unmarshal([]byte(fmt.Sprintf(testAuthorizationFormat, "0x1")), &fields))
		delete(fields, field)
		input, err := json.Marshal(fields)
		require.NoError(t, err)

		var auth authorizationJSON
		require.ErrorContains(t, json.Unmarshal(input, &auth), fmt.Sprintf("missing required field '%s'", field))
	}
}
