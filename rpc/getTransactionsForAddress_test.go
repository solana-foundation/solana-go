// Copyright 2021 github.com/gagliardetto
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
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

func TestClient_GetTransactionsForAddress(t *testing.T) {
	responseBody := `{"data":[{"slot":83994671,"transactionIndex":3,"blockTime":1625231961,"signature":"4Yig3yd33o2hyZV2qZBJkScDArwVmzurkxhBfKdqJeujTrdKHwrR3U8KR6LrhN5eWNTyugS5rkkYagVXCNnk7pks","err":null,"memo":null,"confirmationStatus":"finalized"},{"slot":83994656,"transactionIndex":1,"blockTime":1625231952,"signature":"3oQ7qqpJs5CtH1Xnnn8Ru5MtxkR3SZgshqzXwokuxFRArLihKdvCb9km6gbSiiUaNSHE7zVJqUVUZGfYuEaqWZPV","err":null,"memo":null,"confirmationStatus":"finalized"}],"paginationToken":"83994656:1"}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	limit := 10
	minContextSlot := uint64(123456)
	slotFloor := uint64(83000000)

	opts := GetTransactionsForAddressOpts{
		TransactionDetails: TransactionDetailsSignatures,
		SortOrder:          TransactionsForAddressSortDesc,
		Limit:              &limit,
		PaginationToken:    "83999999:0",
		Commitment:         CommitmentFinalized,
		MinContextSlot:     &minContextSlot,
		Filters: &TransactionsForAddressFilters{
			Status: TransactionStatusSucceeded,
			Slot:   &RangeFilterUint64{Gte: &slotFloor},
		},
	}
	out, err := client.GetTransactionsForAddressWithOpts(
		context.Background(),
		pubKey,
		&opts,
	)
	require.NoError(t, err)

	// The id is random, so we can't assert it; check it is set, then drop it.
	reqBody := server.RequestBody(t)
	assert.NotNil(t, reqBody["id"])
	reqBody["id"] = any(nil)

	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTransactionsForAddress",
			"params": []any{
				pubkeyString,
				map[string]any{
					"transactionDetails": string(TransactionDetailsSignatures),
					"sortOrder":          string(TransactionsForAddressSortDesc),
					"limit":              float64(limit),
					"paginationToken":    "83999999:0",
					"commitment":         string(CommitmentFinalized),
					"minContextSlot":     float64(minContextSlot),
					"filters": map[string]any{
						"status": "succeeded",
						"slot": map[string]any{
							"gte": float64(slotFloor),
						},
					},
				},
			},
		},
		reqBody,
	)

	require.NotNil(t, out)
	require.Len(t, out.Data, 2)
	assert.Equal(t, uint64(83994671), out.Data[0].Slot)
	assert.Equal(t, uint64(3), out.Data[0].TransactionIndex)
	assert.Equal(t,
		"4Yig3yd33o2hyZV2qZBJkScDArwVmzurkxhBfKdqJeujTrdKHwrR3U8KR6LrhN5eWNTyugS5rkkYagVXCNnk7pks",
		out.Data[0].Signature.String(),
	)
	assert.Equal(t, ConfirmationStatusFinalized, out.Data[0].ConfirmationStatus)
	require.NotNil(t, out.PaginationToken)
	assert.Equal(t, "83994656:1", *out.PaginationToken)
}

func TestClient_GetTransactionsForAddress_NoOpts(t *testing.T) {
	responseBody := `{"data":[],"paginationToken":null}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubkeyString := "7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932"
	pubKey := solana.MustPublicKeyFromBase58(pubkeyString)

	out, err := client.GetTransactionsForAddress(context.Background(), pubKey)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)

	// With no opts, only the address is sent (no config object).
	assert.Equal(t,
		map[string]any{
			"id":      any(nil),
			"jsonrpc": "2.0",
			"method":  "getTransactionsForAddress",
			"params":  []any{pubkeyString},
		},
		reqBody,
	)

	require.NotNil(t, out)
	assert.Empty(t, out.Data)
	assert.Nil(t, out.PaginationToken)
}

func TestClient_GetTransactionsForAddress_FullDetail(t *testing.T) {
	responseBody := `{"data":[{"slot":83994671,"transactionIndex":2,"blockTime":1625231961,"transaction":["AQ==","base64"],"meta":{"err":null,"fee":5000,"preBalances":[1000],"postBalances":[995]}}],"paginationToken":null}`
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(responseBody)))
	defer closer()
	client := New(server.URL)

	pubKey := solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")

	out, err := client.GetTransactionsForAddressWithOpts(
		context.Background(),
		pubKey,
		&GetTransactionsForAddressOpts{
			TransactionDetails: TransactionDetailsFull,
			Encoding:           solana.EncodingBase64,
		},
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	reqBody["id"] = any(nil)
	params, ok := reqBody["params"].([]any)
	require.True(t, ok)
	cfg, ok := params[1].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, string(solana.EncodingBase64), cfg["encoding"])
	assert.Equal(t, string(TransactionDetailsFull), cfg["transactionDetails"])

	require.NotNil(t, out)
	require.Len(t, out.Data, 1)
	require.NotNil(t, out.Data[0].Transaction)
	assert.Equal(t, uint64(2), out.Data[0].TransactionIndex)
	require.NotNil(t, out.Data[0].Meta)
	assert.Equal(t, uint64(5000), out.Data[0].Meta.Fee)
}

func TestClient_GetTransactionsForAddress_Base64ZstdEncoding(t *testing.T) {
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(`{"data":[],"paginationToken":null}`)))
	defer closer()
	client := New(server.URL)

	pubKey := solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")

	// base64+zstd is accepted everywhere else in this package (getBlock,
	// getTransaction, blockSubscribe); it must be accepted here too rather
	// than rejected client-side.
	_, err := client.GetTransactionsForAddressWithOpts(
		context.Background(),
		pubKey,
		&GetTransactionsForAddressOpts{
			TransactionDetails: TransactionDetailsFull,
			Encoding:           solana.EncodingBase64Zstd,
		},
	)
	require.NoError(t, err)

	reqBody := server.RequestBody(t)
	params, ok := reqBody["params"].([]any)
	require.True(t, ok)
	cfg, ok := params[1].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, string(solana.EncodingBase64Zstd), cfg["encoding"])
}

func TestClient_GetTransactionsForAddress_InvalidEncoding(t *testing.T) {
	server, closer := mockJSONRPC(t, stdjson.RawMessage(wrapIntoRPC(`{"data":[],"paginationToken":null}`)))
	defer closer()
	client := New(server.URL)

	pubKey := solana.MustPublicKeyFromBase58("7xLk17EQQ5KLDLDe44wCmupJKJjTGd8hs3eSVVhCx932")

	_, err := client.GetTransactionsForAddressWithOpts(
		context.Background(),
		pubKey,
		&GetTransactionsForAddressOpts{
			Encoding: solana.EncodingType("not-a-real-encoding"),
		},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "encoding is not supported")
}
