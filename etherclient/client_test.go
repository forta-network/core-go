package etherclient

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// newTestClientWithReply returns a client of a JSON-RPC server which answers every
// request with the given reply, in which $ID stands for the id of the request.
func newTestClientWithReply(t *testing.T, reply string) *etherClient {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID json.RawMessage `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(strings.ReplaceAll(reply, "$ID", string(req.ID))))
	}))
	t.Cleanup(srv.Close)

	client, err := DialContext(context.Background(), srv.URL)
	require.NoError(t, err)
	t.Cleanup(client.Close)
	client.SetRetryInterval(10 * time.Millisecond)
	return client
}

// newTestClient returns a client of a JSON-RPC server which answers every request
// with the given result.
func newTestClient(t *testing.T, result string) *etherClient {
	return newTestClientWithReply(t, `{"jsonrpc":"2.0","id":$ID,"result":`+result+`}`)
}

// TestSuccessReplyWithNullError guards a behaviour of go-ethereum that a bump changes.
//
// A server can send "error":null in a successful reply, as JSON-RPC 1.0 requires.
// go-ethereum returns the result of such a reply up to 1.17.3. From 1.17.4 it
// returns "json-rpc error 0" instead, and this client retries a failed call, for
// up to 12 hours in GetBlockTransactions.
//
// This test is expected to fail on go-ethereum 1.17.4 and later. Decide how to
// handle such servers before changing the fixture.
func TestSuccessReplyWithNullError(t *testing.T) {
	r := require.New(t)

	// A failed call is retried until the context is done, which hides its error.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var callErr error
	recordCallErr := func(_, _ string, err error) { callErr = err }

	client := newTestClientWithReply(t, `{"jsonrpc":"2.0","id":$ID,"result":"0x10","error":null}`)
	client.SetMetricsHandler(recordCallErr)
	blockNumber, err := client.BlockNumber(ctx)
	r.NoError(err, "the call failed with: %v", callErr)
	r.Equal(uint64(16), blockNumber)

	client = newTestClientWithReply(t, `{"jsonrpc":"2.0","id":$ID,"result":{"hash":"0x01","transactions":[{"hash":"0x02","type":"0x2"}]},"error":null}`)
	client.SetMetricsHandler(recordCallErr)
	txs, err := client.GetBlockTransactions(ctx, big.NewInt(16))
	r.NoError(err, "the call failed with: %v", callErr)
	r.Len(txs, 1)
	r.Equal("0x02", txs[0].Hash)
}
