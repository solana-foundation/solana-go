package ws

import (
	"strconv"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/require"
)

// readSubscribeParams reads one subscribe request from the mock server and
// returns the params map of the second element (the options object), responding
// to the client so the subscribe call can return.
func readSubscribeParams(t *testing.T, m *mockWSServer, wsSubID uint64) map[string]any {
	t.Helper()
	select {
	case msg := <-m.incoming:
		var req struct {
			ID     uint64 `json:"id"`
			Method string `json:"method"`
			Params []any  `json:"params"`
		}
		require.NoError(t, json.Unmarshal(msg, &req))
		resp := `{"jsonrpc":"2.0","result":` + strconv.FormatUint(wsSubID, 10) + `,"id":` + strconv.FormatUint(req.ID, 10) + `}`
		m.send(t, resp)
		// params[0] is the filter ("all"), params[1] is the opts object
		require.GreaterOrEqual(t, len(req.Params), 2, "expected opts object in params[1]: %s", string(msg))
		opts, ok := req.Params[1].(map[string]any)
		require.True(t, ok, "expected map opts in params[1], got %T", req.Params[1])
		return opts
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for subscribe request")
		return nil
	}
}

// TestBlockSubscribeUsesShowRewardsField pins that BlockSubscribe sends the
// Solana-spec field "showRewards" (not "rewards") in its params object.
//
// See https://solana.com/docs/rpc/websocket/blocksubscribe for the spec, and
// issue #208 for the bug report.
func TestBlockSubscribeUsesShowRewardsField(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	rewards := false
	opts := &BlockSubscribeOpts{
		Commitment: rpc.CommitmentConfirmed,
		Rewards:    &rewards,
	}

	type subResult struct {
		err error
	}
	ch := make(chan subResult, 1)
	go func() {
		_, err := c.BlockSubscribe(NewBlockSubscribeFilterAll(), opts)
		ch <- subResult{err}
	}()

	params := readSubscribeParams(t, m, 1)
	require.NoError(t, (<-ch).err)

	_, hasOldKey := params["rewards"]
	require.False(t, hasOldKey, "params must not include legacy 'rewards' key: %v", params)

	v, ok := params["showRewards"]
	require.True(t, ok, "params must include 'showRewards' key: %v", params)
	require.Equal(t, false, v)
}

// TestParsedBlockSubscribeUsesShowRewardsField pins the same fix for the
// parsed-block subscription path.
func TestParsedBlockSubscribeUsesShowRewardsField(t *testing.T) {
	m := newMockWSServer(t)
	defer m.stop()
	c := connectClient(t, m)
	defer c.Close()

	rewards := true
	opts := &BlockSubscribeOpts{
		Commitment: rpc.CommitmentConfirmed,
		Rewards:    &rewards,
	}

	type subResult struct {
		err error
	}
	ch := make(chan subResult, 1)
	go func() {
		_, err := c.ParsedBlockSubscribe(NewBlockSubscribeFilterAll(), opts)
		ch <- subResult{err}
	}()

	params := readSubscribeParams(t, m, 2)
	require.NoError(t, (<-ch).err)

	_, hasOldKey := params["rewards"]
	require.False(t, hasOldKey, "params must not include legacy 'rewards' key: %v", params)

	v, ok := params["showRewards"]
	require.True(t, ok, "params must include 'showRewards' key: %v", params)
	require.Equal(t, true, v)
}
