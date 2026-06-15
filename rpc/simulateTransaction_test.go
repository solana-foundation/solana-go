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
	"testing"

	stdjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gagliardetto/solana-go"
)

// TestClient_SimulateRawTransactionWithOpts_AccountsEncodingDefault pins
// that when the caller passes a SimulateTransactionAccountsOpts without
// setting Encoding explicitly, the SDK substitutes base64 on the wire.
//
// Regression test: previously the empty-string Encoding was forwarded as
// {"encoding":""} and the Solana validator rejected the call with
// `Invalid params: missing field encoding` instead of running the
// simulation.
func TestClient_SimulateRawTransactionWithOpts_AccountsEncodingDefault(t *testing.T) {
	responseBody := `{"context":{"slot":1},"value":{"err":null,"logs":[],"accounts":null}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	address := solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")
	_, err := client.SimulateRawTransactionWithOpts(
		context.Background(),
		[]byte("rawtx"),
		&SimulateTransactionOpts{
			Accounts: &SimulateTransactionAccountsOpts{
				// Encoding intentionally left zero-value.
				Addresses: []solana.PublicKey{address},
			},
		},
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	params, ok := reqBody["params"].([]any)
	require.True(t, ok, "params must be a JSON array, got %T", reqBody["params"])
	require.Len(t, params, 2)
	cfg, ok := params[1].(map[string]any)
	require.True(t, ok, "config object expected, got %T", params[1])
	accounts, ok := cfg["accounts"].(map[string]any)
	require.True(t, ok, "accounts config expected, got %T", cfg["accounts"])
	assert.Equal(t, string(solana.EncodingBase64), accounts["encoding"], "empty Encoding must default to base64 on the wire")
}

// TestClient_SimulateRawTransactionWithOpts_AccountsEncodingExplicit pins
// that an explicit Encoding (e.g. jsonParsed) is preserved unchanged.
func TestClient_SimulateRawTransactionWithOpts_AccountsEncodingExplicit(t *testing.T) {
	responseBody := `{"context":{"slot":1},"value":{"err":null,"logs":[],"accounts":null}}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	address := solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")
	_, err := client.SimulateRawTransactionWithOpts(
		context.Background(),
		[]byte("rawtx"),
		&SimulateTransactionOpts{
			Accounts: &SimulateTransactionAccountsOpts{
				Encoding:  solana.EncodingJSONParsed,
				Addresses: []solana.PublicKey{address},
			},
		},
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	params, ok := reqBody["params"].([]any)
	require.True(t, ok, "params must be a JSON array, got %T", reqBody["params"])
	require.Len(t, params, 2)
	cfg, ok := params[1].(map[string]any)
	require.True(t, ok, "config object expected, got %T", params[1])
	accounts, ok := cfg["accounts"].(map[string]any)
	require.True(t, ok, "accounts config expected, got %T", cfg["accounts"])
	assert.Equal(t, string(solana.EncodingJSONParsed), accounts["encoding"], "explicit Encoding must be forwarded unchanged")
}
