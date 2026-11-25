// Copyright 2026 github.com/solana-foundation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpc

import (
	"context"
	"errors"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	stdjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Valid keys that replace the strings "vote" and "node" used by the Rust tests.
const (
	vatTestVote = "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	vatTestNode = "67omRD8GkXTi8daWceprtQABkDeEm3SZYHUCrthQGB9D"
)

// Port of test_rank_map_stakes_are_positive_json_numbers in
// rpc-client-types/src/response.rs, Agave v4.5.0-beta.0.
func TestVatActiveSetRankMap_StakesArePositiveJSONNumbers(t *testing.T) {
	mapJSON := func(totalStake, stake, rank string) []byte {
		return []byte(`{"epoch":1,"totalStake":` + totalStake + `,"validators":[` +
			`{"rank":` + rank + `,"votePubkey":"` + vatTestVote + `","nodePubkey":"` + vatTestNode + `","blsPubkeyCompressed":"bls","stake":` + stake + `}]}`)
	}

	in := mapJSON("100", "100", "0")
	var got VatActiveSetRankMap
	require.NoError(t, stdjson.Unmarshal(in, &got))
	out, err := stdjson.Marshal(got)
	require.NoError(t, err)
	assert.JSONEq(t, string(in), string(out))

	err = stdjson.Unmarshal(mapJSON("0", "100", "0"), &VatActiveSetRankMap{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "totalStake")

	err = stdjson.Unmarshal(mapJSON("100", "0", "0"), &VatActiveSetRankMap{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validators[0].stake")

	// Rank is a Rust u16.
	var maxRank VatActiveSetRankMap
	require.NoError(t, stdjson.Unmarshal(mapJSON("100", "100", "65535"), &maxRank))
	require.Len(t, maxRank.Validators, 1)
	assert.Equal(t, uint16(65535), maxRank.Validators[0].Rank)
	require.Error(t, stdjson.Unmarshal(mapJSON("100", "100", "65536"), &VatActiveSetRankMap{}))
}

// Port of test_get_vat_active_set_rank_map in rpc-client/src/rpc_client.rs,
// Agave v4.5.0-beta.0.
func TestClient_GetVatActiveSetRankMap(t *testing.T) {
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(`{"context":{"slot":1},"value":null}`)))
	defer closer()
	client := New(server.URL)

	out, err := client.GetVatActiveSetRankMap(context.Background(), 0, nil)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, uint64(1), out.Context.Slot)
	assert.Nil(t, out.Value)
	assert.Equal(t, []any{float64(0)}, server.RequestBody(t)["params"])

	out, err = client.GetVatActiveSetRankMap(context.Background(), 0, &GetVatActiveSetRankMapOpts{})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, []any{float64(0)}, server.RequestBody(t)["params"])

	minSlot := uint64(1)
	out, err = client.GetVatActiveSetRankMap(context.Background(), 0, &GetVatActiveSetRankMapOpts{MinContextSlot: &minSlot})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, uint64(1), out.Context.Slot)
	assert.Nil(t, out.Value)
	assert.Equal(t,
		[]any{float64(0), map[string]any{"minContextSlot": float64(1)}},
		server.RequestBody(t)["params"],
	)
}

// Port of test_get_vat_active_set_rank_map_http in rpc-client/src/rpc_client.rs,
// Agave v4.5.0-beta.0.
func TestClient_GetVatActiveSetRankMap_HTTP(t *testing.T) {
	// A processed default commitment must not be added to this finalized-only request.
	newClient := func(url string) *Client { return NewWithCommitment(url, CommitmentProcessed) }
	vote := solana.MustPublicKeyFromBase58(vatTestVote)
	node := solana.MustPublicKeyFromBase58(vatTestNode)

	t.Run("identity and min context slot", func(t *testing.T) {
		server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(
			`{"context":{"slot":42},"value":{"epoch":3,"totalStake":100,"validators":[`+
				`{"rank":7,"votePubkey":"`+vatTestVote+`","nodePubkey":"`+vatTestNode+`","blsPubkeyCompressed":"bls","stake":10}]}}`)))
		defer closer()

		minSlot := uint64(42)
		out, err := newClient(server.URL).GetVatActiveSetRankMap(context.Background(), 100,
			&GetVatActiveSetRankMapOpts{Identity: &node, MinContextSlot: &minSlot})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, uint64(42), out.Context.Slot)
		assert.Equal(t, &VatActiveSetRankMap{
			Epoch:      3,
			TotalStake: 100,
			Validators: []VatActiveSetRankMapEntry{{
				Rank:                7,
				VotePubkey:          vote,
				NodePubkey:          node,
				BLSPubkeyCompressed: "bls",
				Stake:               10,
			}},
		}, out.Value)

		reqBody := server.RequestBody(t)
		reqBody["id"] = any(nil)
		// Exact equality also proves that no commitment is sent.
		assert.Equal(t,
			map[string]any{
				"id":      any(nil),
				"jsonrpc": "2.0",
				"method":  "getVatActiveSetRankMap",
				"params": []any{
					float64(100),
					map[string]any{"identity": vatTestNode, "minContextSlot": float64(42)},
				},
			},
			reqBody,
		)
	})

	t.Run("nil opts and null value", func(t *testing.T) {
		server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(`{"context":{"slot":42},"value":null}`)))
		defer closer()

		// The Rust client sends {"identity": null, "minContextSlot": null} here,
		// and the server takes a missing config the same way.
		out, err := newClient(server.URL).GetVatActiveSetRankMap(context.Background(), 200, nil)
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, uint64(42), out.Context.Slot)
		assert.Nil(t, out.Value)
		assert.Equal(t, []any{float64(200)}, server.RequestBody(t)["params"])
	})

	t.Run("rpc error", func(t *testing.T) {
		server, closer := mockJSONRPC(t, stdjson.RawMessage(
			`{"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"unexpected slot"}}`))
		defer closer()

		out, err := newClient(server.URL).GetVatActiveSetRankMap(context.Background(), 300, nil)
		var rpcErr *jsonrpc.RPCError
		require.True(t, errors.As(err, &rpcErr))
		assert.Equal(t, -32602, rpcErr.Code)
		assert.Equal(t, "unexpected slot", rpcErr.Message)
		assert.Nil(t, out)
	})
}

// Response shape from test_rpc_rank_map_cache_and_filter in rpc/src/rpc.rs,
// Agave v4.5.0-beta.0. The identity filter keeps the total stake of the full map
// and may leave no validators.
func TestClient_GetVatActiveSetRankMap_UnknownIdentity(t *testing.T) {
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(
		`{"context":{"slot":42},"value":{"epoch":3,"totalStake":100,"validators":[]}}`)))
	defer closer()

	node := solana.MustPublicKeyFromBase58(vatTestNode)
	out, err := New(server.URL).GetVatActiveSetRankMap(context.Background(), 0,
		&GetVatActiveSetRankMapOpts{Identity: &node})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.NotNil(t, out.Value)
	assert.Equal(t, uint64(3), out.Value.Epoch)
	assert.Equal(t, uint64(100), out.Value.TotalStake)
	assert.Empty(t, out.Value.Validators)
	assert.Equal(t,
		[]any{float64(0), map[string]any{"identity": vatTestNode}},
		server.RequestBody(t)["params"],
	)
}

func TestClient_GetVatActiveSetRankMap_NullResult(t *testing.T) {
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(`null`)))
	defer closer()

	out, err := New(server.URL).GetVatActiveSetRankMap(context.Background(), 0, nil)
	require.Error(t, err)
	assert.Nil(t, out)
}
